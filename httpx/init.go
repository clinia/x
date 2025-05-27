package httpx

import (
	"crypto/tls"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type Client struct {
	httpClient *http.Client
	transport  *http.Transport
}

// GetDefaultHTTPClient returns an HTTP client with basic settings
func GetDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: httpClientDefaultTimeout,
	}
}

// NewHTTPClient returns a default HTTP client with default options
func NewHTTPClient() *Client {
	httpClient := GetDefaultHTTPClient()
	httpClient.Transport = &http.Transport{}

	return &Client{
		httpClient: httpClient,
	}
}

// NewClientWithOptions creates a configurable HTTP Client
func NewClientWithOptions(options ...Option) *Client {
	client := &Client{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}

	client.httpClient = &http.Client{}

	for _, opt := range options {
		opt(client)
	}

	client.httpClient.Transport = client.transport

	return client
}
