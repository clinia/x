// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package assertx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEqualAsJSONExcept(t *testing.T) {
	a := map[string]interface{}{"foo": "bar", "baz": "bar", "bar": "baz"}
	b := map[string]interface{}{"foo": "bar", "baz": "bar", "bar": "not-baz"}

	val := EqualAsJSONExcept(t, a, b, []string{"bar"})
	require.True(t, val)
}

func TestTimeDifferenceLess(t *testing.T) {
	val := TimeDifferenceLess(t, time.Now(), time.Now().Add(time.Second), 2)
	require.True(t, val)
}
