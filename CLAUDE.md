# Hyperseer

> Keep this file up-to-date whenever something changes in the project.

A ClickHouse + Go CLI tool for OpenTelemetry logs and traces, focused on terminal usage.

**Multi-tenancy:** Hyperseer is designed to be multi-tenant from the start. Every signal table will have `project_id` as the first column in `ORDER BY` for partition pruning. Tenant isolation is enforced at the query layer — no cross-tenant data leakage.

## Code style

- Prefer simple solutions over complex ones. When in doubt, do less.
- Small, focused packages with struct + constructor patterns: e.g. `config.New()`, not factory functions, registries, or global state.
- Finish what's in front of you before moving to the next thing. If something is broken or incomplete, fix it first.

## Project structure

```
hyperseer/
├── cmd/
│   ├── seer/main.go     — end-user CLI (urfave/cli/v3)
│   ├── serve/main.go    — OTLP receiver + query API daemon (no framework)
│   └── install/main.go  — schema setup + daemon registration (no framework)
├── internal/
│   ├── config/          — shared config struct
│   ├── db/              — ClickHouse connection
│   └── schema/          — table DDL / migrations
├── compose.yml          — ClickHouse dev environment
├── go.mod               — module: github.com/alexisbouchez/hyperseer, go 1.26.1
├── README.md
└── LICENSE.md           — AGPLv3
```

## CLI commands (planned)

- `seer login` — authentication (post-prototype)
- `seer logs` — query logs with timestamp/filter params
- `seer traces` — query traces (params TBD)
- `seer digest` — high-value summarized output (format TBD, should be great)
- `seer install` — install Claude Code skill to `~/.claude/commands/hyperseer.md`

## Current state

- urfave/cli/v3 skeleton in `cmd/seer/main.go`
- `seer logs` and `seer traces` implemented
- `seer install` writes `~/.claude/commands/hyperseer.md` and optionally injects into project CLAUDE.md
- No `seer digest` yet
- ClickHouse Go client: `github.com/ClickHouse/clickhouse-go/v2`
- Schema: `logs_index`, `logs_data`, `spans_index`, `spans_data` with `project_id` as first ORDER BY key

## ClickHouse (compose.yml)

- Image: `clickhouse/clickhouse-server:25.3`
- Credentials: `hyperseer / hyperseer`, DB: `hyperseer`
- Ports: `8123` (HTTP), `9000` (native TCP)
- Named volume: `clickhouse_data`

## Reference: Uptrace schema

Uptrace is licensed **AGPL-3.0** — do not copy its schema or code verbatim. Use as reference only.

### Hot/cold split pattern

Every signal has two tables:
- `_index` — hot, all queryable fields as top-level columns, used for filtering/aggregation
- `_data` — cold, full JSON payload, used for retrieval; joined at query time on `(trace_id, id)`

### Spans (`spans_index` / `spans_data`)

`spans_index` key columns:
- `project_id UInt32`, `type`, `system`, `group_id UInt64`
- `trace_id UUID`, `id UInt64`, `parent_id UInt64`
- `name`, `kind`, `time DateTime`, `duration Int64`, `status_code LowCardinality(String)`
- Promoted resource attrs: `service_name`, `host_name`, `deployment_environment`, …
- Promoted span attrs: `db_system`, `db_statement`, `client_address`, …
- Attributes as parallel arrays: `string_keys Array(LowCardinality(String))`, `string_values Array(String)`
- `ORDER BY (project_id, system, group_id, time)`

`spans_data`: `project_id`, `trace_id`, `id`, `time DateTime64(6)`, `data String` (full JSON)
- `ORDER BY (trace_id, id)`

### Logs (`logs_index` / `logs_data`)

Same structure as spans but without `duration`; adds:
- `log_severity Enum8(...)` — full OTel severity enum (TRACE → FATAL4)
- `log_file_path`, `exception_type`, `exception_stacktrace`

### Events (`events_index` / `events_data`)

Like logs; adds messaging attrs: `messaging_message_id`, `messaging_message_type`, `messaging_message_payload_size_bytes`.

### Unified read view

```sql
CREATE TABLE tracing_data ENGINE=Merge(currentDatabase(), '^(spans|events|logs)_data$')
```

### Metrics (`datapoint_minutes` / `datapoint_hours`)

`AggregatingMergeTree`, aggregated per `(project_id, metric, time, attrs_hash)`.
`datapoint_hours` is fed by a materialized view over `datapoint_minutes`.

### Key design notes

- Attributes stored as **parallel arrays**, not `Map` — predates ClickHouse native JSON type
- `project_id` is first in every `ORDER BY` for multi-tenant partition pruning
- `PARTITION BY toDate(time)` + `ttl_only_drop_parts = 1` on all signal tables

## OTel examples

`../otel-examples/` contains two instrumented apps used for local testing:

- `adonisjs/` — AdonisJS app with `@adonisjs/otel`
- `symfony/` — Symfony 7.4 app with `friendsofopentelemetry/opentelemetry-bundle`

never use return params!

## Writing style

- No em dashes
- Short sentences
