package config

import (
	"strconv"

	"github.com/alexisbouchez/hyperseer/internal/env"
)

type Config struct {
	ClickHouse ClickHouseConfig
	Serve      ServeConfig
	Digest     DigestConfig
}

type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type ServeConfig struct {
	OTLPAddr  string
	QueryAddr string
}

type DigestConfig struct {
	AnthropicAPIKey string
}

func New() Config {
	port, _ := strconv.Atoi(env.GetVar("HYPERSEER_CH_PORT", "9000"))
	return Config{
		ClickHouse: ClickHouseConfig{
			Host:     env.GetVar("HYPERSEER_CH_HOST", "localhost"),
			Port:     port,
			Database: env.GetVar("HYPERSEER_CH_DATABASE", "hyperseer"),
			Username: env.GetVar("HYPERSEER_CH_USERNAME", "hyperseer"),
			Password: env.GetVar("HYPERSEER_CH_PASSWORD", "hyperseer"),
		},
		Serve: ServeConfig{
			OTLPAddr:  env.GetVar("HYPERSEER_OTLP_ADDR", ":4318"),
			QueryAddr: env.GetVar("HYPERSEER_QUERY_ADDR", ":7777"),
		},
		Digest: DigestConfig{
			AnthropicAPIKey: env.GetVar("ANTHROPIC_API_KEY", ""),
		},
	}
}
