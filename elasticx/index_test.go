package elasticx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {
	client, _ := createClient(t)

	ctx := context.Background()

	err := client.Clean(ctx)
	if err != nil {
		t.Error(err)
	}

	err = client.Init(ctx)
	if err != nil {
		t.Error(err)
	}

	engine, err := client.CreateEngine(ctx, "engine-index", nil)
	if err != nil {
		t.Error(err)
	}

	index, err := engine.CreateIndex(ctx, "index-documents", nil)
	if err != nil {
		t.Error(err)
	}

	t.Run("should create a document", func(t *testing.T) {
		type property struct {
			Source string `json:"source"`
		}

		type doc struct {
			ID         string `json:"id"`
			Properties map[string][]property
		}

		d, err := index.CreateDocument(ctx, &doc{
			Properties: map[string][]property{
				"name": {
					{
						Source: "a",
					},
				},
			},
		}, nil)

		assert.NoError(t, err)
		assert.NotEmpty(t, d.ID)
	})
}
