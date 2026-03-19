package config

import (
	"strconv"

	"github.com/alexisbouchez/hyperseer/internal/env"
)

type Config struct {
	ClickHouse ClickHouseConfig
	Serve      ServeConfig
	Digest     DigestConfig
	Auth       AuthConfig
	QueryURL   string
}

type AuthConfig struct {
	// Runtime JWT validation
	JWTSecret string // HS256 — Supabase
	JWKSUrl   string // RS256 — Keycloak

	// Advertised to CLI via GET /auth/config
	Provider string // "keycloak" or "supabase"
	URL      string // Provider base URL
	Realm    string // Keycloak realm
	ClientID string // Keycloak client ID
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
		Auth: AuthConfig{
			JWTSecret: env.GetVar("HYPERSEER_JWT_SECRET", ""),
			JWKSUrl:   env.GetVar("HYPERSEER_JWKS_URL", ""),
			Provider:  env.GetVar("HYPERSEER_AUTH_PROVIDER", ""),
			URL:       env.GetVar("HYPERSEER_AUTH_URL", ""),
			Realm:     env.GetVar("HYPERSEER_AUTH_REALM", "hyperseer"),
			ClientID:  env.GetVar("HYPERSEER_AUTH_CLIENT_ID", "hyperseer-cli"),
		},
		QueryURL: env.GetVar("HYPERSEER_QUERY_URL", "http://localhost:7777"),
	}
}
