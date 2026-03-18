package main

import (
	"log/slog"
	"net/http"

	"github.com/alexisbouchez/hyperseer/internal/api"
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

	errc := make(chan error, 2)

	go func() {
		slog.Info("otlp receiver listening", "addr", cfg.Serve.OTLPAddr)
		errc <- http.ListenAndServe(cfg.Serve.OTLPAddr, receiver.New(conn).Handler())
	}()

	go func() {
		slog.Info("query api listening", "addr", cfg.Serve.QueryAddr)
		errc <- http.ListenAndServe(cfg.Serve.QueryAddr, api.New(conn, cfg.Auth).Handler())
	}()

	exit.WithError(<-errc)
}
