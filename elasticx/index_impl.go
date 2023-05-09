package elasticx

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type index struct {
	name   string
	engine *engine
	es     *elasticsearch.TypedClient
}

// newIndex creates a new Index implementation.
func newIndex(name string, engine *engine) (*index, error) {
	if name == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if engine == nil {
		return nil, fmt.Errorf("engine is nil")
	}
	return &index{
		name:   name,
		engine: engine,
		es:     engine.es,
	}, nil
}

// relPath creates the relative path to this index (`.clinia-engine~<engine-name>~<name>`)
func (i *index) relPath() string {
	escapedName := pathEscape(i.name)
	return join(i.engine.relPath(), escapedName)
}

// Name returns the name of the index.
func (i *index) Name() string {
	return i.name
}

// Engine returns the engine containing the index.
func (i *index) Engine() Engine {
	return i.engine
}

// Remove removes the entire index.
// If the view does not exists, a NotFoundError us returned.
func (i *index) Remove(ctx context.Context) error {
	res, err := esapi.IndicesDeleteRequest{
		Index: []string{i.relPath()},
	}.Do(ctx, i.es)

	if err != nil {
		return err
	}

	if res.IsError() {
		return withElasticError(res)
	}

	return nil
}
