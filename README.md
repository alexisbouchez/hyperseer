# Hyperseer

> Observability for AI agents

## About

Hyperseer is a self-hosted OpenTelemetry backend that stores logs and traces in ClickHouse and exposes them through a command-line interface (CLI).

## Self-hosting

### Interactive mode

```bash
go run github.com/alexisbouchez/hyperseer/cmd/install@latest
```

### Manual mode

```bash
# to be written
```

## Usage

### CLI

#### Installation

```bash
go install github.com/alexisbouchez/hyperseer/cmd/seer@latest
```

#### Authentication

```bash
seer login
```

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE.md).
