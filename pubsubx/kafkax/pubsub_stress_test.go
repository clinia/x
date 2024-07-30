//go:build stress
// +build stress

package kafkax_test

import (
	"runtime"
	"testing"

	"github.com/ThreeDotsLabs/watermill/pubsub/tests"
	"github.com/clinia/x/pubsubx/kafkax"
)

func init() {
	// Set GOMAXPROCS to double the number of CPUs
	runtime.GOMAXPROCS(runtime.GOMAXPROCS(0) * 2)
}

func TestPublishSubscribe_stress(t *testing.T) {
	t.Run("default consumption model", func(t *testing.T) {
		tests.TestPubSubStressTest(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     false,
				Persistent:          true,
			},
			createPubSub(kafkax.Default),
			createPubSubWithConsumerGroup(kafkax.Default),
		)
	})

	t.Run("batch consumption model", func(t *testing.T) {
		testBulkMessageHandlerPubSubStressTest(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     true,
				Persistent:          true,
			},
			createPartitionedPubSub(kafkax.Batch),
			createPubSubWithConsumerGroup(kafkax.Batch),
		)
	})

	t.Run("partition concurrent consumption model", func(t *testing.T) {
		tests.TestPubSubStressTest(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     false,
				Persistent:          true,
			},
			createPubSub(kafkax.PartitionConcurrent),
			createPubSubWithConsumerGroup(kafkax.PartitionConcurrent),
		)
	})
}

func TestPublishSubscribe_ordered_stress(t *testing.T) {
	t.Run("default consumption model", func(t *testing.T) {
		tests.TestPubSubStressTest(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     true,
				Persistent:          true,
			},
			createPartitionedPubSub(kafkax.Default),
			createPubSubWithConsumerGroup(kafkax.Default),
		)
	})

	t.Run("batch consumption model", func(t *testing.T) {
		testBulkMessageHandlerPubSubStressTest(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     true,
				Persistent:          true,
			},
			createPartitionedPubSub(kafkax.Batch),
			createPubSubWithConsumerGroup(kafkax.Batch),
		)
	})

	t.Run("partition concurrent consumption model", func(t *testing.T) {
		tests.TestPubSubStressTest(
			t,
			tests.Features{
				ConsumerGroups:      true,
				ExactlyOnceDelivery: false,
				GuaranteedOrder:     true,
				Persistent:          true,
			},
			createPartitionedPubSub(kafkax.PartitionConcurrent),
			createPubSubWithConsumerGroup(kafkax.PartitionConcurrent),
		)
	})
}
