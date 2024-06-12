package pubsubx

import (
	"bytes"
	_ "embed"
	"io"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	Scope     string          `json:"scope"`
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

type pubSubOptions struct {
	SaramaPublisherConfig  *sarama.Config
	SaramaSubscriberConfig *sarama.Config
	tracerProvider         trace.TracerProvider
	propagator             propagation.TextMapPropagator
}
type PubSubOption func(*pubSubOptions)

// WithSaramaPublisherConfig specifies the sarama config to use for the publisher.
func WithSaramaPublisherConfig(config *sarama.Config) PubSubOption {
	return func(opts *pubSubOptions) {
		opts.SaramaPublisherConfig = config
	}
}

// WithSaramaSubscriberConfig specifies the sarama config to use for the subscriber.
func WithSaramaSubscriberConfig(config *sarama.Config) PubSubOption {
	return func(opts *pubSubOptions) {
		opts.SaramaSubscriberConfig = config
	}
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, no tracer is configured
func WithTracerProvider(provider trace.TracerProvider) PubSubOption {
	return func(opts *pubSubOptions) {
		if provider != nil {
			opts.tracerProvider = provider
		}
	}
}

func WithPropagator(propagator propagation.TextMapPropagator) PubSubOption {
	return func(opts *pubSubOptions) {
		if propagator != nil {
			opts.propagator = propagator
		}
	}
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
