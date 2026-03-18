package query

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Span struct {
	TraceID     string
	SpanID      string
	ParentID    string
	Name        string
	ServiceName string
	Kind        string
	StatusCode  string
	Time        time.Time
	Duration    int64 // nanoseconds
}

type TracesParams struct {
	Service string
	Since   time.Time
	Limit   int
}

func Traces(ctx context.Context, conn driver.Conn, p TracesParams) ([]Span, error) {
	q := fmt.Sprintf(`
		SELECT trace_id, span_id, parent_span_id, name, service_name, kind, status_code, time, duration
		FROM spans_index
		WHERE project_id = 1
		  AND parent_span_id = ''
		  AND time >= ?
		  AND (? = '' OR service_name = ?)
		ORDER BY time DESC
		LIMIT %d
	`, p.Limit)

	return scanSpans(conn.Query(ctx, q,
		p.Since,
		p.Service, p.Service,
	))
}

func TraceSpans(ctx context.Context, conn driver.Conn, traceID string) ([]Span, error) {
	return scanSpans(conn.Query(ctx,
		`SELECT trace_id, span_id, parent_span_id, name, service_name, kind, status_code, time, duration
		 FROM spans_index
		 WHERE project_id = 1 AND trace_id = ?
		 ORDER BY time ASC`,
		traceID,
	))
}

func scanSpans(rows driver.Rows, err error) ([]Span, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spans []Span
	for rows.Next() {
		var s Span
		if err := rows.Scan(&s.TraceID, &s.SpanID, &s.ParentID, &s.Name, &s.ServiceName, &s.Kind, &s.StatusCode, &s.Time, &s.Duration); err != nil {
			return nil, err
		}
		spans = append(spans, s)
	}
	return spans, rows.Err()
}
