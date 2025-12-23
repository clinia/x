package tracex

import (
	"context"

	"github.com/clinia/x/loggerx"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/otelx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type (
	loggerProvider     func(ctx context.Context) *logrusx.Logger
	tracerProvider     func(ctx context.Context) *otelx.Tracer
	loggerNextProvider func() *loggerx.Logger
)

const ComponentNameSeparator = "."

func ComponentName(packageName, structName string) string {
	return packageName + ComponentNameSeparator + structName
}

/*
This allows us to easily instrument our code with a unify way and reduce the boilerplate of instrumentation. Reduce the possibles errors. `span.End()` must be called at the end of using the span.

	const myComponentName = "xpackage.xStruct"

	func (xs *xStruct) instrument(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span, *logrusx.Logger) {
	    return tracex.Instrument(ctx, xs.l.Logger, xs.d.Tracer, myComponentName, name, opts...)
	}

	func (xs *xStruct) process(ctx context.Context) error {
		ctx, span, l := xs.instrument(ctx, "process")
		defer span.End()
	}
*/
func Instrument(ctx context.Context, lp loggerProvider, tp tracerProvider, componentName string, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span, *logrusx.Logger) {
	ctx, span := tp(ctx).Tracer().Start(ctx, ComponentName(componentName, name), opts...)
	l := lp(ctx).WithSpanStartOptions(opts...)
	return ctx, span, l
}

func InstrumentNext(ctx context.Context, lp loggerNextProvider, tp tracerProvider, componentName string, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span, *loggerx.Logger) {
	fullComponentName := ComponentName(componentName, name)
	ctx, span := tp(ctx).Tracer().Start(ctx, fullComponentName, opts...)
	l := lp().
		WithSpanStartOptions(opts...).
		WithFields(attribute.Key("component").String(fullComponentName))
	return ctx, span, l
}
