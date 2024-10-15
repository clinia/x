package persistencex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListResponseMeta(t *testing.T) {
	for _, tt := range []struct {
		name     string
		page     int
		perPage  int
		total    int64
		expected ListResponseMeta
	}{
		{
			"Should generate list response meta when perPage is 0",
			0,
			0,
			int64(100),
			ListResponseMeta{
				Page:     0,
				PerPage:  0,
				NumPages: 0,
				Total:    100,
			},
		},
		{
			"Should generate list response meta when total < perPage",
			0,
			2,
			int64(1),
			ListResponseMeta{
				Page:     0,
				PerPage:  2,
				NumPages: 1,
				Total:    1,
			},
		},
		{
			"Should generate list response meta when total > perPage",
			1,
			2,
			int64(3),
			ListResponseMeta{
				Page:     1,
				PerPage:  2,
				NumPages: 2,
				Total:    3,
			},
		},
		{
			"Should generate list response meta when total == perPage",
			1,
			3,
			int64(3),
			ListResponseMeta{
				Page:     1,
				PerPage:  3,
				NumPages: 1,
				Total:    3,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			meta := NewListResponseMeta(tt.page, tt.perPage, tt.total)
			assert.Equal(t, tt.expected, meta)
		})
	}
}
