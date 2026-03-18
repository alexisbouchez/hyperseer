package receiver

import (
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

func (r *Receiver) handleLogs(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var export collogspb.ExportLogsServiceRequest
	if err := unmarshal(req.Header.Get("Content-Type"), body, &export); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	indexBatch, err := r.conn.PrepareBatch(req.Context(), "INSERT INTO logs_index")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataBatch, err := r.conn.PrepareBatch(req.Context(), "INSERT INTO logs_data")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, rl := range export.ResourceLogs {
		serviceName := attrString(rl.GetResource().GetAttributes(), "service.name")

		for _, sl := range rl.ScopeLogs {
			for _, record := range sl.LogRecords {
				keys, vals := attrKV(record.Attributes)
				t := time.Unix(0, int64(record.TimeUnixNano))

				if err := indexBatch.Append(
					uint32(1),
					hexID(record.TraceId),
					hexID(record.SpanId),
					serviceName,
					record.SeverityText,
					record.Body.GetStringValue(),
					t,
					keys,
					vals,
				); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				data, _ := protojson.Marshal(record)
				if err := dataBatch.Append(
					uint32(1),
					hexID(record.TraceId),
					hexID(record.SpanId),
					t,
					string(data),
				); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	}

	if err := indexBatch.Send(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := dataBatch.Send(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respond(w, req, &collogspb.ExportLogsServiceResponse{})
}
