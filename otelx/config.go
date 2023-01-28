// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"bytes"
	_ "embed"
	"io"
)

type OTLPConfig struct {
	ServerURL string       `json:"server_url"`
	Insecure  bool         `json:"insecure"`
	Sampling  OTLPSampling `json:"sampling"`
}

type OTLPSampling struct {
	SamplingRatio float64 `json:"sampling_ratio"`
}

type ProvidersConfig struct {
	OTLP OTLPConfig `json:"otlp"`
}

type Config struct {
	ServiceName string          `json:"service_name"`
	Provider    string          `json:"provider"`
	Providers   ProvidersConfig `json:"providers"`
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
