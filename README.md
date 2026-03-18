# Hyperseer

> Observability for AI agents

## About

Hyperseer is a self-hosted OpenTelemetry backend that stores logs and traces in ClickHouse and exposes them through a command-line interface (CLI).

## Self-hosting

Run the interactive install wizard. It sets up ClickHouse, the Hyperseer server, and optionally Caddy for HTTPS:

```bash
go run github.com/alexisbouchez/hyperseer/cmd/install@latest
```

All prompts can be passed as flags for non-interactive installs:

```bash
go run github.com/alexisbouchez/hyperseer/cmd/install@latest \
  --domain otel.example.com \
  --password secret \
  --retention 30 \
  --caddy new \
  --yes
```

| Flag | Description | Default |
|---|---|---|
| `--domain` | Domain where Hyperseer will be accessible | prompted |
| `--password` | ClickHouse password | auto-generated |
| `--retention` | Data retention in days (`7`, `30`, `90`, `365`) | `30` |
| `--caddy` | Proxy mode: `new`, `existing`, or `skip` | prompted |
| `--yes` | Skip confirmation prompt | `false` |

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

## Authentication

By default the query API is open. Set one of the following environment variables on the server to enable JWT verification.

### Supabase

Set `HYPERSEER_JWT_SECRET` to the JWT secret from your Supabase project (Dashboard → Project Settings → API → JWT Secret):

```bash
HYPERSEER_JWT_SECRET=your-supabase-jwt-secret ./serve
```

Then log in from the CLI:

```bash
# Supabase Cloud
seer login --provider supabase --url https://<project-ref>.supabase.co/auth/v1

# Self-hosted
seer login --provider supabase --url https://auth.your-domain.com
```

### Keycloak

Set `HYPERSEER_JWKS_URL` to the JWKS endpoint of your Keycloak realm:

```bash
HYPERSEER_JWKS_URL=https://auth.your-domain.com/realms/your-realm/protocol/openid-connect/certs ./serve
```

Then log in from the CLI:

```bash
seer login --provider keycloak --url https://auth.your-domain.com --realm your-realm
```

The client ID defaults to `hyperseer-cli`. Your Keycloak client must have the Authorization Code flow enabled and `http://localhost:*` in its redirect URIs.

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE.md).
