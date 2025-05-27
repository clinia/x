package httpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetDefaultHTTPClient(t *testing.T) {
	client := GetDefaultHTTPClient()

	assert.Equal(t, httpClientDefaultTimeout, client.Timeout)
}

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()

	assert.Equal(t, httpClientDefaultTimeout, client.httpClient.Timeout)
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClientWithOptions(WithTimeout(30*time.Second), WithSkipTLSVerification())

	assert.Equal(t, true, client.transport.TLSClientConfig.InsecureSkipVerify)
	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}
