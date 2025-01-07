package pubsubx

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

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
			PoisonQueue: NewPoisonQueueConfig(
				true,
				"poison-queue",
			),
			TopicRetry: true,
			ConsumerGroupMonitoring: NewConsumerGroupMonitoringConfig(
				true,
				time.Duration(10*time.Minute),
				time.Duration(1*time.Minute),
			),
		}

		rawConfig, err := sjson.Set("{}", "pubsubx", &conf)
		require.NoError(t, err)

		require.NoError(t, c.AddResource("config", bytes.NewBufferString(fmt.Sprintf(rootSchema, ConfigSchemaID))))

		schema, err := c.Compile(context.Background(), "config")
		require.NoError(t, err)

		assert.NoError(t, schema.Validate(bytes.NewBufferString(rawConfig)))
	})
}
