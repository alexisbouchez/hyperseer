package receiver

import (
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Receiver struct {
	conn driver.Conn
}

func New(conn driver.Conn) *Receiver {
	return &Receiver{conn: conn}
}

func (r *Receiver) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/traces", r.handleTraces)
	mux.HandleFunc("POST /v1/logs", r.handleLogs)
	return mux
}
