package kgox

//go:generate mockery --name=KgoClient --output=./mocks --filename=mock_kgo_client.go --structname=MockKgoClient

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// KgoClient is an interface that wraps the kgo.Client
// This is mainly for testing purposes so that we can mock the kgo.Client
type KgoClient interface {
	Close()
	PollRecords(ctx context.Context, maxPollRecords int) kgo.Fetches
}
