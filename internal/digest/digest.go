package digest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ServiceStats struct {
	Name   string
	Spans  uint64
	Errors uint64
	AvgMs  float64
	P95Ms  float64
}

type ErrorSpan struct {
	Time        time.Time
	ServiceName string
	Name        string
	DurationMs  float64
}

type ErrorLog struct {
	Time        time.Time
	ServiceName string
	Severity    string
	Body        string
}

func Gather(ctx context.Context, conn driver.Conn, since time.Time) (string, error) {
	stats, err := serviceStats(ctx, conn, since)
	if err != nil {
		return "", err
	}

	errorSpans, err := recentErrorSpans(ctx, conn, since)
	if err != nil {
		return "", err
	}

	errorLogs, err := recentErrorLogs(ctx, conn, since)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Telemetry snapshot — last %s\n\n", time.Since(since).Round(time.Minute)))

	sb.WriteString("## Services\n")
	if len(stats) == 0 {
		sb.WriteString("(no data)\n")
	}
	for _, s := range stats {
		errRate := 0.0
		if s.Spans > 0 {
			errRate = float64(s.Errors) / float64(s.Spans) * 100
		}
		sb.WriteString(fmt.Sprintf("- %s: %d spans, %.0f%% errors, avg %.1fms, p95 %.1fms\n",
			s.Name, s.Spans, errRate, s.AvgMs, s.P95Ms))
	}

	sb.WriteString("\n## Recent error spans\n")
	if len(errorSpans) == 0 {
		sb.WriteString("(none)\n")
	}
	for _, s := range errorSpans {
		sb.WriteString(fmt.Sprintf("- [%s] %s › %s (%.1fms)\n",
			s.Time.Format("15:04:05"), s.ServiceName, s.Name, s.DurationMs))
	}

	sb.WriteString("\n## Recent error/warning logs\n")
	if len(errorLogs) == 0 {
		sb.WriteString("(none)\n")
	}
	for _, l := range errorLogs {
		sb.WriteString(fmt.Sprintf("- [%s] %s %s: %s\n",
			l.Time.Format("15:04:05"), l.ServiceName, l.Severity, l.Body))
	}

	return sb.String(), nil
}

func serviceStats(ctx context.Context, conn driver.Conn, since time.Time) ([]ServiceStats, error) {
	rows, err := conn.Query(ctx, `
		SELECT
			service_name,
			count() AS spans,
			countIf(status_code = 'STATUS_CODE_ERROR') AS errors,
			avg(duration) / 1e6 AS avg_ms,
			quantile(0.95)(duration) / 1e6 AS p95_ms
		FROM spans_index
		WHERE project_id = 1 AND time >= ?
		GROUP BY service_name
		ORDER BY spans DESC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []ServiceStats
	for rows.Next() {
		var s ServiceStats
		if err := rows.Scan(&s.Name, &s.Spans, &s.Errors, &s.AvgMs, &s.P95Ms); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func recentErrorSpans(ctx context.Context, conn driver.Conn, since time.Time) ([]ErrorSpan, error) {
	rows, err := conn.Query(ctx, `
		SELECT time, service_name, name, duration / 1e6 AS duration_ms
		FROM spans_index
		WHERE project_id = 1 AND status_code = 'STATUS_CODE_ERROR' AND time >= ?
		ORDER BY time DESC
		LIMIT 10
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var spans []ErrorSpan
	for rows.Next() {
		var s ErrorSpan
		if err := rows.Scan(&s.Time, &s.ServiceName, &s.Name, &s.DurationMs); err != nil {
			return nil, err
		}
		spans = append(spans, s)
	}
	return spans, rows.Err()
}

func recentErrorLogs(ctx context.Context, conn driver.Conn, since time.Time) ([]ErrorLog, error) {
	rows, err := conn.Query(ctx, `
		SELECT time, service_name, severity, body
		FROM logs_index
		WHERE project_id = 1
		  AND lower(severity) IN ('error','fatal','warn','warning')
		  AND time >= ?
		ORDER BY time DESC
		LIMIT 10
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ErrorLog
	for rows.Next() {
		var l ErrorLog
		if err := rows.Scan(&l.Time, &l.ServiceName, &l.Severity, &l.Body); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
