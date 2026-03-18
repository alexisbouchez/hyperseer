package db_test

import (
	"testing"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/db"
)

func TestNew(t *testing.T) {
	cfg := config.New()
	conn, err := db.New(cfg.ClickHouse)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}
