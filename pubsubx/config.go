package pubsubx

import (
	"bytes"
	_ "embed"
	"io"
	"math"

	"go.opentelemetry.io/otel/metric"
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

type PubSubOptions struct {
	TracerProvider                  trace.TracerProvider
	Propagator                      propagation.TextMapPropagator
	MeterProvider                   metric.MeterProvider
	DefaultCreateTopicConfigEntries map[string]*string
}

type PubSubOption func(*PubSubOptions)

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, no tracer is configured
func WithTracerProvider(provider trace.TracerProvider) PubSubOption {
	return func(opts *PubSubOptions) {
		if provider != nil {
			opts.TracerProvider = provider
		}
	}
}

func WithPropagator(propagator propagation.TextMapPropagator) PubSubOption {
	return func(opts *PubSubOptions) {
		if propagator != nil {
			opts.Propagator = propagator
		}
	}
}

func WithMeterProvider(provider metric.MeterProvider) PubSubOption {
	return func(opts *PubSubOptions) {
		if provider != nil {
			opts.MeterProvider = provider
		}
	}
}

func WithDefaultCreateTopicConfigEntries(entries map[string]*string) PubSubOption {
	return func(opts *PubSubOptions) {
		opts.DefaultCreateTopicConfigEntries = entries
	}
}

type SubscriberOptions struct {
	// MaxBatchSize max amount of elements the batch will contain.
	// Default value is 100 if nothing is specified.
	MaxBatchSize uint16
}

func NewDefaultSubscriberOptions() *SubscriberOptions {
	return &SubscriberOptions{
		MaxBatchSize: 100,
	}
}

type SubscriberOption func(*SubscriberOptions)

func WithMaxBatchSize(maxBatchSize int) SubscriberOption {
	return func(o *SubscriberOptions) {
		if maxBatchSize > math.MaxUint16 {
			o.MaxBatchSize = math.MaxUint16
			return
		} else {
			//#nosec G115 -- Remove once https://github.com/securego/gosec/issues/1187 is solved
			o.MaxBatchSize = uint16(maxBatchSize)
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
},
) error {
	return c.AddResource(ConfigSchemaID, bytes.NewBufferString(ConfigSchema))
}
