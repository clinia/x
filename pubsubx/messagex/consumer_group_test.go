package messagex_test

import (
	"testing"

	"github.com/clinia/x/errorx"
	"github.com/clinia/x/pubsubx/messagex"
	"github.com/stretchr/testify/assert"
)

func TestConsumerGroup(t *testing.T) {
	t.Run("ConsumerGroup", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			group         messagex.ConsumerGroup
			scope         string
			expectedGroup string
		}{
			{
				name:          "with scope",
				group:         "group",
				scope:         "scope",
				expectedGroup: "scope.group",
			},
			{
				name:          "without scope",
				group:         "group",
				scope:         "",
				expectedGroup: "group",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expectedGroup, tc.group.ConsumerGroup(tc.scope))
			})
		}
	})
	t.Run("ExtractScopeFromConsumerGroup", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			group         string
			expectedScope string
			expectedGroup messagex.ConsumerGroup
			expectedErr   error
		}{
			{
				name:          "valid group with scope",
				group:         "scope.group",
				expectedScope: "scope",
				expectedGroup: "group",
			},
			{
				name:        "group without scope",
				group:       "group",
				expectedErr: errorx.InvalidArgumentError("group 'group' does not have a valid format, expected at least scope and group name"),
			},
			{
				name:          "valid group with multiple segments",
				group:         "scope.group1.group2",
				expectedScope: "scope",
				expectedGroup: "group1.group2",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				scope, group, err := messagex.ExtractScopeFromConsumerGroup(tc.group)
				if tc.expectedErr != nil {
					assert.EqualError(t, err, tc.expectedErr.Error())
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedScope, scope)
				assert.Equal(t, tc.expectedGroup, group)
			})
		}
	})

	t.Run("RenameConsumerGroupWithScope", func(t *testing.T) {
		for _, tc := range []struct {
			name          string
			group         string
			scope         string
			expectedGroup string
			expectedErr   error
		}{
			{
				name:          "valid group with scope",
				group:         "oldScope.group",
				scope:         "newScope",
				expectedGroup: "newScope.group",
			},
			{
				name:        "group without scope",
				group:       "group",
				scope:       "",
				expectedErr: errorx.InvalidArgumentError("group 'group' does not have a valid format, expected at least scope and group name"),
			},
			{
				name:          "valid group with multiple segments",
				group:         "whateverScope.group1.group2",
				scope:         "my-new-scope",
				expectedGroup: "my-new-scope.group1.group2",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				group, err := messagex.RenameConsumerGroupWithScope(tc.group, tc.scope)
				if tc.expectedErr != nil {
					assert.EqualError(t, err, tc.expectedErr.Error())
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGroup, group)
			})
		}
	})
}
