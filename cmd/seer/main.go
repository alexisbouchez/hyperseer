package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/db"
	"github.com/alexisbouchez/hyperseer/internal/exit"
	"github.com/alexisbouchez/hyperseer/internal/query"
)

func main() {
	app := &cli.Command{
		Name:  "seer",
		Usage: "Hyperseer CLI",
		Commands: []*cli.Command{
			{
				Name:  "logs",
				Usage: "Query logs",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "service", Aliases: []string{"s"}, Usage: "Filter by service name"},
					&cli.StringFlag{Name: "severity", Usage: "Filter by severity"},
					&cli.DurationFlag{Name: "since", Value: time.Hour, Usage: "How far back to look"},
					&cli.IntFlag{Name: "limit", Aliases: []string{"n"}, Value: 50, Usage: "Max results"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg := config.New()
					conn, err := db.New(cfg.ClickHouse)
					if err != nil {
						return err
					}
					defer conn.Close()

					logs, err := query.Logs(ctx, conn, query.LogsParams{
						Service:  cmd.String("service"),
						Severity: cmd.String("severity"),
						Since:    time.Now().Add(-cmd.Duration("since")),
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
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		exit.WithError(err)
	}
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
