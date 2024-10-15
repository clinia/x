package elasticx

import (
	"testing"

	"github.com/clinia/x/assertx"
	"github.com/clinia/x/jsonx"
	"github.com/clinia/x/persistencex"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/stretchr/testify/assert"
)

func TestIndexRemove(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine, err := f.client.CreateEngine(ctx, "test-index-remove")
	assert.NoError(t, err)

	t.Run("should remove an index", func(t *testing.T) {
		// Prepare
		index, err := engine.CreateIndex(ctx, "test-index-remove", nil)
		assert.NoError(t, err)

		// Remove the index
		err = index.Remove(ctx)
		assert.NoError(t, err)

		// Assert the index does not exist via es
		exists, err := f.es.Indices.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String()).Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestIndexCreateDocument(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine, err := f.client.CreateEngine(ctx, "test-index-create-document")
	assert.NoError(t, err)

	index, err := engine.CreateIndex(ctx, "index-1", &CreateIndexOptions{
		Mappings: &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"foo": types.NewKeywordProperty(),
			},
		},
	})
	assert.NoError(t, err)

	t.Run("should create a document", func(t *testing.T) {
		meta, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, meta.ID)
		assert.Equal(t, "index-1", meta.Index)
		assert.Equal(t, int64(1), meta.Version)

		// Assert the document exists via es
		exists, err := f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta.ID).Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return failed precondition error when creating a document non valid mapping", func(t *testing.T) {
		meta, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
		})
		assert.Nil(t, meta)
		assert.EqualError(t, err, "[FAILED_PRECONDITION] document contains fields that are not allowed by the index mapping: [1:8] mapping set to strict, dynamic introduction of [baz] within [_doc] is not allowed")
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestIndexUpsertDocument(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine, err := f.client.CreateEngine(ctx, "test-index-replace-document")
	assert.NoError(t, err)

	index, err := engine.CreateIndex(ctx, "index-1", &CreateIndexOptions{
		Mappings: &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"foo": types.NewKeywordProperty(),
			},
		},
	})
	assert.NoError(t, err)

	t.Run("should create a non existing document with specified id", func(t *testing.T) {
		result, err := index.UpsertDocument(ctx, "1", map[string]interface{}{
			"foo": "bar",
		})
		assert.NoError(t, err)
		assertx.Equal(t, &persistencex.UpsertResponse[DocumentMeta]{
			Result: persistencex.UpsertResultCreated,
			Meta: DocumentMeta{
				ID:      "1",
				Index:   "index-1",
				Version: 1,
			},
		}, result)

		// Assert the document exists via es
		exists, err := f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), result.Meta.ID).Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should replace a document", func(t *testing.T) {
		// Prepare
		meta, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		})
		assert.NoError(t, err)

		// Act
		result, err := index.UpsertDocument(ctx, meta.ID, map[string]interface{}{
			"foo": "baz",
		})
		assert.NoError(t, err)

		// Assert
		assertx.Equal(t, &persistencex.UpsertResponse[DocumentMeta]{
			Result: persistencex.UpsertResultUpdated,
			Meta: DocumentMeta{
				ID:      meta.ID,
				Index:   "index-1",
				Version: 2,
			},
		}, result)

		// Assert the document exists via es
		doc, err := f.es.Get(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta.ID).Do(ctx)
		assert.NoError(t, err)
		assert.Equal(t, jsonx.RawMessage(`{
			"foo": "baz"
		}`), doc.Source_)
	})

	t.Run("should return failed precondition error when replacing a document with non valid mapping", func(t *testing.T) {
		// Prepare
		meta, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		})
		assert.NoError(t, err)

		// Act
		response, err := index.UpsertDocument(ctx, meta.ID, map[string]interface{}{
			"foo": "baz",
			"baz": "qux",
		})
		assert.Nil(t, response)
		assert.EqualError(t, err, "[FAILED_PRECONDITION] document contains fields that are not allowed by the index mapping: [1:8] mapping set to strict, dynamic introduction of [baz] within [_doc] is not allowed")
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestIndexDeleteDocument(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine, err := f.client.CreateEngine(ctx, "test-index-delete-document")
	assert.NoError(t, err)

	index, err := engine.CreateIndex(ctx, "index-1", &CreateIndexOptions{
		Mappings: &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"foo": types.NewKeywordProperty(),
			},
		},
	})
	assert.NoError(t, err)

	t.Run("should delete a document", func(t *testing.T) {
		// Prepare
		meta, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		})
		assert.NoError(t, err)

		// Act
		err = index.DeleteDocument(ctx, meta.ID)
		assert.NoError(t, err)

		// Assert the document does not exist via es
		exists, err := f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta.ID).Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should return not found error when deleting a non existing document", func(t *testing.T) {
		err := index.DeleteDocument(ctx, "1")
		assert.EqualError(t, err, "[NOT_FOUND] document with key '1' does not exist")
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestIndexReadDocument(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine, err := f.client.CreateEngine(ctx, "test-index-read-document")
	assert.NoError(t, err)

	index, err := engine.CreateIndex(ctx, "index-1", &CreateIndexOptions{
		Mappings: &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"foo": types.NewKeywordProperty(),
			},
		},
	})
	assert.NoError(t, err)

	t.Run("should return not found error when reading a non existing document", func(t *testing.T) {
		_, err := index.ReadDocument(ctx, "unknown", nil)
		assert.EqualError(t, err, "[NOT_FOUND] document with key 'unknown' does not exist")
	})

	t.Run("should read a document", func(t *testing.T) {
		// Prepare
		meta, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		})
		assert.NoError(t, err)

		// Act
		var document map[string]interface{}
		meta, err = index.ReadDocument(ctx, meta.ID, &document)
		assert.NoError(t, err)

		// Assert
		assertx.Equal(t, &DocumentMeta{
			ID:      meta.ID,
			Index:   "index-1",
			Version: 1,
		}, meta)
		assert.Equal(t, map[string]interface{}{
			"foo": "bar",
		}, document)
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}
