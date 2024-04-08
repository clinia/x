package arangox

import (
	arangoDriver "github.com/arangodb/go-driver"
)

type CollectionParam struct {
	Name    string
	Options *arangoDriver.CreateCollectionOptions
	Indexes []IndexParam
}

type ViewParam struct {
	Name    string
	Options *arangoDriver.ArangoSearchViewProperties
}

type IndexParam struct {
	Type                   IndexType
	Fields                 []string
	ExpireAfter            int
	InvertedIndexOptions   *arangoDriver.InvertedIndexOptions
	ZKDIndexOptions        *arangoDriver.EnsureZKDIndexOptions
	SkipListIndexOptions   *arangoDriver.EnsureSkipListIndexOptions
	PersistentIndexOptions *arangoDriver.EnsurePersistentIndexOptions
	HashIndexOptions       *arangoDriver.EnsureHashIndexOptions
	GeoIndexOptions        *arangoDriver.EnsureGeoIndexOptions
	TTLIndexOptions        *arangoDriver.EnsureTTLIndexOptions
}

type EdgeCollectionParam struct {
	Name        string
	Constraints arangoDriver.VertexConstraints
}

type GraphParam struct {
	Name              string
	Options           *arangoDriver.CreateGraphOptions
	VertexCollections []string
	EdgeCollections   []EdgeCollectionParam
}
