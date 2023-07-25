package arangox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	arangoDriver "github.com/arangodb/go-driver"
	"github.com/clinia/x/errorx"
)

type IndexType int

const (
	IndexTypePersistent IndexType = iota
	IndexTypeGeo
	IndexTypeHash
	IndexTypeInverted
	IndexTypeTTL
	IndexTypeZKD
	IndexTypeSkipList
)

var (
	IndexType_name = map[int]string{
		0: "PERSISTENT",
		1: "GEO",
		2: "HASH",
		3: "INVERTED",
		4: "TIME_TO_LIVE",
		5: "ZKD",
		6: "SKIP_LIST",
	}
	IndexType_value = map[string]int{
		"PERSISTENT":   0,
		"GEO":          1,
		"HASH":         2,
		"INVERTED":     3,
		"TIME_TO_LIVE": 4,
		"ZKD":          5,
		"SKIP_LIST":    6,
	}
)

func (u IndexType) String() string {
	return IndexType_name[int(u)]
}

func ParseIndexType(s string) (IndexType, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	value, ok := IndexType_value[s]
	if !ok {
		return IndexType(0), errorx.NewInternalError(fmt.Errorf("%q is not a valid index type", s))
	}

	return IndexType(value), nil
}

func (u IndexType) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *IndexType) UnmarshalJSON(data []byte) (err error) {
	var indexType string
	if err := json.Unmarshal(data, &indexType); err != nil {
		return err
	}

	if *u, err = ParseIndexType(indexType); err != nil {
		return err
	}

	return nil
}

func MigrateIndexes(ctx context.Context, col arangoDriver.Collection, indexes []IndexParam) error {
	if len(indexes) == 0 {
		return nil
	}

	for _, idx := range indexes {
		var err error
		switch idx.Type {
		case IndexTypePersistent:
			_, _, err = col.EnsurePersistentIndex(ctx, idx.Fields, idx.PersistentIndexOptions)
		case IndexTypeGeo:
			_, _, err = col.EnsureGeoIndex(ctx, idx.Fields, idx.GeoIndexOptions)
		case IndexTypeHash:
			_, _, err = col.EnsureHashIndex(ctx, idx.Fields, idx.HashIndexOptions)
		case IndexTypeInverted:
			_, _, err = col.EnsureInvertedIndex(ctx, idx.InvertedIndexOptions)
		case IndexTypeSkipList:
			_, _, err = col.EnsureSkipListIndex(ctx, idx.Fields, idx.SkipListIndexOptions)
		case IndexTypeTTL:
			_, _, err = col.EnsureTTLIndex(ctx, idx.Fields[0], idx.ExpireAfter, idx.TTLIndexOptions)
		case IndexTypeZKD:
			_, _, err = col.EnsureZKDIndex(ctx, idx.Fields, idx.ZKDIndexOptions)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
