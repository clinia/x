package elasticx

import (
	"context"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/dynamicmapping"
	"github.com/stretchr/testify/assert"
)

func TestEngine(t *testing.T) {
	client, _ := createClient(t)

	ctx := context.Background()

	err := client.Init(ctx)
	if err != nil {
		t.Error(err)
	}

	t.Run("should remove an engine", func(t *testing.T) {
		ctx := context.Background()

		name := "test-create-engine"
		engine, err := client.CreateEngine(ctx, name, nil)
		assert.NoError(t, err)

		assert.Equal(t, engine.Name(), name)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})

	name := "engine-indexes"
	engine, err := client.CreateEngine(ctx, name, nil)
	if err != nil {
		t.Error(err)
	}

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
		assert.Equal(t, index.Name(), name)

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
		assert.Equal(t, index.Name(), name)

		err = index.Remove(ctx)
		assert.NoError(t, err)
	})

	t.Run("should return exists", func(t *testing.T) {
		name := "test-index-exist"

		index, err := engine.CreateIndex(ctx, name, nil)
		assert.NoError(t, err)
		assert.Equal(t, index.Name(), name)

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
		indices, err := engine.Indexes(ctx)
		assert.NoError(t, err)

		indexNames := []string{}
		for _, index := range indices {
			indexNames = append(indexNames, index.Name())
		}

		assert.ElementsMatch(t, indexNames, names)

		for _, index := range indices {
			err := index.Remove(ctx)
			assert.NoError(t, err)
		}
	})

	t.Run("should be able to execute queries", func(t *testing.T) {
		ctx := context.Background()

		name := "test-engine-queries"
		engine, err := client.CreateEngine(ctx, name, nil)
		assert.NoError(t, err)

		index, err := engine.CreateIndex(ctx, "index-1", nil)
		assert.NoError(t, err)

		res, err := engine.Queries(ctx, []MultiQuery{
			{
				Index: []string{index.Name()},
				Name:  "query-1",
				Query: `{
					"query": {
						"match_all": {}
					}
				}`,
			},
			{
				Index: []string{index.Name()},
				Name:  "query-2",
				Query: `{
					"query": {
						"match_all": {}
					},
					"from": 0
				}`,
			},
		}...)

		assert.NoError(t, err)
		assert.Len(t, res, 2)

		err = engine.Remove(ctx)
		assert.NoError(t, err)
	})
}
