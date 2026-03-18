package api

import (
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/alexisbouchez/hyperseer/internal/config"
)

type API struct {
	conn driver.Conn
	cfg  config.AuthConfig
}

func New(conn driver.Conn, cfg config.AuthConfig) *API {
	return &API{conn: conn, cfg: cfg}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/logs", a.handleLogs)
	mux.HandleFunc("GET /v1/traces", a.handleTraces)
	mux.HandleFunc("GET /v1/traces/{id}", a.handleTraceSpans)
	return authMiddleware(a.cfg)(mux)
}
