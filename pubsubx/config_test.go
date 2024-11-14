package pubsubx

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
    "pubsub": {
      "$ref": "%s"
    }
  }
}
`

func TestConfigSchema(t *testing.T) {
	t.Run("func=AddConfigSchema", func(t *testing.T) {
		c := jsonschema.NewCompiler()
		require.NoError(t, AddConfigSchema(c))

		conf := Config{
			Provider: "kafka",
			Providers: ProvidersConfig{
				Kafka: KafkaConfig{
					Brokers: []string{"localhost:9092"},
				},
			},
			PoisonQueue: PoisonQueueConfig{
				Enabled:   true,
				TopicName: "poison-queue",
			},
			TopicRetry: true,
		}

		rawConfig, err := sjson.Set("{}", "pubsubx", &conf)
		require.NoError(t, err)

		require.NoError(t, c.AddResource("config", bytes.NewBufferString(fmt.Sprintf(rootSchema, ConfigSchemaID))))

		schema, err := c.Compile(context.Background(), "config")
		require.NoError(t, err)

		assert.NoError(t, schema.Validate(bytes.NewBufferString(rawConfig)))
	})
}
