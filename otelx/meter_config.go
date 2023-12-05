// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"bytes"
	_ "embed"
	"io"

	"go.opentelemetry.io/otel/attribute"
)

type OTLPMeterConfig struct {
	Protocol  string `json:"protocol"`
	ServerURL string `json:"server_url"`
	Insecure  bool   `json:"insecure"`
}

type MeterProvidersConfig struct {
	OTLP OTLPMeterConfig `json:"otlp"`
}

type MeterConfig struct {
	ServiceName        string               `json:"service_name"`
	Name               string               `json:"name"`
	Provider           string               `json:"provider"`
	Providers          MeterProvidersConfig `json:"providers,omitempty"`
	ResourceAttributes []attribute.KeyValue `json:"resource_attributes"`
}

//go:embed meter.schema.json
var MeterConfigSchema string

const MeterConfigSchemaID = "clinia://meter-config"

// AddMeterConfigSchema adds the meter schema to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddMeterConfigSchema(c interface {
	AddResource(url string, r io.Reader) error
}) error {
	return c.AddResource(MeterConfigSchemaID, bytes.NewBufferString(MeterConfigSchema))
}
