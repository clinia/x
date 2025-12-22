package slogx

import (
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
)

func NewLogFields(kvs ...attribute.KeyValue) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(kvs))
	for _, kv := range kvs {
		switch kv.Value.Type() {
		case attribute.BOOL:
			attrs = append(attrs, slog.Bool(string(kv.Key), kv.Value.AsBool()))
		case attribute.INT64:
			attrs = append(attrs, slog.Int64(string(kv.Key), kv.Value.AsInt64()))
		case attribute.FLOAT64:
			attrs = append(attrs, slog.Float64(string(kv.Key), kv.Value.AsFloat64()))
		case attribute.STRING:
			attrs = append(attrs, slog.String(string(kv.Key), kv.Value.AsString()))
		case attribute.BOOLSLICE, attribute.INT64SLICE, attribute.FLOAT64SLICE, attribute.STRINGSLICE:
			attrs = append(attrs, slog.Any(string(kv.Key), kv.Value.AsInterface()))
		default:
			attrs = append(attrs, slog.Any(string(kv.Key), kv.Value.AsInterface()))
		}
	}
	return attrs
}

func ErrorAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
