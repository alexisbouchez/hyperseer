package query

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Log struct {
	Time        time.Time
	Severity    string
	ServiceName string
	Body        string
}

type LogsParams struct {
	Service  string
	Severity string
	Since    time.Time
	Until    time.Time
	Limit    int
}

func Logs(ctx context.Context, conn driver.Conn, p LogsParams) ([]Log, error) {
	q := fmt.Sprintf(`
		SELECT time, severity, service_name, body
		FROM logs_index
		WHERE project_id = 1
		  AND time >= ? AND time <= ?
		  AND (? = '' OR service_name = ?)
		  AND (? = '' OR severity = ?)
		ORDER BY time DESC
		LIMIT %d
	`, p.Limit)

	rows, err := conn.Query(ctx, q,
		p.Since, p.Until,
		p.Service, p.Service,
		p.Severity, p.Severity,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var l Log
		if err := rows.Scan(&l.Time, &l.Severity, &l.ServiceName, &l.Body); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
