// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"bytes"
	_ "embed"
	"io"

	"go.opentelemetry.io/otel/attribute"
)

type JaegerConfig struct {
	LocalAgentAddress string         `json:"local_agent_address"`
	Sampling          JaegerSampling `json:"sampling"`
}

type JaegerSampling struct {
	ServerURL    string  `json:"server_url"`
	TraceIdRatio float64 `json:"trace_id_ratio"`
}

type OTLPConfig struct {
	Protocol  string       `json:"protocol"`
	ServerURL string       `json:"server_url"`
	Insecure  bool         `json:"insecure"`
	Sampling  OTLPSampling `json:"sampling"`
}

type OTLPSampling struct {
	SamplingRatio float64 `json:"sampling_ratio"`
}

type StdoutConfig struct {
	Pretty bool `json:"pretty"`
}

type ProvidersConfig struct {
	Jaeger JaegerConfig `json:"jaeger"`
	OTLP   OTLPConfig   `json:"otlp"`
	Stdout StdoutConfig `json:"stdout"`
}

type Config struct {
	ServiceName        string               `json:"service_name"`
	Provider           string               `json:"provider"`
	Providers          ProvidersConfig      `json:"providers"`
	ResourceAttributes []attribute.KeyValue `json:"resource_attributes"`
}

//go:embed config.schema.json
var ConfigSchema string

const ConfigSchemaID = "clinia://tracing-config"

// AddConfigSchema adds the tracing schema to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddConfigSchema(c interface {
	AddResource(url string, r io.Reader) error
}) error {
	return c.AddResource(ConfigSchemaID, bytes.NewBufferString(ConfigSchema))
}
