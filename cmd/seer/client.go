package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alexisbouchez/hyperseer/internal/query"
)

func fetchLogs(baseURL string, p query.LogsParams) ([]query.Log, error) {
	u, err := url.Parse(baseURL + "/v1/logs")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("from", p.Since.Format(time.RFC3339))
	q.Set("to", p.Until.Format(time.RFC3339))
	q.Set("limit", strconv.Itoa(p.Limit))
	if p.Service != "" {
		q.Set("service", p.Service)
	}
	if p.Severity != "" {
		q.Set("severity", p.Severity)
	}
	u.RawQuery = q.Encode()

	return get[[]query.Log](u.String())
}

func fetchTraces(baseURL string, p query.TracesParams) ([]query.Span, error) {
	u, err := url.Parse(baseURL + "/v1/traces")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("from", p.Since.Format(time.RFC3339))
	q.Set("to", p.Until.Format(time.RFC3339))
	q.Set("limit", strconv.Itoa(p.Limit))
	if p.Service != "" {
		q.Set("service", p.Service)
	}
	u.RawQuery = q.Encode()

	return get[[]query.Span](u.String())
}

func fetchTraceSpans(baseURL, traceID string) ([]query.Span, error) {
	return get[[]query.Span](baseURL + "/v1/traces/" + traceID)
}

func get[T any](rawURL string) (T, error) {
	var zero T
	resp, err := http.Get(rawURL)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("query api returned %s", resp.Status)
	}
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, err
	}
	return result, nil
}
