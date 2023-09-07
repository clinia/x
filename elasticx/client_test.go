package elasticx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	t.Run("should return an engine", func(t *testing.T) {
		name := "test-client-get-engine"
		e, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)
		assert.NotNil(t, e)

		engine, err := client.Engine(ctx, name)
		assert.NoError(t, err)

		assert.NotEmpty(t, engine)
		assert.Equal(t, engine.Name(), name)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return exists", func(t *testing.T) {
		name := "test-client-engine-exist"

		engine, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		assert.Equal(t, engine.Name(), name)

		exists, err := client.EngineExists(ctx, name)
		assert.NoError(t, err)

		assert.True(t, exists)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return all engines", func(t *testing.T) {
		// Setup
		names := []string{
			"test-client-engines-1",
			"test-client-engines-2",
			"test-client-engines-3",
		}
		for _, name := range names {
			engine, err := client.CreateEngine(ctx, name)
			assert.NotNil(t, engine)
			assert.NoError(t, err)
		}

		// Act
		engines, err := client.Engines(ctx)
		assert.NoError(t, err)

		engineNames := []string{}
		for _, engine := range engines {
			engineNames = append(engineNames, engine.Name())
		}

		assert.ElementsMatch(t, names, engineNames)

		for _, engine := range engines {
			err := engine.Remove(ctx)
			assert.NoError(t, err)
		}
	})
}

func TestClientCreateEngine(t *testing.T) {
	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	engines := []string{}

	t.Run("should create an engine", func(t *testing.T) {
		// Act
		name := "test-client-create-engine"
		engine, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		assert.Equal(t, engine.Name(), name)

		engines = append(engines, name)
	})

	t.Run("should return already exists error", func(t *testing.T) {
		// Prepare
		name := "test-client-create-engine-exists"
		_, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		// Act
		engine, err := client.CreateEngine(ctx, name)
		assert.Nil(t, engine)
		assert.EqualError(t, err, "[ALREADY_EXISTS] engine 'test-client-create-engine-exists' already exists")

		engines = append(engines, name)
	})

	t.Cleanup(func() {
		for _, name := range engines {
			f.cleanEngine(t, name)
		}
	})
}
