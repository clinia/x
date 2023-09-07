package elasticx

import (
	"context"
	"testing"

	"github.com/clinia/x/pointerx"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/stretchr/testify/assert"
)

func TestEngine(t *testing.T) {
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
			exists, err := f.es.Indices.Exists(index).Do(ctx)
			assert.NoError(t, err)
			assert.True(t, exists)
		}

		// Remove the engine
		err = engine.Remove(ctx)
		assert.NoError(t, err)

		// Assert the indexes do not exist via es
		for _, index := range indexNames {
			exists, err := f.es.Indices.Exists(index).Do(ctx)
			assert.NoError(t, err)
			assert.False(t, exists)
		}

		// Assert the engine does not exist via es
		res, err := f.es.Get(enginesIndexName, name).Do(ctx)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	name := "engine-indexes"
	engine, err := f.client.CreateEngine(ctx, name)
	assert.NoError(t, err)

	t.Run("should create an index", func(t *testing.T) {
		name := "test-create-index"
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

		err = index.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return an index", func(t *testing.T) {
		name := "test-get-index"
		i, err := engine.CreateIndex(ctx, name, nil)
		assert.NoError(t, err)
		assert.NotNil(t, i)

		index, err := engine.Index(ctx, name)
		assert.NoError(t, err)

		assert.NotEmpty(t, index)
		assert.Equal(t, index.Info().Name, name)

		err = index.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return exists", func(t *testing.T) {
		name := "test-index-exist"

		index, err := engine.CreateIndex(ctx, name, nil)
		assert.NoError(t, err)
		assert.Equal(t, index.Info().Name, name)

		exists, err := engine.IndexExists(ctx, name)
		assert.NoError(t, err)
		assert.True(t, exists)

		err = index.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return all indexes", func(t *testing.T) {
		// Setup
		names := []string{
			"test-indexes-1",
			"test-indexes-2",
			"test-indexes-3",
		}
		for _, name := range names {
			index, err := engine.CreateIndex(ctx, name, nil)
			assert.NotNil(t, index)
			assert.NoError(t, err)
		}

		// Act
		indexInfos, err := engine.Indexes(ctx)
		assert.NoError(t, err)

		assert.ElementsMatch(t, []IndexInfo{
			{Name: "test-indexes-1"},
			{Name: "test-indexes-2"},
			{Name: "test-indexes-3"},
		}, indexInfos)

		for _, indexInfo := range indexInfos {
			index, err := engine.Index(ctx, indexInfo.Name)
			assert.NoError(t, err)

			err = index.Remove(ctx)
			assert.NoError(t, err)
		}
	})

	t.Run("should be able to execute queries", func(t *testing.T) {
		ctx := context.Background()

		name := "test-engine-queries"
		engine, err := f.client.CreateEngine(ctx, name)
		assert.NoError(t, err)

		index, err := engine.CreateIndex(ctx, "index-1", nil)
		assert.NoError(t, err)

		res, err := engine.Queries(ctx, []MultiQuery{
			{
				IndexName: index.Info().Name,
				Request: types.MultisearchBody{
					Query: &types.Query{
						MatchAll: &types.MatchAllQuery{},
					},
				},
			},
			{
				IndexName: index.Info().Name,
				Request: types.MultisearchBody{
					Query: &types.Query{
						MatchAll: &types.MatchAllQuery{},
					},
					From: pointerx.Ptr(0),
				},
			},
		}...)

		assert.NoError(t, err)
		assert.Len(t, res, 2)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})
}
