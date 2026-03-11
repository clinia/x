package otelx

import (
	"context"

	"go.opentelemetry.io/otel/metric"
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
)

var (
	_ nInstrument[int64]   = (*instrumentWithDefaults[int64, nInstrument[int64]])(nil)
	_ nInstrument[float64] = (*instrumentWithDefaults[float64, nInstrument[float64]])(nil)
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
