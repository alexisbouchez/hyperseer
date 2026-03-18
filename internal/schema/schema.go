package schema

import (
	"context"
	"embed"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

//go:embed sql/*.sql
var files embed.FS

func Migrate(conn driver.Conn) error {
	entries, err := files.ReadDir("sql")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		ddl, err := files.ReadFile("sql/" + entry.Name())
		if err != nil {
			return err
		}
		if err := conn.Exec(context.Background(), string(ddl)); err != nil {
			return err
		}
	}

	return nil
}
