package migrate

import (
	"context"
	"testing"

	"github.com/clinia/x/elasticx"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
)

type testFixture struct {
	ctx    context.Context
	client elasticx.Client
	es     *elasticsearch.TypedClient
}

func newTestFixture(t *testing.T) *testFixture {
	ctx := context.Background()
	config := elasticsearch.Config{}

	client, err := elasticx.NewClient(config)
	assert.NoError(t, err)

	err = client.Init(ctx)
	assert.NoError(t, err)

	es, err := elasticsearch.NewTypedClient(config)
	assert.NoError(t, err)

	return &testFixture{
		ctx:    ctx,
		client: client,
		es:     es,
	}
}

// cleanEngine deletes the engine with the given name.
func (f *testFixture) cleanEngine(t *testing.T, name string) {
	engine, err := f.client.Engine(f.ctx, name)
	assert.NoError(t, err)

	err = engine.Remove(f.ctx)
	assert.NoError(t, err)
}
