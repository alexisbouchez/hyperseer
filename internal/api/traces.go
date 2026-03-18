package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alexisbouchez/hyperseer/internal/query"
)

func (a *API) handleTraces(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from, to := defaultTimeRange()
	if s := q.Get("from"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			from = t
		}
	}
	if s := q.Get("to"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			to = t
		}
	}

	limit := 50
	if s := q.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}

	spans, err := query.Traces(r.Context(), a.conn, query.TracesParams{
		Service: q.Get("service"),
		Since:   from,
		Until:   to,
		Limit:   limit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spans)
}

func (a *API) handleTraceSpans(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("id")

	spans, err := query.TraceSpans(r.Context(), a.conn, traceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spans)
}

func defaultTimeRange() (from, to time.Time) {
	to = time.Now()
	from = to.Add(-time.Hour)
	return
}
