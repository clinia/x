package elasticx

import "context"

// ClientEngines provides access to the engines in a single elastic server, or an entire cluster of elastic servers.
type ClientEngines interface {
	// Engine opens a connection to an exisiting engine.
	// If no engine with given name exists, a NotFoundError is returned.
	Engine(ctx context.Context, name string) (Engine, error)

	// EngineExists returns true if an engine with given name exists.
	EngineExists(ctx context.Context, name string) (bool, error)

	// Engines returns a list of all engines found by the client.
	Engines(ctx context.Context) ([]Engine, error)

	// CreateEngine creates a new engine with given name and opens a connection to it.
	// If the a database with given name already exists, a DuplicateError is returned.
	CreateEngine(ctx context.Context, name string, options *CreateEngineOptions) (Engine, error)
}

// CreateEngineOptions contains options that customize the creation of a database.
type CreateEngineOptions struct {
}
