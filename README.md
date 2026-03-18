# Hyperseer

> Observability for AI agents

## About

Hyperseer is a self-hosted OpenTelemetry backend that stores logs and traces in ClickHouse and exposes them through a command-line interface (CLI).

## Self-hosting

```bash
go run github.com/alexisbouchez/hyperseer/cmd/install@latest
go run github.com/alexisbouchez/hyperseer/cmd/serve@latest
```

## CLI

### Installation

```bash
go install github.com/alexisbouchez/hyperseer/cmd/seer@latest
```

### Logs

```bash
# last hour (default)
seer logs

# filter by service and severity
seer logs --service my-api --severity error

# absolute time range
seer logs --from 2026-01-01T10:00:00Z --to 2026-01-01T11:00:00Z

# shorthand: last 30 minutes
seer logs --last 30m

# limit results
seer logs --limit 100
```

### Traces

```bash
# list recent root spans
seer traces

# filter by service
seer traces --service my-api

# absolute time range
seer traces --from 2026-01-01T10:00:00Z --to 2026-01-01T11:00:00Z

# shorthand: last 15 minutes
seer traces --last 15m

# inspect a specific trace (waterfall view)
seer traces <trace-id>
```

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE.md).
