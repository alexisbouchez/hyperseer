package receiver

import (
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
)

func (r *Receiver) handleTraces(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var export coltracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	indexBatch, err := r.conn.PrepareBatch(req.Context(), "INSERT INTO spans_index")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dataBatch, err := r.conn.PrepareBatch(req.Context(), "INSERT INTO spans_data")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, rs := range export.ResourceSpans {
		serviceName := attrString(rs.GetResource().GetAttributes(), "service.name")

		for _, ss := range rs.ScopeSpans {
			for _, span := range ss.Spans {
				keys, vals := attrKV(span.Attributes)
				t := time.Unix(0, int64(span.StartTimeUnixNano))
				duration := int64(span.EndTimeUnixNano - span.StartTimeUnixNano)

				if err := indexBatch.Append(
					uint32(1),
					hexID(span.TraceId),
					hexID(span.SpanId),
					hexID(span.ParentSpanId),
					span.Name,
					span.Kind.String(),
					span.Status.GetCode().String(),
					span.Status.GetMessage(),
					serviceName,
					t,
					duration,
					keys,
					vals,
				); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				data, _ := protojson.Marshal(span)
				if err := dataBatch.Append(
					uint32(1),
					hexID(span.TraceId),
					hexID(span.SpanId),
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

	respond(w, &coltracepb.ExportTraceServiceResponse{})
}
