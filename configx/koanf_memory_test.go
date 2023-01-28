// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package configx

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/clinia/x/assertx"
	"github.com/stretchr/testify/require"
)

func TestKoanfMemory(t *testing.T) {
	doc := []byte(`{
  "foo": {
    "bar": "baz"
  }
}`)
	kf := NewKoanfMemory(context.Background(), doc)

	actual, err := kf.Read()
	require.NoError(t, err)
	assertx.EqualAsJSON(t, json.RawMessage(doc), actual)
}
