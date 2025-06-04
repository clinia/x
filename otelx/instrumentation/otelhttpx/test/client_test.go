// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/clinia/x/otelx/instrumentation/otelhttpx"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestConvenienceWrappers(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(sr))
	orig := otelhttpx.DefaultClient
	otelhttpx.DefaultClient = &http.Client{
		Transport: otelhttpx.NewTransport(
			http.DefaultTransport,
			otelhttpx.WithTracerProvider(provider),
		),
	}
	defer func() { otelhttpx.DefaultClient = orig }()

	content := []byte("Hello, world!")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	ctx := context.Background()
	res, err := otelhttpx.Get(ctx, ts.URL)
	require.NoError(t, err)

	require.NoError(t, res.Body.Close())

	res, err = otelhttpx.Head(ctx, ts.URL)
	require.NoError(t, err)

	require.NoError(t, res.Body.Close())

	res, err = otelhttpx.Post(ctx, ts.URL, "text/plain", strings.NewReader("test"))
	require.NoError(t, err)

	require.NoError(t, res.Body.Close())

	form := make(url.Values)
	form.Set("foo", "bar")
	res, err = otelhttpx.PostForm(ctx, ts.URL, form)
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, res.Body.Close())

	spans := sr.Ended()
	require.Equal(t, 4, len(spans))
	assert.Equal(t, "HTTP GET", spans[0].Name())
	assert.Equal(t, "HTTP HEAD", spans[1].Name())
	assert.Equal(t, "HTTP POST", spans[2].Name())
	assert.Equal(t, "HTTP POST", spans[3].Name())
}

func TestClientWithTraceContext(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(sr))

	tracer := provider.Tracer("")
	ctx, span := tracer.Start(context.Background(), "http requests")

	content := []byte("Hello, world!")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
	}))
	defer ts.Close()

	res, err := otelhttpx.Get(ctx, ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, res.Body.Close())

	span.End()

	spans := sr.Ended()
	require.Equal(t, 2, len(spans))
	assert.Equal(t, "HTTP GET", spans[0].Name())
	assert.Equal(t, "http requests", spans[1].Name())
	assert.NotEmpty(t, spans[0].Parent().SpanID())
	assert.Equal(t, spans[1].SpanContext().SpanID(), spans[0].Parent().SpanID())
}
