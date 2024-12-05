package logrusx

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type TestKeyType string

var testKey TestKeyType

func TestFire(t *testing.T) {
	t.Run("should update entry data with requestId", func(t *testing.T) {
		e := logrus.Entry{}
		e.Data = make(logrus.Fields, 6)
		e.Context = context.WithValue(context.Background(), testKey, "rid")
		assert.Nil(t, e.Data["RequestId"])
		err := NewRequestIdHook(testKey, "RequestId").Fire(&e)
		assert.NoError(t, err)
		assert.Equal(t, "rid", e.Data["RequestId"])
	})
	t.Run("should do nothing on empty context", func(t *testing.T) {
		e := logrus.Entry{}
		e.Data = make(logrus.Fields, 6)
		e.Context = context.Background()
		assert.Nil(t, e.Data["RequestId"])
		err := NewRequestIdHook(testKey, "RequestId").Fire(&e)
		assert.NoError(t, err)
		assert.Nil(t, e.Data["RequestId"])
	})
	t.Run("should do nothing on nil Context", func(t *testing.T) {
		e := logrus.Entry{}
		e.Data = make(logrus.Fields, 6)
		e.Context = nil
		assert.Nil(t, e.Data["RequestId"])
		err := NewRequestIdHook(testKey, "RequestId").Fire(&e)
		assert.NoError(t, err)
		assert.Nil(t, e.Data["RequestId"])
	})
	t.Run("should do nothing on nil Data", func(t *testing.T) {
		e := logrus.Entry{}
		e.Data = nil
		e.Context = context.WithValue(context.Background(), testKey, "rid")
		assert.Nil(t, e.Data["RequestId"])
		err := NewRequestIdHook(testKey, "RequestId").Fire(&e)
		assert.NoError(t, err)
		assert.Nil(t, e.Data["RequestId"])
	})

	t.Run("should do nothing on nil Entry", func(t *testing.T) {
		err := NewRequestIdHook(testKey, "RequestId").Fire(nil)
		assert.NoError(t, err)
	})
}

func TestLevels(t *testing.T) {
	t.Run("shoul return all levels", func(t *testing.T) {
		levels := NewRequestIdHook(testKey, "RequestId").Levels()
		assert.Equal(t, logrus.AllLevels, levels)
	})
}
