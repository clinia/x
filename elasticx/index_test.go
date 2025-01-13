package elasticx

import (
	"testing"

	"github.com/clinia/x/assertx"
	"github.com/clinia/x/jsonx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestIndexUpdateMappings(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine := f.setupEngine(t, "test-index-update-mappings")

	t.Run("should update an index", func(t *testing.T) {
		name := "index-1"
		index, err := engine.CreateIndex(ctx, name, &CreateIndexOptions{
			Aliases: map[string]types.Alias{
				"test": {},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"id": types.NewKeywordProperty(),
					"p1": types.NewKeywordProperty(),
					"object": types.NestedProperty{
						Dynamic: &dynamicmapping.Strict,
						Properties: map[string]types.Property{
							"p1": types.NewKeywordProperty(),
						},
					},
				},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, index.Info().Name, name)

		// Assert the index exists via es
		esIndexName := NewIndexName(enginesIndexName, engine.Name(), name).String()
		esIndices, err := f.es.Indices.Get(esIndexName).Do(ctx)
		assert.NoError(t, err)
		assert.Len(t, esIndices, 1)

		esindex := esIndices[esIndexName]

		assertx.Equal(t, types.IndexState{
			Aliases: map[string]types.Alias{
				NewIndexName(enginesIndexName, engine.Name(), "test").String(): {},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"id": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
					"p1": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
					"object": &types.NestedProperty{
						Type:    "nested",
						Dynamic: &dynamicmapping.Strict,
						Fields:  map[string]types.Property{},
						Meta:    map[string]string{},
						Properties: map[string]types.Property{
							"p1": &types.KeywordProperty{
								Type:       "keyword",
								Fields:     map[string]types.Property{},
								Meta:       map[string]string{},
								Properties: map[string]types.Property{},
							},
						},
					},
				},
			},
		}, esindex,
			cmpopts.IgnoreFields(types.IndexState{}, "Settings"),
		)

		// update index mapping
		err = index.UpdateMappings(ctx, &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"id": types.NewKeywordProperty(),
				"object": types.NestedProperty{
					Dynamic: &dynamicmapping.Strict,
					Properties: map[string]types.Property{
						"p1":  types.NewKeywordProperty(),
						"new": types.NewKeywordProperty(),
					},
				},
				"new": types.NewKeywordProperty(),
				"newObject": types.NestedProperty{
					Dynamic: &dynamicmapping.Strict,
					Properties: map[string]types.Property{
						"p1": types.NewKeywordProperty(),
					},
				},
			},
		})

		assert.NoError(t, err)

		// Assert the index exists via es
		esIndices, err = f.es.Indices.Get(esIndexName).Do(ctx)
		assert.NoError(t, err)
		assert.Len(t, esIndices, 1)

		esindex = esIndices[esIndexName]

		assertx.Equal(t, types.IndexState{
			Aliases: map[string]types.Alias{
				NewIndexName(enginesIndexName, engine.Name(), "test").String(): {},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"id": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
					"p1": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
					"object": &types.NestedProperty{
						Type:    "nested",
						Dynamic: &dynamicmapping.Strict,
						Fields:  map[string]types.Property{},
						Meta:    map[string]string{},
						Properties: map[string]types.Property{
							"p1": &types.KeywordProperty{
								Type:       "keyword",
								Fields:     map[string]types.Property{},
								Meta:       map[string]string{},
								Properties: map[string]types.Property{},
							},
							"new": &types.KeywordProperty{
								Type:       "keyword",
								Fields:     map[string]types.Property{},
								Meta:       map[string]string{},
								Properties: map[string]types.Property{},
							},
						},
					},
					"new": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
					"newObject": &types.NestedProperty{
						Type:    "nested",
						Dynamic: &dynamicmapping.Strict,
						Fields:  map[string]types.Property{},
						Meta:    map[string]string{},
						Properties: map[string]types.Property{
							"p1": &types.KeywordProperty{
								Type:       "keyword",
								Fields:     map[string]types.Property{},
								Meta:       map[string]string{},
								Properties: map[string]types.Property{},
							},
						},
					},
				},
			},
		}, esindex,
			cmpopts.IgnoreFields(types.IndexState{}, "Settings"),
		)
	})

	t.Run("should fail to update an existing property", func(t *testing.T) {
		name := "index-2"
		index, err := engine.CreateIndex(ctx, name, &CreateIndexOptions{
			Aliases: map[string]types.Alias{
				"test": {},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"p1": types.NewKeywordProperty(),
				},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, index.Info().Name, name)

		// Assert the index exists via es
		esIndexName := NewIndexName(enginesIndexName, engine.Name(), name).String()
		esIndices, err := f.es.Indices.Get(esIndexName).Do(ctx)
		assert.NoError(t, err)
		assert.Len(t, esIndices, 1)

		esindex := esIndices[esIndexName]

		assertx.Equal(t, types.IndexState{
			Aliases: map[string]types.Alias{
				NewIndexName(enginesIndexName, engine.Name(), "test").String(): {},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"p1": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
				},
			},
		}, esindex,
			cmpopts.IgnoreFields(types.IndexState{}, "Settings"),
		)

		// update index mapping
		err = index.UpdateMappings(ctx, &types.TypeMapping{
			Dynamic: &dynamicmapping.Strict,
			Properties: map[string]types.Property{
				"p1": types.NestedProperty{
					Dynamic: &dynamicmapping.Strict,
					Properties: map[string]types.Property{
						"new": types.NewKeywordProperty(),
					},
				},
			},
		})

		assert.EqualError(t, err, "status: 400, failed: [illegal_argument_exception], reason: can't merge a non-nested mapping [p1] with a nested mapping")

		// Assert the index exists via es
		esIndices, err = f.es.Indices.Get(esIndexName).Do(ctx)
		assert.NoError(t, err)
		assert.Len(t, esIndices, 1)

		esindex = esIndices[esIndexName]

		assertx.Equal(t, types.IndexState{
			Aliases: map[string]types.Alias{
				NewIndexName(enginesIndexName, engine.Name(), "test").String(): {},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"p1": &types.KeywordProperty{
						Type:       "keyword",
						Fields:     map[string]types.Property{},
						Meta:       map[string]string{},
						Properties: map[string]types.Property{},
					},
				},
			},
		}, esindex,
			cmpopts.IgnoreFields(types.IndexState{}, "Settings"),
		)
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
		assertx.Equal(t, &UpsertResponse[DocumentMeta]{
			Result: UpsertResultCreated,
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
		assertx.Equal(t, &UpsertResponse[DocumentMeta]{
			Result: UpsertResultUpdated,
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

func TestIndexQueryDeleteDocuments(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine, err := f.client.CreateEngine(ctx, "test-index-query-delete-documents")
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

	t.Run("should delete all foo key documents", func(t *testing.T) {
		// Prepare
		meta1, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		}, WithRefresh(refresh.True))
		assert.NoError(t, err)

		meta2, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "barbar",
		}, WithRefresh(refresh.True))
		assert.NoError(t, err)

		meta3, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		}, WithRefresh(refresh.True))
		assert.NoError(t, err)
		query := &types.Query{
			Term: map[string]types.TermQuery{
				"foo": {Value: "bar"},
			},
		}

		// Act
		result, err := index.DeleteDocumentsByQuery(ctx, query,
			WithWaitForCompletion(true),
			WithRefresh(refresh.True),
		)
		assert.NoError(t, err)
		assert.Equal(t, 2, int(result.DeleteCount))

		// Assert the documents do or do not exist via es
		exists, err := f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta1.ID).
			Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
		exists, err = f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta2.ID).
			Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)
		exists, err = f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta3.ID).
			Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("should fail on empty query", func(t *testing.T) {
		_, err := index.DeleteDocumentsByQuery(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("should delete all foo key documents async", func(t *testing.T) {
		// Prepare
		meta1, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		}, WithRefresh(refresh.True))
		assert.NoError(t, err)

		meta2, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "barbar",
		}, WithRefresh(refresh.True))
		assert.NoError(t, err)

		meta3, err := index.CreateDocument(ctx, map[string]interface{}{
			"foo": "bar",
		}, WithRefresh(refresh.True))
		assert.NoError(t, err)
		query := &types.Query{
			Term: map[string]types.TermQuery{
				"foo": {Value: "bar"},
			},
		}

		// Act
		result, err := index.DeleteDocumentsByQuery(ctx, query,
			WithWaitForCompletion(false),
		)
		assert.NoError(t, err)

		taskId := (*result.TaskId).(string)

		_, err = f.es.Tasks.Get(taskId).WaitForCompletion(true).
			Do(ctx)
		assert.NoError(t, err)

		// Assert the documents do or do not exist via es
		exists, err := f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta1.ID).
			Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
		exists, err = f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta2.ID).
			Do(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)
		exists, err = f.es.Exists(NewIndexName(enginesIndexName, engine.Name(), index.Info().Name).String(), meta3.ID).
			Do(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}
