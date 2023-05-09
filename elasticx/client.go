package elasticx

import "context"

// Client provides access to a single Elastic server, or an entire cluster of Elastic servers.
type Client interface {
	// Init makes sure that the elastic server or cluster is configured for clinia
	Init(ctx context.Context) error

	// Clean makes sure the the elastic server or cluster removes all clinia configuration
	Clean(ctx context.Context) error

	ClientEngines
}
