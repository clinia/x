package kgox

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
)

type State struct {
	LastConsumptionTimePerTopic map[string]time.Time
	TotalLagPerTopic            kadm.GroupTopicsLag
}

// monitor periodically refreshes the lag state of the consumer group.
func (c *consumer) monitor(ctx context.Context) {
	c.adminClient = kadm.NewClient(c.cl)
	defer c.adminClient.Close()

	ticker := time.NewTicker(c.conf.ConsumerGroupMonitoring.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refreshLagState(ctx)
		}
	}
}

// refreshLagState refreshes the lag state of the consumer group.
func (c *consumer) refreshLagState(ctx context.Context) error {
	groupLags, err := c.adminClient.Lag(ctx, c.group.ConsumerGroup(c.conf.Scope))
	if err != nil {
		return err
	}

	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	c.state.TotalLagPerTopic = groupLags[c.group.ConsumerGroup(c.conf.Scope)].Lag.TotalByTopic()

	return nil
}
