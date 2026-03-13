package otelx

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.38.0/dbconv"
)

type (
	numericInstrumentType interface {
		int64 | float64
	}
	nInstrument[N numericInstrumentType] interface {
		Record(ctx context.Context, incr N, options ...metric.RecordOption)
	}
	InstrumentWithDefaults[N numericInstrumentType, InstType nInstrument[N]] struct {
		t           InstType
		defaultOpts []metric.RecordOption
	}

	nInstrumentWithDBSystem[N numericInstrumentType] interface {
		Record(ctx context.Context, val N, systemName dbconv.SystemNameAttr, attrs ...attribute.KeyValue)
	}

	InstrumentWithDBSystemAndDefaults[N numericInstrumentType, InstType nInstrumentWithDBSystem[N]] struct {
		t            InstType
		defaultAttrs []attribute.KeyValue
	}
)

var (
	_ nInstrument[int64]               = (*InstrumentWithDefaults[int64, nInstrument[int64]])(nil)
	_ nInstrument[float64]             = (*InstrumentWithDefaults[float64, nInstrument[float64]])(nil)
	_ nInstrumentWithDBSystem[int64]   = (*InstrumentWithDBSystemAndDefaults[int64, nInstrumentWithDBSystem[int64]])(nil)
	_ nInstrumentWithDBSystem[float64] = (*InstrumentWithDBSystemAndDefaults[float64, nInstrumentWithDBSystem[float64]])(nil)
)

func NewInstrumentWithDefaults[N numericInstrumentType, InstType nInstrument[N]](t InstType, defaults ...metric.RecordOption) *InstrumentWithDefaults[N, InstType] {
	return &InstrumentWithDefaults[N, InstType]{
		t:           t,
		defaultOpts: defaults,
	}
}

// Record implements [nInstrument].
func (i *InstrumentWithDefaults[N, InstType]) Record(ctx context.Context, incr N, options ...metric.RecordOption) {
	i.t.Record(ctx, incr, append(i.defaultOpts, options...)...)
}

func NewInstrumentWithDBSystemAndDefaults[N int64 | float64, InstType nInstrumentWithDBSystem[N]](t InstType, defaultAttrs ...attribute.KeyValue) *InstrumentWithDBSystemAndDefaults[N, InstType] {
	return &InstrumentWithDBSystemAndDefaults[N, InstType]{
		t:            t,
		defaultAttrs: defaultAttrs,
	}
}

// Record implements [nInstrumentWithDBSystem].
func (i *InstrumentWithDBSystemAndDefaults[N, InstType]) Record(ctx context.Context, val N, systemName dbconv.SystemNameAttr, attrs ...attribute.KeyValue) {
	i.t.Record(ctx, val, systemName, append(i.defaultAttrs, attrs...)...)
}
