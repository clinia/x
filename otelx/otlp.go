// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupOTLPTracer(tracerName string, c *TracerConfig) (trace.Tracer, propagation.TextMapPropagator, error) {
	exp, err := getExporter(c)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	atts := append([]attribute.KeyValue{}, semconv.ServiceNameKey.String(c.ServiceName))
	atts = append(atts, c.ResourceAttributes...)

	tpOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			atts...,
		)),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(
			c.Providers.OTLP.Sampling.SamplingRatio,
		))),
	}

	tp := sdktrace.NewTracerProvider(tpOpts...)

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	return tp.Tracer(tracerName), prop, nil
}

func getExporter(c *TracerConfig) (*otlptrace.Exporter, error) {
	ctx := context.Background()

	if c.Providers.OTLP.Protocol == "http" {
		clientOpts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(c.Providers.OTLP.ServerURL),
		}

		if c.Providers.OTLP.Insecure {
			clientOpts = append(clientOpts, otlptracehttp.WithInsecure())
		}

		exp, err := otlptrace.New(
			ctx, otlptracehttp.NewClient(clientOpts...),
		)
		if err != nil {
			return nil, err
		}
		return exp, nil
	}

	if c.Providers.OTLP.Protocol == "grpc" {
		// Set up a connection to the OTLP server.
		conn, err := grpc.DialContext(ctx, c.Providers.OTLP.ServerURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			errors.Errorf("failed to connect to OTLP gRPC endpoint: %s", err)
		}

		// Set up a trace exporter
		exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, errors.Errorf("failed to create trace exporter: %s", err)
		}

		return exp, nil
	}

	return nil, errors.Errorf("unknown protocol: %s", c.Providers.OTLP.Protocol)
}
