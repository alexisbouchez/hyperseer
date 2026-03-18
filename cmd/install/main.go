package main

import (
	"fmt"
	"log"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/db"
	"github.com/alexisbouchez/hyperseer/internal/schema"
)

func main() {
	cfg := config.New()

	conn, err := db.New(cfg.ClickHouse)
	if err != nil {
		log.Fatalf("clickhouse: %v", err)
	}
	defer conn.Close()

	if err := schema.Migrate(conn); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	fmt.Println("\033[32m•\033[0m ClickHouse schema ready")
}
