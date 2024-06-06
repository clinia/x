package pubsubx

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/pubsubx/kafkax"
	"github.com/stretchr/testify/assert"
)

func initBroker(t *testing.T) *sarama.MockBroker {
	broker := sarama.NewMockBrokerAddr(t, 0, "localhost:9092")
	broker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest": sarama.NewMockMetadataResponse(t).
			SetBroker(broker.Addr(), broker.BrokerID()).
			SetController(broker.BrokerID()),
	})
	return broker
}

func TestSetupKafka(t *testing.T) {
	l := logrusx.New("", "")

	broker := initBroker(t)
	defer broker.Close()

	conf := &Config{
		Provider: "kafka",
		Providers: ProvidersConfig{
			Kafka: KafkaConfig{
				Brokers: []string{broker.Addr()},
			},
		},
	}

	t.Run("SaramaConfig is nil", func(t *testing.T) {
		opts := &pubSubOptions{}

		pub, err := setupKafkaPublisher(l, conf, opts)
		assert.NoError(t, err)
		pub.(*kafkax.Publisher)

		_, err = setupKafkaSubscriber(l, conf, opts, "group", &subscriberOptions{}, nil)
		assert.NoError(t, err)
	})

	t.Run("with SaramaConfig", func(t *testing.T) {
		opts := &pubSubOptions{}

		_, err := setupKafkaPublisher(l, conf, opts)
		assert.NoError(t, err)

		_, err = setupKafkaSubscriber(l, conf, opts, "", &subscriberOptions{}, nil)
		assert.NoError(t, err)
	})
}
