package testx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *http.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Handler.ServeHTTP(rr, req)

	return rr
}

func unmarsharlBody[T any](res *httptest.ResponseRecorder) T {
	body, _ := io.ReadAll(res.Body)
	var data T
	_ = json.Unmarshal(body, &data)
	return data
}

type requestOption func(req *http.Request)

func PutJson[T any](s *http.Server, url string, jsonStr string, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(strings.ReplaceAll(jsonStr, "\n", ""))))
	for _, opt := range opts {
		opt(req)
	}
	req.Header.Set("Content-Type", "application/json")

	res := executeRequest(req, s)

	return res, unmarsharlBody[T](res)
}

func PostJson[T any](s *http.Server, url string, jsonStr string, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(strings.ReplaceAll(jsonStr, "\n", ""))))
	for _, opt := range opts {
		opt(req)
	}
	req.Header.Set("Content-Type", "application/json")

	res := executeRequest(req, s)
	return res, unmarsharlBody[T](res)
}

func PatchJson[T any](s *http.Server, url string, jsonStr string, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(strings.ReplaceAll(jsonStr, "\n", ""))))
	for _, opt := range opts {
		opt(req)
	}
	req.Header.Set("Content-Type", "application/json")

	res := executeRequest(req, s)

	return res, unmarsharlBody[T](res)
}

func GetJson[T any](s *http.Server, url string, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	req, _ := http.NewRequest("GET", url, nil)
	for _, opt := range opts {
		opt(req)
	}
	for _, opt := range opts {
		opt(req)
	}

	req.Header.Set("Content-Type", "application/json")

	res := executeRequest(req, s)
	return res, unmarsharlBody[T](res)
}

func DeleteJson[T any](s *http.Server, url string, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	req, _ := http.NewRequest("DELETE", url, nil)
	for _, opt := range opts {
		opt(req)
	}

	res := executeRequest(req, s)
	return res, unmarsharlBody[T](res)
}

func Delete(s *http.Server, url string, opts ...requestOption) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("DELETE", url, nil)
	for _, opt := range opts {
		opt(req)
	}

	return executeRequest(req, s)
}

func WithHeader(key, value string) requestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}
