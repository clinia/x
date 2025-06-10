package tracex

import (
	"context"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/otelx"
	"go.opentelemetry.io/otel/trace"
)

type loggerProvider func(ctx context.Context) *logrusx.Logger
type tracerProvider func(ctx context.Context) *otelx.Tracer

func ComponentName(packageName, structName string) string {
	return packageName + "." + structName
}

func Instrument(ctx context.Context, lp loggerProvider, tp tracerProvider, componentName string, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span, *logrusx.Logger) {
	ctx, span := tp(ctx).Tracer().Start(ctx, componentName+"."+name, opts...)
	l := lp(ctx).WithSpanStartOptions(opts...)
	return ctx, span, l
}
