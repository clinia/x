// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"

	"github.com/clinia/x/logrusx"
)

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

func TestHTTPOTLPTracer(t *testing.T) {
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

	tracerConfig := &TracerConfig{
		ServiceName: "Clinia X",
		Name:        "X",
		Provider:    "otel",
		Providers: TracerProvidersConfig{
			OTLP: OTLPConfig{
				Protocol:  "http",
				ServerURL: tsu.Host,
				Insecure:  true,
				Sampling: OTLPSampling{
					SamplingRatio: 1,
				},
			},
		},
	}

	tr := &Tracer{}
	err = tr.setup(logrusx.New("clinia/x", "1"), tracerConfig)
	assert.NoError(t, err)

	_, span := tr.Tracer().Start(context.Background(), "testSpan")
	span.SetAttributes(attribute.Bool("testAttribute", true))
	span.End()

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatalf("Test server did not receive spans")
	}
}

type TraceServiceServer struct {
	tracepb.TraceServiceServer
	t    *testing.T
	done chan struct{}
}

func (s *TraceServiceServer) Export(ctx context.Context, req *tracepb.ExportTraceServiceRequest) (*tracepb.ExportTraceServiceResponse, error) {
	resourceSpans := req.GetResourceSpans()
	spans := resourceSpans[0].GetScopeSpans()[0].GetSpans()
	assert.Equal(s.t, len(spans), 1)

	assert.NotEmpty(s.t, spans[0].GetSpanId())
	assert.NotEmpty(s.t, spans[0].GetTraceId())
	assert.Equal(s.t, "testSpan", spans[0].GetName())
	assert.Equal(s.t, "testAttribute", spans[0].Attributes[0].Key)

	close(s.done)

	return nil, nil
}

func TestGRPCOTLPTracer(t *testing.T) {
	done := make(chan struct{})

	grpcServer := grpc.NewServer()
	service := &TraceServiceServer{t: t, done: done}

	tracepb.RegisterTraceServiceServer(grpcServer, service)
	lis, err := net.Listen("tcp", "localhost:0") // Listen on a random available port
	assert.NoError(t, err)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
	defer grpcServer.Stop()

	tracerConfig := &TracerConfig{
		ServiceName: "Clinia X",
		Name:        "X",
		Provider:    "otel",
		Providers: TracerProvidersConfig{
			OTLP: OTLPConfig{
				Protocol:  "grpc",
				ServerURL: lis.Addr().String(),
				Insecure:  true,
				Sampling: OTLPSampling{
					SamplingRatio: 1,
				},
			},
		},
	}

	tr := &Tracer{}
	err = tr.setup(logrusx.New("clinia/x", "1"), tracerConfig)
	assert.NoError(t, err)

	_, span := tr.Tracer().Start(context.Background(), "testSpan")
	span.SetAttributes(attribute.Bool("testAttribute", true))
	span.End()

	select {
	case <-service.done:
	case <-time.After(15 * time.Second):
		t.Fatalf("Test server did not receive spans")
	}
}
