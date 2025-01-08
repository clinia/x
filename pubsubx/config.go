package pubsubx

import (
	"bytes"
	_ "embed"
	"io"
	"math"
	"time"

	"github.com/clinia/x/pointerx"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	PoisonQueue             PoisonQueueConfig             `json:"poisonQueue"`
	Scope                   string                        `json:"scope"`
	Provider                string                        `json:"provider"`
	Providers               ProvidersConfig               `json:"providers"`
	TopicRetry              bool                          `json:"topicRetry"`
	ConsumerGroupMonitoring ConsumerGroupMonitoringConfig `json:"consumerGroup"`
	EnableAutoCommit        bool                          `json:"enableAutoCommit"`
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
	TracerProvider trace.TracerProvider
	Propagator     propagation.TextMapPropagator
	MeterProvider  metric.MeterProvider
	MaxMessageByte *int32
	RetentionMs    *int32
}

type PoisonQueueConfig struct {
	enabled   bool
	TopicName string `json:"topicName"`
}

func NewPoisonQueueConfig(enabled bool, topicName string) PoisonQueueConfig {
	return PoisonQueueConfig{
		enabled:   enabled,
		TopicName: topicName,
	}
}

func (pqc PoisonQueueConfig) IsEnabled() bool {
	return pqc.enabled && pqc.TopicName != ""
}

type ConsumerGroupMonitoringConfig struct {
	enabled         bool
	HealthTimeout   time.Duration `json:"healthTimeout"`
	RefreshInterval time.Duration `json:"refreshInterval"`
}

func NewConsumerGroupMonitoringConfig(enabled bool, healthTimeout time.Duration, refreshInterval time.Duration) ConsumerGroupMonitoringConfig {
	return ConsumerGroupMonitoringConfig{
		enabled:         enabled,
		HealthTimeout:   healthTimeout,
		RefreshInterval: refreshInterval,
	}
}

func (cgm ConsumerGroupMonitoringConfig) IsEnabled() bool {
	return cgm.enabled && cgm.HealthTimeout > 0 && cgm.RefreshInterval > 0
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

// WithMaxMessageByte specifies the max message size in bytes.
// If none is specified, the default value is 1 MB.
func WithMaxMessageByte(max int32) PubSubOption {
	return func(opts *PubSubOptions) {
		opts.MaxMessageByte = pointerx.Ptr(max)
	}
}

// WithRetentionMs specifies the retention time in milliseconds.
func WithRetentionMs(retentionMs int32) PubSubOption {
	return func(opts *PubSubOptions) {
		opts.RetentionMs = pointerx.Ptr(retentionMs)
	}
}

type SubscriberOptions struct {
	// MaxBatchSize max amount of elements the batch will contain.
	// Default value is 100 if nothing is specified.
	MaxBatchSize uint16
	// MaxTopicRetryCount indicate how many time we allow to push to
	// the retry topic before considering a retryable error non retryable
	MaxTopicRetryCount uint16
	// Allows the handler to run each topic handling in parallel
	EnableAsyncExecution bool
	// Define the number of maximum topic handler that can run in parallel
	// on record processing
	MaxParallelAsyncExecution int16
}

func NewDefaultSubscriberOptions() *SubscriberOptions {
	return &SubscriberOptions{
		MaxBatchSize:              100,
		MaxTopicRetryCount:        3,
		EnableAsyncExecution:      false,
		MaxParallelAsyncExecution: -1,
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

func WithMaxTopicRetryCount(maxTopicRetryCount int) SubscriberOption {
	return func(o *SubscriberOptions) {
		if maxTopicRetryCount > math.MaxUint16 {
			o.MaxTopicRetryCount = math.MaxUint16
			return
		} else {
			//#nosec G115 -- Remove once https://github.com/securego/gosec/issues/1187 is solved
			o.MaxTopicRetryCount = uint16(maxTopicRetryCount)
		}
	}
}

func WithAsyncExecution() SubscriberOption {
	return func(o *SubscriberOptions) {
		o.EnableAsyncExecution = true
	}
}

func WithMaxParalleAsyncExecution(max int16) SubscriberOption {
	return func(o *SubscriberOptions) {
		if o.MaxParallelAsyncExecution <= 0 {
			o.MaxParallelAsyncExecution = -1
		} else {
			o.MaxParallelAsyncExecution = max
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
