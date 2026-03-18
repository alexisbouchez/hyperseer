package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/exit"
	"github.com/alexisbouchez/hyperseer/internal/query"
)

func ms(ns int64) string {
	return fmt.Sprintf("%dms", ns/1_000_000)
}

func statusColor(s string) string {
	switch s {
	case "STATUS_CODE_ERROR":
		return "\033[31m✗\033[0m"
	case "STATUS_CODE_OK":
		return "\033[32m✓\033[0m"
	default:
		return "\033[2m•\033[0m"
	}
}

func timeRange(cmd *cli.Command) (from, to time.Time) {
	to = time.Now()
	from = to.Add(-time.Hour)
	if t := cmd.String("from"); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			from = parsed
		}
	}
	if t := cmd.String("to"); t != "" {
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			to = parsed
		}
	}
	if d := cmd.Duration("last"); d != 0 {
		from = to.Add(-d)
	}
	return
}

var timeFlags = []cli.Flag{
	&cli.StringFlag{Name: "from", Usage: "Start time (RFC3339), defaults to 1h ago"},
	&cli.StringFlag{Name: "to", Usage: "End time (RFC3339), defaults to now"},
	&cli.DurationFlag{Name: "last", Usage: "Shorthand: last N duration ending at --to (e.g. --last 30m)"},
}

func serverURL(cmd *cli.Command) string {
	return cmd.Root().String("server")
}

func main() {
	cfg := config.New()

	app := &cli.Command{
		Name:  "seer",
		Usage: "Hyperseer CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Sources: cli.EnvVars("HYPERSEER_QUERY_URL"),
				Value:   cfg.QueryURL,
				Usage:   "Query API URL",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "logs",
				Usage: "Query logs",
				Flags: append([]cli.Flag{
					&cli.StringFlag{Name: "service", Aliases: []string{"s"}, Usage: "Filter by service name"},
					&cli.StringFlag{Name: "severity", Usage: "Filter by severity"},
					&cli.IntFlag{Name: "limit", Aliases: []string{"n"}, Value: 50, Usage: "Max results"},
				}, timeFlags...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					from, to := timeRange(cmd)
					logs, err := fetchLogs(serverURL(cmd), query.LogsParams{
						Service:  cmd.String("service"),
						Severity: cmd.String("severity"),
						Since:    from,
						Until:    to,
						Limit:    cmd.Int("limit"),
					})
					if err != nil {
						return err
					}
					for _, l := range logs {
						fmt.Fprintf(os.Stdout, "%s  %s  %s  %s\n",
							l.Time.Format("2006-01-02 15:04:05"),
							severityColor(l.Severity),
							"\033[2m"+l.ServiceName+"\033[0m",
							l.Body,
						)
					}
					return nil
				},
			},
			{
				Name:      "traces",
				Usage:     "Query traces. Pass a trace ID to inspect a specific trace.",
				ArgsUsage: "[trace-id]",
				Flags: append([]cli.Flag{
					&cli.StringFlag{Name: "service", Aliases: []string{"s"}, Usage: "Filter by service name"},
					&cli.IntFlag{Name: "limit", Aliases: []string{"n"}, Value: 50, Usage: "Max results"},
				}, timeFlags...),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if traceID := cmd.Args().First(); traceID != "" {
						spans, err := fetchTraceSpans(serverURL(cmd), traceID)
						if err != nil {
							return err
						}
						printWaterfall(spans)
						return nil
					}

					from, to := timeRange(cmd)
					spans, err := fetchTraces(serverURL(cmd), query.TracesParams{
						Service: cmd.String("service"),
						Since:   from,
						Until:   to,
						Limit:   cmd.Int("limit"),
					})
					if err != nil {
						return err
					}
					for _, s := range spans {
						fmt.Fprintf(os.Stdout, "%s  %s  %-6s  \033[2m%s\033[0m  %s\n",
							s.Time.Format("2006-01-02 15:04:05"),
							statusColor(s.StatusCode),
							ms(s.Duration),
							s.ServiceName,
							s.Name,
						)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		exit.WithError(err)
	}
}

func printWaterfall(spans []query.Span) {
	children := map[string][]query.Span{}
	for _, s := range spans {
		children[s.ParentID] = append(children[s.ParentID], s)
	}
	var print func(parentID, prefix string)
	print = func(parentID, prefix string) {
		nodes := children[parentID]
		for i, s := range nodes {
			connector := "├── "
			childPrefix := prefix + "│   "
			if i == len(nodes)-1 {
				connector = "└── "
				childPrefix = prefix + "    "
			}
			fmt.Fprintf(os.Stdout, "%s%s%s  %s  %s  \033[2m%s\033[0m\n",
				prefix, connector, s.Name,
				ms(s.Duration),
				statusColor(s.StatusCode),
				strings.TrimPrefix(s.Kind, "SPAN_KIND_"),
			)
			print(s.SpanID, childPrefix)
		}
	}
	if len(spans) > 0 {
		root := spans[0]
		fmt.Fprintf(os.Stdout, "\033[1m%s\033[0m  %s  \033[2m%s\033[0m\n",
			root.TraceID[:16]+"…",
			ms(root.Duration),
			root.ServiceName,
		)
	}
	print("", "")
}

func severityColor(s string) string {
	padded := fmt.Sprintf("%-5s", strings.ToUpper(s))
	switch strings.ToLower(s) {
	case "error", "fatal", "error2", "error3", "error4":
		return "\033[31m" + padded + "\033[0m"
	case "warn", "warning":
		return "\033[33m" + padded + "\033[0m"
	case "info":
		return "\033[32m" + padded + "\033[0m"
	default:
		return "\033[2m" + padded + "\033[0m"
	}
}
