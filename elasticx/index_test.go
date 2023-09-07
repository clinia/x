package elasticx

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestIndex(t *testing.T) {
// 	client, _ := createClient(t)

// 	ctx := context.Background()

// 	err := client.Init(ctx)
// 	assert.NoError(t, err)

// 	engine, err := client.CreateEngine(ctx, "test-index")
// 	assert.NoError(t, err)

// 	t.Run("should create a document", func(t *testing.T) {
// 		index, err := engine.CreateIndex(ctx, "create-documents", nil)
// 		assert.NoError(t, err)

// 		type property struct {
// 			Source string `json:"source"`
// 		}

// 		type doc struct {
// 			ID         string `json:"id"`
// 			Properties map[string][]property
// 		}

// 		d, err := index.CreateDocument(ctx, &doc{
// 			Properties: map[string][]property{
// 				"name": {
// 					{
// 						Source: "a",
// 					},
// 				},
// 			},
// 		})

// 		assert.NoError(t, err)
// 		assert.NotEmpty(t, d.ID)
// 	})

// 	t.Cleanup(func() {
// 		err := engine.Remove(ctx)
// 		assert.NoError(t, err)
// 	})
// }
