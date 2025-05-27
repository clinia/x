package httpx

import (
	"time"
)

// Option is a named func that will help set custom options to the HTTP Client
type Option func(*Client)

// WithTimeout sets a customizable timeout to the http client
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithSkipTLSVerification() Option {
	return func(c *Client) {
		c.transport.TLSClientConfig.InsecureSkipVerify = true
	}
}
