package main

import (
	"log/slog"
	"net/http"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/db"
	"github.com/alexisbouchez/hyperseer/internal/exit"
	"github.com/alexisbouchez/hyperseer/internal/receiver"
	"github.com/alexisbouchez/hyperseer/internal/schema"
)

func main() {
	cfg := config.New()

	conn, err := db.New(cfg.ClickHouse)
	if err != nil {
		exit.WithError(err)
	}
	defer conn.Close()

	if err := schema.Migrate(conn); err != nil {
		exit.WithError(err)
	}

	r := receiver.New(conn)

	slog.Info("http server listening for requests", "addr", cfg.Serve.OTLPAddr)
	if err := http.ListenAndServe(cfg.Serve.OTLPAddr, r.Handler()); err != nil {
		exit.WithError(err)
	}
}
