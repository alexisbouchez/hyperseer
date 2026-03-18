package schema_test

import (
	"testing"

	"github.com/alexisbouchez/hyperseer/internal/config"
	"github.com/alexisbouchez/hyperseer/internal/db"
	"github.com/alexisbouchez/hyperseer/internal/schema"
)

func TestMigrate(t *testing.T) {
	conn, err := db.New(config.New().ClickHouse)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := schema.Migrate(conn); err != nil {
		t.Fatal(err)
	}
}
