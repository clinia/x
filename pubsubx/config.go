package pubsubx

import (
	"bytes"
	_ "embed"
	"io"
)

type Config struct {
	Provider  string          `json:"provider"`
	Providers ProvidersConfig `json:"providers"`
}

type ProvidersConfig struct {
	InMemory InMemoryConfig `json:"inmemory"`
	Kafka    KafkaConfig    `json:"kafka"`
}

type InMemoryConfig struct{}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
}

//go:embed config.schema.json
var ConfigSchema string

const ConfigSchemaID = "clinia://pubsub-config"

// AddConfigSchema adds the tracing schema to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddConfigSchema(c interface {
	AddResource(url string, r io.Reader) error
}) error {
	return c.AddResource(ConfigSchemaID, bytes.NewBufferString(ConfigSchema))
}
