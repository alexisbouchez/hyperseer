package main

import (
	"fmt"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/db"
	"github.com/alexisbouchez/hyperseer/internal/exit"
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

	fmt.Println("\033[32m•\033[0m ClickHouse schema ready")
}
