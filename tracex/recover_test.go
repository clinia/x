package tracex

import (
	"errors"
	"io"
	"testing"

	"github.com/clinia/x/logrusx"
	"github.com/clinia/x/testx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverWithStackTrace(t *testing.T) {
	t.Run("should recover from panic and log stack trace", func(t *testing.T) {
		l := logrusx.New("test", "")
		buf := testx.NewConcurrentBuffer(t)
		l.Entry.Logger.SetOutput(buf)
		assertCount := 0
		t.Cleanup(func() {
			require.GreaterOrEqual(t, assertCount, 3)
		})
		defer func() {
			assert.Contains(t, buf.String(), "panic at the disco")
			assertCount++
			assert.Contains(t, buf.String(), "test panic")
			assertCount++
			assert.Contains(t, buf.String(), "tracex/recover.go")
			assertCount++
		}()

		defer RecoverWithStackTracef(l, "panic at the disco")

		panic("test panic")
	})

	t.Run("should recover from panic and log stack trace with error panic", func(t *testing.T) {
		l := logrusx.New("test", "")
		buf := testx.NewConcurrentBuffer(t)
		l.Entry.Logger.SetOutput(buf)
		assertCount := 0
		t.Cleanup(func() {
			require.GreaterOrEqual(t, assertCount, 1)
		})
		defer func() {
			assert.Contains(t, buf.String(), "test panic")
			assertCount++
		}()

		defer RecoverWithStackTracef(l, "")

		panic(errors.New("test panic"))
	})

	t.Run("should recover from panic and log stack trace with random types or nil", func(t *testing.T) {
		l := logrusx.New("test", "")
		buf := testx.NewConcurrentBuffer(t)
		l.Entry.Logger.SetOutput(buf)
		assertCount := 0
		t.Cleanup(func() {
			require.GreaterOrEqual(t, assertCount, 3)
		})

		testPanic := func(v interface{}, expectedStr string) {
			defer func() {
				assert.Contains(t, buf.String(), expectedStr)
				assertCount++
			}()

			defer RecoverWithStackTracef(l, "")

			panic(v)
		}

		testPanic([]int{}, "unknown panic")
		testPanic(123, "unknown panic")
		testPanic(struct{}{}, "unknown panic")
	})

	t.Run("should not allow panics in the recoverer", func(t *testing.T) {
		l := logrusx.New("test", "")
		w := &writerThatpanics{}
		l.Entry.Logger.SetOutput(w)
		assertCount := 0
		t.Cleanup(func() {
			require.GreaterOrEqual(t, assertCount, 1)
		})

		defer func() {
			assertCount++
			assert.Equal(t, w.panics, 1)
		}()

		defer RecoverWithStackTracef(l, "panic while handling messages")

		panic("test panic")
	})

	t.Run("should not log if logger is nil", func(t *testing.T) {
		defer RecoverWithStackTracef(nil, "panic while handling messages")

		panic("test panic")

		// No assertion needed, just ensure no panic occurs
	})
}

func TestGetStackTrace(t *testing.T) {
	t.Run("should return stack trace", func(t *testing.T) {
		stackTrace := GetStackTrace()
		assert.Contains(t, stackTrace, "tracex/recover_test.go")
	})

	t.Run("should return empty stack trace if no panic", func(t *testing.T) {
		stackTrace := GetStackTrace()
		assert.NotEmpty(t, stackTrace)
	})
}

type writerThatpanics struct {
	panics int
}

// Write implements io.Writer.
func (w *writerThatpanics) Write(p []byte) (n int, err error) {
	w.panics++
	panic("expected")
}

var _ io.Writer = (*writerThatpanics)(nil)
