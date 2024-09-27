package logrusx

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSensitiveValues(t *testing.T) {
	t.Run("should leak sensitive values when explicitly set", func(t *testing.T) {
		var buf bytes.Buffer
		redacted := "REDACTED"
		l := New("foo", "bar", LeakSensitive(), WithSensitiveHeaders("x-my-api-key"), RedactionText(redacted))
		l.Entry.Logger.SetOutput(&buf)

		assert.True(t, l.LeakSensitiveData(), "should leak sensitive data when explicitly set")

		// Mock a request with a sensitive header
		req, err := http.NewRequest("GET", "https://example.com", nil)
		assert.NoError(t, err)
		sensitiveValue := "sensitive-value"
		req.Header.Set("x-my-api-key", sensitiveValue)
		req.Header.Set("authorization", sensitiveValue)
		req.Header.Set("cookie", sensitiveValue)
		req.Header.Set("set-cookie", sensitiveValue)

		l.WithRequest(req).Info("test")
		output := buf.String()
		assert.Equal(t, strings.Count(output, sensitiveValue), 4, "all sensitive headers should be shown since we are leaking sensitive values")
		assert.NotContains(t, output, redacted, "should not redact sensitive values when explicitly set to leak")
	})

	t.Run("should not leak sensitive values", func(t *testing.T) {
		var buf bytes.Buffer
		redacted := "REDACTED"
		l := New("foo", "bar", WithSensitiveHeaders("x-my-api-key"), RedactionText(redacted))
		l.Entry.Logger.SetOutput(&buf)

		assert.False(t, l.LeakSensitiveData(), "should not leak sensitive data by default")

		// Mock a request with a sensitive header
		req, err := http.NewRequest("GET", "https://example.com", nil)
		assert.NoError(t, err)
		sensitiveValue := "sensitive-value"
		req.Header.Set("x-my-api-key", sensitiveValue)
		req.Header.Set("authorization", sensitiveValue)
		req.Header.Set("cookie", sensitiveValue)
		req.Header.Set("set-cookie", sensitiveValue)

		l.WithRequest(req).Info("test")
		output := buf.String()
		assert.Equal(t, strings.Count(output, redacted), 4, "all sensitive headers should be redacted")
		assert.NotContains(t, output, sensitiveValue, "should not show sensitive values")

		req.Header.Set("not-yet-sensitive", sensitiveValue)
		buf.Reset()

		l.WithRequest(req).Info("test")
		output = buf.String()
		assert.Equal(t, strings.Count(output, redacted), 4, "all so far sensitive headers should be redacted")
		assert.Equal(t, strings.Count(output, sensitiveValue), 1, "not-yet-sensitive should not be sensitive yet")

		buf.Reset()
		l = l.WithSensitiveHeaders("not-yet-sensitive")
		l.WithRequest(req).Info("test")
		output = buf.String()
		assert.Equal(t, strings.Count(output, redacted), 5, "all so far sensitive headers should be redacted")
		assert.NotContains(t, output, sensitiveValue, "should not show sensitive values")
	})
}
