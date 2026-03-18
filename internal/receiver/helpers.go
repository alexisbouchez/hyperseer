package receiver

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
)

func unmarshal(contentType string, body []byte, msg proto.Message) error {
	if strings.Contains(contentType, "application/json") {
		return protojson.Unmarshal(body, msg)
	}
	return proto.Unmarshal(body, msg)
}

func respond(w http.ResponseWriter, req *http.Request, msg proto.Message) {
	if strings.Contains(req.Header.Get("Content-Type"), "application/json") {
		out, _ := protojson.Marshal(msg)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}
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
