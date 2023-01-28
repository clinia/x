package stringsx

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSingleLine(t *testing.T) {

	for k, tc := range []struct {
		value    string
		expected string
	}{
		{
			value:    ``,
			expected: "",
		},
		{
			value: `{
				"query": {
					"bool": {
						"must": [
							{
								"term": {
									"data.key": "name"
								}
							}
						]
					}
				}
			}`,
			expected: `{"query": {"bool": {"must": [{"term": {"data.key": "name"}}]}}}`,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			assert.Equal(t, tc.expected, SingleLine(tc.value))
		})
	}
}
