package elasticx

import (
	"testing"

	"github.com/clinia/x/assertx"
	elasticxbulk "github.com/clinia/x/elasticx/bulk"
	elasticxmsearch "github.com/clinia/x/elasticx/msearch"
	"github.com/clinia/x/jsonx"
	"github.com/clinia/x/pointerx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/msearch"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operationtype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/totalhitsrelation"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngineRemove(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	t.Run("should remove an engine", func(t *testing.T) {
		name := "test-engine-remove"
		engine, err := f.client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		assert.Equal(t, engine.Name(), name)

		indexNames := []string{"index-1", "index-2"}
		for _, index := range indexNames {
			_, err := engine.CreateIndex(ctx, index, nil)
			assert.NoError(t, err)
		}

		// Assert the indexes exist via es
		for _, index := range indexNames {
			exists, err := f.es.Indices.Exists(NewIndexName(enginesIndexNameSegment, name, index).String()).Do(ctx)
			assert.NoError(t, err)
			assert.True(t, exists)
		}

		// Remove the engine
		err = engine.Remove(ctx)
		assert.NoError(t, err)

		// Assert the indexes do not exist via es
		for _, index := range indexNames {
			exists, err := f.es.Indices.Exists(NewIndexName(enginesIndexNameSegment, name, index).String()).Do(ctx)
			assert.NoError(t, err)
			assert.False(t, exists)
		}

		// Assert the engine does not exist via es
		res, err := f.es.Get(enginesIndexNameSegment, name).Do(ctx)
		assert.NoError(t, err)
		assert.False(t, res.Found)
	})
}

func TestEngineCreateIndex(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine := f.setupEngine(t, "test-engine-create-index")

	t.Run("should create an index", func(t *testing.T) {
		name := "index-1"
		index, err := engine.CreateIndex(ctx, name, &CreateIndexOptions{
			Aliases: map[string]types.Alias{
				"test": {},
			},
			Settings: &types.IndexSettings{
				Index: &types.IndexSettings{
					NumberOfShards:   "1",
					NumberOfReplicas: "2",
				},
				Analysis: &types.IndexSettingsAnalysis{
					Analyzer: map[string]types.Analyzer{
						"my_custom_analyzer": types.CustomAnalyzer{
							Type:       "custom",
							Tokenizer:  "standard",
							CharFilter: []string{"html_strip"},
							Filter:     []string{"lowercase", "asciifolding"},
						},
					},
				},
			},
			Mappings: &types.TypeMapping{
				Dynamic: &dynamicmapping.Strict,
				Properties: map[string]types.Property{
					"id": types.NewKeywordProperty(),
				},
			},
		})

		assert.NoError(t, err)
		assert.Equal(t, index.Info().Name, name)

		// Assert the index exists via es
		esIndexName := NewIndexName(enginesIndexNameSegment, engine.Name(), name).String()
		esIndices, err := f.es.Indices.Get(esIndexName).Do(ctx)
		assert.NoError(t, err)
		assert.Len(t, esIndices, 1)

		esindex := esIndices[esIndexName]

		assertx.Equal(t, types.IndexState{
			Aliases: map[string]types.Alias{
				NewIndexName(enginesIndexNameSegment, engine.Name(), "test").String(): {},
			},
			Settings: &types.IndexSettings{
				Index: &types.IndexSettings{
					NumberOfShards:   "1",
					NumberOfReplicas: "2",
					Analysis: &types.IndexSettingsAnalysis{
						Analyzer: map[string]types.Analyzer{
							"my_custom_analyzer": &types.CustomAnalyzer{
								Type:       "custom",
								Tokenizer:  "standard",
								CharFilter: []string{"html_strip"},
								Filter:     []string{"lowercase", "asciifolding"},
							},
						},
					},
				},
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
				},
			},
		}, esindex,
			cmpopts.IgnoreFields(types.IndexSettings{}, "CreationDate", "ProvidedName", "Routing", "Uuid", "Version"),
		)
	})

	t.Run("should return already exists error", func(t *testing.T) {
		name := "index-2"
		_, err := engine.CreateIndex(ctx, name, nil)
		assert.NoError(t, err)

		_, err = engine.CreateIndex(ctx, name, nil)
		assert.EqualError(t, err, "[ALREADY_EXISTS] duplicate index with name 'index-2'")
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestEngineIndexExists(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine := f.setupEngine(t, "test-engine-index-exists")

	t.Run("should return true if index exists", func(t *testing.T) {
		name := "index-1"
		_, err := engine.CreateIndex(ctx, name, nil)
		assert.NoError(t, err)

		exists, err := engine.IndexExists(ctx, name)
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("should return false if index does not exists", func(t *testing.T) {
		exists, err := engine.IndexExists(ctx, "index-2")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestEngineGetIndex(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine := f.setupEngine(t, "test-engine-get-index")

	t.Run("should return an index", func(t *testing.T) {
		name := "index-1"
		index, err := engine.CreateIndex(ctx, name, nil)
		assert.NoError(t, err)

		res, err := engine.Index(ctx, name)
		assert.NoError(t, err)
		assert.Equal(t, index, res)
	})

	t.Run("should return not found error", func(t *testing.T) {
		_, err := engine.Index(ctx, "index-2")
		assert.EqualError(t, err, "[NOT_FOUND] index with name 'index-2' does not exist")
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestEngineIndexes(t *testing.T) {
	t.Parallel()

	f := newTestFixture(t)
	ctx := f.ctx

	engine := f.setupEngine(t, "test-engine-indexes")

	t.Run("should return all indexes", func(t *testing.T) {
		names := []string{
			"index-1",
			"index-2",
			"index-3",
		}
		for _, name := range names {
			_, err := engine.CreateIndex(ctx, name, nil)
			assert.NoError(t, err)
		}

		indexes, err := engine.Indexes(ctx)
		assert.NoError(t, err)

		assert.ElementsMatch(t, []IndexInfo{
			{Name: "index-1"},
			{Name: "index-2"},
			{Name: "index-3"},
		}, indexes)
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestEngineQuery(t *testing.T) {
	f := newTestFixture(t)
	ctx := f.ctx

	name := "test-engine-query"
	engine, err := f.client.CreateEngine(ctx, name)
	assert.NoError(t, err)
	index, err := engine.CreateIndex(ctx, "index-1", nil)
	assert.NoError(t, err)

	_, err = f.es.Index(NewIndexName(enginesIndexNameSegment, name, index.Info().Name).String()).
		Document(map[string]interface{}{
			"id":   "1",
			"name": "test",
		}).
		Refresh(refresh.Waitfor).
		Do(ctx)
	assert.NoError(t, err)

	t.Run("should be able to execute a query", func(t *testing.T) {
		res, err := engine.Search(ctx, &search.Request{
			Query: &types.Query{
				MatchAll: &types.MatchAllQuery{},
			},
		}, []string{index.Info().Name})

		assert.NoError(t, err)
		assert.Equal(t, 1, len(res.Hits.Hits))
	})

	t.Run("should fail for invalid size", func(t *testing.T) {
		_, err := engine.Search(ctx, &search.Request{
			Size: pointerx.Ptr(-1),
			Query: &types.Query{
				MatchAll: &types.MatchAllQuery{},
			},
		}, []string{index.Info().Name})

		assert.EqualError(t, err, "[INVALID_ARGUMENT] invalid search request: 'size' must be greater than or equal to 0")
	})

	t.Run("should fail for invalid from", func(t *testing.T) {
		_, err := engine.Search(ctx, &search.Request{
			From: pointerx.Ptr(-1),
			Size: pointerx.Ptr(100),
			Query: &types.Query{
				MatchAll: &types.MatchAllQuery{},
			},
		}, []string{index.Info().Name})

		assert.EqualError(t, err, "[INVALID_ARGUMENT] invalid search request: 'from' must be greater than or equal to 0")
	})

	t.Run("should fail for invalid window", func(t *testing.T) {
		_, err := engine.Search(ctx, &search.Request{
			From: pointerx.Ptr(9950),
			Size: pointerx.Ptr(100),
			Query: &types.Query{
				MatchAll: &types.MatchAllQuery{},
			},
		}, []string{index.Info().Name})

		assert.EqualError(t, err, "[INVALID_ARGUMENT] invalid search request. The maximum size of the search window is 10000")
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestEngineQueries(t *testing.T) {
	f := newTestFixture(t)
	ctx := f.ctx

	name := "test-engine-queries"
	if exists, err := f.client.EngineExists(ctx, name); err == nil && exists {
		engine, err := f.client.Engine(ctx, name)
		require.NoError(t, err)
		require.NoError(t, engine.Remove(ctx))
	}
	engine, err := f.client.CreateEngine(ctx, name)
	assert.NoError(t, err)

	t.Run("should not return an error when no queries are provided", func(t *testing.T) {
		expected := &msearch.Response{
			Took:      0,
			Responses: make([]types.MsearchResponseItem, 0),
		}

		actual, err := engine.MultiSearch(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)

		actual, err = engine.MultiSearch(ctx, []elasticxmsearch.Item{})
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)
	})

	t.Run("should be able to execute multi search", func(t *testing.T) {
		index, err := engine.CreateIndex(ctx, "index-1", &CreateIndexOptions{
			Settings: &types.IndexSettings{},
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"id":   types.NewKeywordProperty(),
					"name": &types.TextProperty{},
				},
			},
		})
		assert.NoError(t, err)

		_, err = f.es.Index(NewIndexName(enginesIndexNameSegment, name, index.Info().Name).String()).
			Document(map[string]interface{}{
				"id":   "1",
				"name": "test",
			}).
			Refresh(refresh.Waitfor).
			Do(ctx)
		assert.NoError(t, err)

		res, err := engine.MultiSearch(ctx, []elasticxmsearch.Item{
			{
				Header: types.MultisearchHeader{
					Index: []string{index.Info().Name},
				},
				Body: types.MultisearchBody{
					Query: &types.Query{
						Match: map[string]types.MatchQuery{
							"name": {
								Query:      "test",
								QueryName_: pointerx.Ptr("match-name"),
							},
						},
					},
				},
			},
			{
				Header: types.MultisearchHeader{
					Index: []string{index.Info().Name},
				},
				Body: types.MultisearchBody{
					Query: &types.Query{
						MatchAll: &types.MatchAllQuery{},
					},
					From: pointerx.Ptr(0),
				},
			},
		})

		assert.NoError(t, err)

		for i, item := range []types.MsearchResponseItem{
			&types.MultiSearchItem{
				Status:   pointerx.Ptr(200),
				TimedOut: false,
				Hits: types.HitsMetadata{
					Total: &types.TotalHits{
						Value: 1,
						Relation: totalhitsrelation.TotalHitsRelation{
							Name: "eq",
						},
					},
					Hits: []types.Hit{
						{
							Source_: jsonx.RawMessage(`{"id":"1","name":"test"}`),
							MatchedQueries: []string{
								"match-name",
							},
						},
					},
				},
			},
			&types.MultiSearchItem{
				Status:   pointerx.Ptr(200),
				TimedOut: false,
				Hits: types.HitsMetadata{
					Total: &types.TotalHits{
						Value: 1,
						Relation: totalhitsrelation.TotalHitsRelation{
							Name: "eq",
						},
					},
					Hits: []types.Hit{
						{
							Source_: jsonx.RawMessage(`{"id":"1","name":"test"}`),
						},
					},
				},
			},
		} {
			actual := res.Responses[i]
			assertx.Equal(t, item, actual,
				cmpopts.IgnoreFields(types.MultiSearchItem{}, "Took", "Shards_", "Aggregations", "Suggest", "Fields"),
				cmpopts.IgnoreFields(types.HitsMetadata{}, "MaxScore"),
				cmpopts.IgnoreFields(types.Hit{}, "Id_", "Index_", "Score_"),
			)
		}
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}

func TestEngineBulk(t *testing.T) {
	f := newTestFixture(t)
	ctx := f.ctx

	name := "test-engine-bulk"
	engine, err := f.client.CreateEngine(ctx, name)
	assert.NoError(t, err)

	t.Run("should not return an error when no operations are provided", func(t *testing.T) {
		expected := &bulk.Response{
			Took:   0,
			Errors: false,
			Items:  make([]map[operationtype.OperationType]types.ResponseItem, 0),
		}

		actual, err := engine.Bulk(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)

		actual, err = engine.Bulk(ctx, []elasticxbulk.Operation{})
		assert.NoError(t, err)
		assert.Equal(t, actual, expected)
	})

	t.Run("should be able to execute a bulk", func(t *testing.T) {
		index, err := engine.CreateIndex(ctx, "index-1", nil)
		assert.NoError(t, err)

		res, err := engine.Bulk(ctx,
			[]elasticxbulk.Operation{
				{
					IndexName:  index.Info().Name,
					Action:     elasticxbulk.ActionIndex,
					DocumentID: "1",
					Doc: map[string]interface{}{
						"id":   "1",
						"name": "test",
					},
				},
				{
					IndexName:  index.Info().Name,
					Action:     elasticxbulk.ActionIndex,
					DocumentID: "2",
					Doc: map[string]interface{}{
						"id":   "2",
						"name": "test",
					},
				},
				{
					IndexName:  index.Info().Name,
					Action:     elasticxbulk.ActionUpdate,
					DocumentID: "1",
					Doc: map[string]interface{}{
						"name": "updated test",
					},
				},
			},
			elasticxbulk.Refresh(refresh.Waitfor),
		)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(res.Items))

		// Assert the documents exist via es
		esIndexName := NewIndexName(enginesIndexNameSegment, engine.Name(), index.Info().Name).String()
		doc1, err := f.es.Get(esIndexName, "1").Do(ctx)
		assert.NoError(t, err)
		assert.Equal(t, jsonx.RawMessage(`{
			"id": "1",
			"name": "updated test"
		}`), doc1.Source_)

		doc2, err := f.es.Get(esIndexName, "2").Do(ctx)
		assert.NoError(t, err)
		assert.Equal(t, jsonx.RawMessage(`{
			"id": "2",
			"name": "test"
		}`), doc2.Source_)
	})

	t.Cleanup(func() {
		err := engine.Remove(ctx)
		assert.NoError(t, err)
	})
}
