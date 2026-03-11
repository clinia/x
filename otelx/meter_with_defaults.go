package otelx

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.32.0/dbconv"
)

type (
	numericInstrumentType interface {
		int64 | float64
	}
	nInstrument[N numericInstrumentType] interface {
		Record(ctx context.Context, incr N, options ...metric.RecordOption)
	}
	instrumentWithDefaults[N numericInstrumentType, T nInstrument[N]] struct {
		t           T
		defaultOpts []metric.RecordOption
	}

	nInstrumentWithDBSystem[N numericInstrumentType] interface {
		Record(ctx context.Context, val N, systemName dbconv.SystemNameAttr, attrs ...attribute.KeyValue)
	}

	instrumentWithDBSystemAndDefaults[N numericInstrumentType, T nInstrumentWithDBSystem[N]] struct {
		t            T
		defaultAttrs []attribute.KeyValue
	}
)

var (
	_ nInstrument[int64]               = (*instrumentWithDefaults[int64, nInstrument[int64]])(nil)
	_ nInstrument[float64]             = (*instrumentWithDefaults[float64, nInstrument[float64]])(nil)
	_ nInstrumentWithDBSystem[int64]   = (*instrumentWithDBSystemAndDefaults[int64, nInstrumentWithDBSystem[int64]])(nil)
	_ nInstrumentWithDBSystem[float64] = (*instrumentWithDBSystemAndDefaults[float64, nInstrumentWithDBSystem[float64]])(nil)
)

func NewInstrumentWithDefaults[N numericInstrumentType, T nInstrument[N]](t T, defaults ...metric.RecordOption) *instrumentWithDefaults[N, T] {
	return &instrumentWithDefaults[N, T]{
		t:           t,
		defaultOpts: defaults,
	}
}

// Record implements [nInstrument].
func (i *instrumentWithDefaults[N, T]) Record(ctx context.Context, incr N, options ...metric.RecordOption) {
	i.t.Record(ctx, incr, append(i.defaultOpts, options...)...)
}

func NewInstrumentWithDBSystemAndDefaults[N int64 | float64, T nInstrumentWithDBSystem[N]](t T, defaultAttrs ...attribute.KeyValue) *instrumentWithDBSystemAndDefaults[N, T] {
	return &instrumentWithDBSystemAndDefaults[N, T]{
		t:            t,
		defaultAttrs: defaultAttrs,
	}
}

// Record implements [nInstrumentWithDBSystem].
func (i *instrumentWithDBSystemAndDefaults[N, T]) Record(ctx context.Context, val N, systemName dbconv.SystemNameAttr, attrs ...attribute.KeyValue) {
	i.t.Record(ctx, val, systemName, append(i.defaultAttrs, attrs...)...)
}
