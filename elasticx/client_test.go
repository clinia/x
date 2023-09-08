package elasticx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCreateEngine(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	engines := []string{}

	t.Run("should create an engine", func(t *testing.T) {
		// Act
		name := "test-client-create-engine"
		engine, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		// Assert
		assert.Equal(t, engine.Name(), name)

		// Check that doc exists in es engines index
		doc, err := f.es.Get(enginesIndexName, name).Do(ctx)
		assert.NoError(t, err)
		assert.True(t, doc.Found)

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

func TestClientGetEngine(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	engines := []string{}

	t.Run("should return an engine", func(t *testing.T) {
		// Prepare
		name := "test-client-get-engine"
		_, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		// Act
		engine, err := client.Engine(ctx, name)
		assert.NoError(t, err)
		assert.Equal(t, name, engine.Name())

		engines = append(engines, name)
	})

	t.Run("should return not found error", func(t *testing.T) {
		// Act
		engine, err := client.Engine(ctx, "test-client-get-engine-not-found")
		assert.Nil(t, engine)
		assert.EqualError(t, err, "[NOT_FOUND] engine 'test-client-get-engine-not-found' does not exist")
	})

	t.Cleanup(func() {
		for _, name := range engines {
			f.cleanEngine(t, name)
		}
	})
}

func TestClientEngineExists(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	engines := []string{}

	t.Run("should return true", func(t *testing.T) {
		// Prepare
		name := "test-client-engine-exists"
		_, err := client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		// Act
		exists, err := client.EngineExists(ctx, name)
		assert.NoError(t, err)
		assert.True(t, exists)

		engines = append(engines, name)
	})

	t.Run("should return false", func(t *testing.T) {
		// Act
		exists, err := client.EngineExists(ctx, "test-client-engine-exists-not-found")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Cleanup(func() {
		for _, name := range engines {
			f.cleanEngine(t, name)
		}
	})
}

func TestClientEngines(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx
	client := f.client

	engines := []string{}

	t.Run("should return all engines", func(t *testing.T) {
		// Prepare
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
		enginInfos, err := client.Engines(ctx)
		assert.NoError(t, err)

		for _, name := range names {
			assert.Contains(t, enginInfos, EngineInfo{Name: name})
		}

		engines = append(engines, names...)
	})

	t.Cleanup(func() {
		for _, name := range engines {
			f.cleanEngine(t, name)
		}
	})
}
