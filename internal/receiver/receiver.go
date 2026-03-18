package receiver

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Receiver struct {
	conn driver.Conn
}

func New(conn driver.Conn) *Receiver {
	return &Receiver{conn: conn}
}

func (r *Receiver) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/traces", r.handleTraces)
	mux.HandleFunc("POST /v1/logs", r.handleLogs)
	return mux
}

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

func (r *Receiver) handleLogs(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var export collogspb.ExportLogsServiceRequest
	if err := proto.Unmarshal(body, &export); err != nil {
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

	respond(w, &collogspb.ExportLogsServiceResponse{})
}

func respond(w http.ResponseWriter, msg proto.Message) {
	out, _ := proto.Marshal(msg)
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.Write(out)
}

func hexID(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return hex.EncodeToString(b)
}

func attrString(attrs []*commonpb.KeyValue, key string) string {
	for _, a := range attrs {
		if a.Key == key {
			return a.Value.GetStringValue()
		}
	}
	return ""
}

func attrKV(attrs []*commonpb.KeyValue) ([]string, []string) {
	keys := make([]string, len(attrs))
	vals := make([]string, len(attrs))
	for i, a := range attrs {
		keys[i] = a.Key
		vals[i] = anyValueString(a.Value)
	}
	return keys, vals
}

func anyValueString(v *commonpb.AnyValue) string {
	if v == nil {
		return ""
	}
	switch x := v.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return x.StringValue
	case *commonpb.AnyValue_BoolValue:
		return fmt.Sprintf("%t", x.BoolValue)
	case *commonpb.AnyValue_IntValue:
		return fmt.Sprintf("%d", x.IntValue)
	case *commonpb.AnyValue_DoubleValue:
		return fmt.Sprintf("%g", x.DoubleValue)
	case *commonpb.AnyValue_ArrayValue:
		return fmt.Sprintf("%v", x.ArrayValue)
	case *commonpb.AnyValue_KvlistValue:
		return fmt.Sprintf("%v", x.KvlistValue)
	default:
		return ""
	}
}
