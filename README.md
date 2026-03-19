# Hyperseer

> Observability for AI agents

## Table of Contents

- [About](#about)
- [Self-hosting](#self-hosting)
- [CLI](#cli)
- [Authentication](#authentication)
- [License](#license)

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

By default the query API is open. To enable JWT verification, configure an auth provider on the server. The CLI discovers the provider automatically — no flags needed at login time.

### How it works

When you run `seer login`, the CLI calls `GET /auth/config` on the server (unauthenticated) and receives the provider details. It then opens a browser to complete the flow and stores the resulting token at `~/.config/hyperseer/token.json`.

```bash
seer login                                      # uses HYPERSEER_QUERY_URL
seer login --server https://hyperseer.example.com
```

### Supabase

```bash
HYPERSEER_AUTH_PROVIDER=supabase \
HYPERSEER_AUTH_URL=https://<project-ref>.supabase.co/auth/v1 \
HYPERSEER_JWT_SECRET=<jwt-secret> \
./serve
```

`HYPERSEER_JWT_SECRET` is the JWT secret from your Supabase project (Dashboard → Project Settings → API → JWT Secret).

### Keycloak

```bash
HYPERSEER_AUTH_PROVIDER=keycloak \
HYPERSEER_AUTH_URL=https://auth.your-domain.com \
HYPERSEER_AUTH_REALM=your-realm \
HYPERSEER_AUTH_CLIENT_ID=hyperseer-cli \
HYPERSEER_JWKS_URL=https://auth.your-domain.com/realms/your-realm/protocol/openid-connect/certs \
./serve
```

The Keycloak client must have the Authorization Code + PKCE flow enabled and `http://localhost:*` in its redirect URIs.

### Environment variables

| Variable | Description |
|---|---|
| `HYPERSEER_AUTH_PROVIDER` | `supabase` or `keycloak` |
| `HYPERSEER_AUTH_URL` | Provider base URL (advertised to the CLI) |
| `HYPERSEER_AUTH_REALM` | Keycloak realm (default: `hyperseer`) |
| `HYPERSEER_AUTH_CLIENT_ID` | Keycloak client ID (default: `hyperseer-cli`) |
| `HYPERSEER_JWT_SECRET` | Supabase JWT secret (HS256 validation) |
| `HYPERSEER_JWKS_URL` | Keycloak JWKS endpoint URL (RS256 validation) |

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE.md).
