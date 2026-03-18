package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/alexisbouchez/hyperseer/internal/query"
)

func (a *API) handleLogs(w http.ResponseWriter, r *http.Request) {
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

	logs, err := query.Logs(r.Context(), a.conn, query.LogsParams{
		Service:  q.Get("service"),
		Severity: q.Get("severity"),
		Since:    from,
		Until:    to,
		Limit:    limit,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
