// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package otelx

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/jsonschema/v3"
)

const rootSchema = `{
  "properties": {
    "tracing": {
      "$ref": "%s"
    }
  }
}
`

func TestConfigSchema(t *testing.T) {
	t.Run("func=AddConfigSchema", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		require.NoError(t, AddTracerConfigSchema(c))

		conf := TracerConfig{
			ServiceName: "Clinia X",
			TracerName:  "X",
			Provider:    "otel",
			Providers: TracerProvidersConfig{
				OTLP: OTLPConfig{
					ServerURL: "http://localhost:5778/sampling",
					Sampling: OTLPSampling{
						SamplingRatio: 1,
					},
				},
			},
		}

		rawConfig, err := sjson.Set("{}", "otelx", &conf)
		require.NoError(t, err)

		require.NoError(t, c.AddResource("config", bytes.NewBufferString(fmt.Sprintf(rootSchema, TracerConfigSchemaID))))

		schema, err := c.Compile(context.Background(), "config")
		require.NoError(t, err)

		assert.NoError(t, schema.Validate(bytes.NewBufferString(rawConfig)))
	})
}
