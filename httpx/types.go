package httpx

import (
	"net/http"
	"net/url"
	"time"
)

const httpClientDefaultTimeout = 60 * time.Second

// Request is the input parameters that will need to be sent with an HTTP request
type Request struct {
	Method          string `validate:"required"`
	URL             string `validate:"required"`
	Body            any
	Headers         http.Header
	QueryParameters url.Values
}

// Validate validates if the struct contains the required entities or not
func (r *Request) Validate() error {
	return validate.Struct(r)
}

// Response struct will contain the entities returned with the HTTP response
type Response struct {
	StatusCode int `validate:"required"`
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
}

// Validate validates if the struct contains the required entities or not
func (r *Response) Validate() error {
	return validate.Struct(r)
}
