package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

var installCommand = &cli.Command{
	Name:  "install",
	Usage: "Install the Hyperseer skill for Claude Code (and compatible AI agents)",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "project",
			Usage: "Also inject Hyperseer context into the current project's CLAUDE.md",
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "Overwrite existing skill file if present",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		server := serverURL(cmd)

		// 1. Write ~/.claude/commands/hyperseer.md
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		commandsDir := filepath.Join(home, ".claude", "commands")
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			return fmt.Errorf("could not create %s: %w", commandsDir, err)
		}

		skillPath := filepath.Join(commandsDir, "hyperseer.md")
		if _, err := os.Stat(skillPath); err == nil && !cmd.Bool("force") {
			fmt.Printf("  skill already exists: %s\n", skillPath)
			fmt.Printf("  use --force to overwrite\n")
		} else {
			if err := os.WriteFile(skillPath, []byte(skillContent(server)), 0644); err != nil {
				return fmt.Errorf("could not write skill file: %w", err)
			}
			fmt.Printf("  wrote %s\n", skillPath)
		}

		// 2. Inject into project CLAUDE.md if requested
		if cmd.Bool("project") {
			if err := injectCLAUDEmd(server); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not update CLAUDE.md: %v\n", err)
			}
		}

		fmt.Println()
		fmt.Println("Claude Code is ready. Use /hyperseer in any session, or just ask Claude to check logs/traces.")
		if !cmd.Bool("project") {
			fmt.Println("Tip: run with --project to inject Hyperseer context into this project's CLAUDE.md.")
		}
		return nil
	},
}

func skillContent(server string) string {
	return fmt.Sprintf(`---
description: Query Hyperseer for logs, traces, and digests from the connected OpenTelemetry backend. Use this whenever the user asks about errors, slow requests, logs, traces, latency, or production issues.
allowed-tools: Bash(seer *)
argument-hint: [logs|traces|digest] [flags]
---

# Hyperseer

Hyperseer is a ClickHouse-backed OpenTelemetry platform. The CLI is **seer**.

Server: %s

## Commands

### seer logs
Query application logs.

Flags:
  --server URL        Query API URL (use the server above)
  --service, -s       Filter by service name
  --severity          Filter by severity: TRACE, DEBUG, INFO, WARN, ERROR, FATAL
  --limit, -n         Max results (default 50)
  --last DURATION     Time window, e.g. --last 1h, --last 30m, --last 24h
  --from RFC3339      Start time
  --to RFC3339        End time

Examples:
  seer logs --server %s --last 1h
  seer logs --server %s --last 30m --severity ERROR
  seer logs --server %s --last 1h --service my-api
  seer logs --server %s --last 24h --severity ERROR --limit 100

### seer traces
Query distributed traces. Pass a trace ID to see the full waterfall.

Flags:
  --server URL        Query API URL
  --service, -s       Filter by service name
  --limit, -n         Max results (default 50)
  --last DURATION     Time window
  --from / --to       Explicit time bounds

Examples:
  seer traces --server %s --last 1h
  seer traces --server %s --last 30m --service my-api
  seer traces --server %s <trace-id>       # full waterfall for one trace

## How to use

When the user asks about:
- "errors", "500s", "exceptions" → run seer logs with --severity ERROR
- "slow", "latency", "performance" → run seer traces --last 1h, look at durations
- "what happened at <time>" → use --from / --to with the time they mention
- "logs for <service>" → use --service flag
- specific trace ID → pass it as argument to seer traces

Always include --server %s in every command.

After running a command, summarize the key findings: how many errors, which services are affected, any patterns in the messages or trace names.
`, server,
		server, server, server, server,
		server, server, server,
		server)
}

func injectCLAUDEmd(server string) error {
	const marker = "<!-- hyperseer -->"
	const block = `<!-- hyperseer -->
## Hyperseer

OpenTelemetry logs and traces are available via the **seer** CLI.
Run ` + "`" + `/hyperseer` + "`" + ` or ask Claude to check logs/traces directly.

Server: %s

Quick reference:
- ` + "`" + `seer logs --server %s --last 1h --severity ERROR` + "`" + `
- ` + "`" + `seer traces --server %s --last 1h` + "`" + `
<!-- /hyperseer -->
`

	content := fmt.Sprintf(block, server, server, server)

	data, err := os.ReadFile("CLAUDE.md")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	existing := string(data)

	// Already injected — update in place
	if strings.Contains(existing, marker) {
		start := strings.Index(existing, marker)
		end := strings.Index(existing, "<!-- /hyperseer -->")
		if end != -1 {
			end += len("<!-- /hyperseer -->")
			existing = existing[:start] + strings.TrimRight(content, "\n") + existing[end:]
		}
		if err := os.WriteFile("CLAUDE.md", []byte(existing), 0644); err != nil {
			return err
		}
		fmt.Println("  updated CLAUDE.md")
		return nil
	}

	// Append to end
	updated := strings.TrimRight(existing, "\n") + "\n\n" + content
	if err := os.WriteFile("CLAUDE.md", []byte(updated), 0644); err != nil {
		return err
	}
	fmt.Println("  updated CLAUDE.md")
	return nil
}
