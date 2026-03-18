package api

import (
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type API struct {
	conn driver.Conn
}

func New(conn driver.Conn) *API {
	return &API{conn: conn}
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/logs", a.handleLogs)
	mux.HandleFunc("GET /v1/traces", a.handleTraces)
	mux.HandleFunc("GET /v1/traces/{id}", a.handleTraceSpans)
	return mux
}
