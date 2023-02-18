// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"

	"github.com/clinia/x/logrusx"
)

const testTracingComponent = "github.com/clinia/x/otelx"

func decodeResponseBody(t *testing.T, r *http.Request) []byte {
	var reader io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		var err error
		reader, err = gzip.NewReader(r.Body)
		if err != nil {
			t.Fatal(err)
		}
	case "deflate":
		var err error
		reader, err = zlib.NewReader(r.Body)
		if err != nil {
			t.Fatal(err)
		}

	default:
		reader = r.Body
	}
	respBody, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	return respBody
}

func TestOTLPTracer(t *testing.T) {
	done := make(chan struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := decodeResponseBody(t, r)

		var res tracepb.ExportTraceServiceRequest
		err := proto.Unmarshal(body, &res)
		require.NoError(t, err, "must be able to unmarshal traces")

		resourceSpans := res.GetResourceSpans()
		spans := resourceSpans[0].GetScopeSpans()[0].GetSpans()
		assert.Equal(t, len(spans), 1)

		assert.NotEmpty(t, spans[0].GetSpanId())
		assert.NotEmpty(t, spans[0].GetTraceId())
		assert.Equal(t, "testSpan", spans[0].GetName())
		assert.Equal(t, "testAttribute", spans[0].Attributes[0].Key)

		close(done)
	}))
	defer ts.Close()

	tsu, err := url.Parse(ts.URL)
	require.NoError(t, err)

	ot, err := New(testTracingComponent, logrusx.New("clinia/x", "1"), &Config{
		ServiceName: "Clinia X",
		Provider:    "otel",
		Providers: ProvidersConfig{
			OTLP: OTLPConfig{
				Protocol:  "http",
				ServerURL: tsu.Host,
				Insecure:  true,
				Sampling: OTLPSampling{
					SamplingRatio: 1,
				},
			},
		},
	})
	assert.NoError(t, err)

	trc := ot.Tracer()
	_, span := trc.Start(context.Background(), "testSpan")
	span.SetAttributes(attribute.Bool("testAttribute", true))
	span.End()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatalf("Test server did not receive spans")
	}
}
