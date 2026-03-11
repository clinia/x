package elasticx

import (
	"testing"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pointerx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLazyIndex(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	eng := f.setupEngine(t, "test-new-lazy-index")
	t.Cleanup(func() {
		assert.NoError(t, eng.Remove(ctx))
	})

	concreteEngine := eng.(*engine)

	t.Run("should return error when name is empty", func(t *testing.T) {
		_, err := newLazyIndex("", concreteEngine)
		assert.EqualError(t, err, "name is empty")
	})

	t.Run("should return error when engine is nil", func(t *testing.T) {
		_, err := newLazyIndex("some-index", nil)
		assert.EqualError(t, err, "engine is nil")
	})

	t.Run("should return a lazy index when name and engine are valid", func(t *testing.T) {
		li, err := newLazyIndex("some-index", concreteEngine)
		require.NoError(t, err)
		assert.Equal(t, "some-index", li.Info().Name)
	})
}

func TestLazyIndexInfoAndEngine(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	eng := f.setupEngine(t, "test-lazy-index-info")
	t.Cleanup(func() {
		assert.NoError(t, eng.Remove(ctx))
	})

	// Info and Engine are pure delegations to the wrapped index – they work
	// regardless of whether the index exists in Elasticsearch.
	index, err := eng.IndexLazy(ctx, "non-existent-index")
	require.NoError(t, err)

	t.Run("Info returns correct name without ES call", func(t *testing.T) {
		assert.Equal(t, "non-existent-index", index.Info().Name)
	})

	t.Run("Engine returns the owning engine", func(t *testing.T) {
		assert.Equal(t, eng, index.Engine())
	})
}

// TestLazyIndexCheckIndexExistsCaching is the most important test for lazyIndex.
// It verifies that checkIndexExists caches the result so that Elasticsearch is
// only queried once per lazyIndex lifetime, even if the real index state changes
// afterwards.
func TestLazyIndexCheckIndexExistsCaching(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	eng := f.setupEngine(t, "test-lazy-index-caching")
	t.Cleanup(func() {
		assert.NoError(t, eng.Remove(ctx))
	})

	t.Run("should cache exists=true and not re-check after index is deleted", func(t *testing.T) {
		// Create the real index and a document so the existence-checking methods
		// have something to work with.
		_, err := eng.CreateIndex(ctx, "cache-true-index", nil)
		require.NoError(t, err)

		lazy, err := eng.IndexLazy(ctx, "cache-true-index")
		require.NoError(t, err)

		// First call: hits ES, caches exists=true.
		exists, err := lazy.DocumentExists(ctx, "missing-doc")
		require.NoError(t, err)
		assert.False(t, exists)

		// Delete the actual ES index directly (bypassing lazyIndex).
		esIndexName := NewIndexName(enginesIndexNameSegment, eng.Name(), "cache-true-index").String()
		_, err = f.es.Indices.Delete(esIndexName).Do(ctx)
		require.NoError(t, err)

		// Second call: the cache still says the index exists, so checkIndexExists
		// does NOT return a NotFoundError even though the index is gone.
		_, err = lazy.DocumentExists(ctx, "missing-doc")
		assert.NoError(t, err)
	})

	t.Run("should cache exists=false and not re-check after index is created", func(t *testing.T) {
		const name = "cache-false-index"

		lazy, err := eng.IndexLazy(ctx, name)
		require.NoError(t, err)

		// First call: hits ES, caches exists=false.
		_, err = lazy.DocumentExists(ctx, "some-id")
		require.True(t, errorx.IsNotFoundError(err))

		// Now create the real index directly (bypassing lazyIndex).
		_, err = eng.CreateIndex(ctx, name, nil)
		require.NoError(t, err)
		t.Cleanup(func() {
			esIndexName := NewIndexName(enginesIndexNameSegment, eng.Name(), name).String()
			_, _ = f.es.Indices.Delete(esIndexName).Do(ctx)
		})

		// Second call: the cache still says the index is missing, so the lazy
		// index still returns a NotFoundError even though the index now exists.
		_, err = lazy.DocumentExists(ctx, "some-id")
		assert.True(t, errorx.IsNotFoundError(err))
	})
}

func TestLazyIndexMethodsOnExistingIndex(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	eng := f.setupEngine(t, "test-lazy-index-methods")
	t.Cleanup(func() {
		assert.NoError(t, eng.Remove(ctx))
	})

	_, err := eng.CreateIndex(ctx, "methods-index", nil)
	require.NoError(t, err)

	esIndexName := NewIndexName(enginesIndexNameSegment, eng.Name(), "methods-index").String()

	lazy, err := eng.IndexLazy(ctx, "methods-index")
	require.NoError(t, err)

	t.Run("DocumentExists returns false for missing document", func(t *testing.T) {
		exists, err := lazy.DocumentExists(ctx, "no-such-doc")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CreateDocument creates a document", func(t *testing.T) {
		meta, err := lazy.CreateDocument(ctx, map[string]any{"field": "value"}, WithRefresh(refresh.Waitfor))
		require.NoError(t, err)
		require.NotEmpty(t, meta.ID)

		exists, err := lazy.DocumentExists(ctx, meta.ID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("UpsertDocument upserts a document", func(t *testing.T) {
		resp, err := lazy.UpsertDocument(ctx, "upsert-key", map[string]any{"field": "v1"}, WithRefresh(refresh.Waitfor))
		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("DeleteDocument deletes an existing document", func(t *testing.T) {
		// Index directly via ES so we have a known ID.
		res, err := f.es.Index(esIndexName).
			Id("delete-me").
			Document(map[string]any{"field": "value"}).
			Refresh(refresh.Waitfor).
			Do(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, res.Id_)

		err = lazy.DeleteDocument(ctx, "delete-me", WithRefresh(refresh.Waitfor))
		require.NoError(t, err)

		exists, err := lazy.DocumentExists(ctx, "delete-me")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("DeleteDocumentsByQuery deletes matching documents", func(t *testing.T) {
		_, err := f.es.Index(esIndexName).
			Id("to-bulk-delete").
			Document(map[string]any{"field": "bulk-delete-target"}).
			Refresh(refresh.Waitfor).
			Do(ctx)
		require.NoError(t, err)

		resp, err := lazy.DeleteDocumentsByQuery(ctx, &types.Query{MatchAll: &types.MatchAllQuery{}}, WithRefresh(refresh.Waitfor), WithWaitForCompletion(true))
		require.NoError(t, err)
		assert.Greater(t, resp.DeleteCount, int64(0))
	})

	t.Run("UpdateDocumentsByQuery updates matching documents", func(t *testing.T) {
		_, err := f.es.Index(esIndexName).
			Id("to-update").
			Document(map[string]any{"field": "old-value"}).
			Refresh(refresh.Waitfor).
			Do(ctx)
		require.NoError(t, err)

		_, err = lazy.UpdateDocumentsByQuery(ctx,
			&types.Query{MatchAll: &types.MatchAllQuery{}},
			&types.Script{Source: pointerx.Ptr("ctx._source.field = 'new-value'")},
			WithRefresh(refresh.Waitfor),
		)
		require.NoError(t, err)
	})
}
