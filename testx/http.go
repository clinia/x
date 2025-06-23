package testx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

func GenerateMultipartBody(formData FormData) (io.Reader, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range formData.Fields {
		err := writer.WriteField(key, value)
		if err != nil {
			panic(err)
		}
	}

	// Add files
	for _, fileToUpload := range formData.Files {
		part, err := writer.CreateFormFile(fileToUpload.FieldName, fileToUpload.FileName)
		if err != nil {
			panic(err)
		}
		_, err = part.Write(fileToUpload.Content)
		if err != nil {
			panic(err)
		}
	}

	writer.Close() // Important: Close the writer to finalize the multipart body

	return body, writer
}

func MultipartRequest[T any](s *http.Server, method, url string, formData FormData, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	body, writer := GenerateMultipartBody(formData)
	req, _ := http.NewRequest(method, url, body)
	for _, opt := range opts {
		opt(req)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType()) // Set the correct Content-Type header with boundary

	res := executeRequest(req, s)
	return res, unmarsharlBody[T](res)
}

func PostMultipart[T any](s *http.Server, url string, formData FormData, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	return MultipartRequest[T](s, "POST", url, formData, opts...)
}

func PutMultipart[T any](s *http.Server, url string, formData FormData, opts ...requestOption) (*httptest.ResponseRecorder, T) {
	return MultipartRequest[T](s, "PUT", url, formData, opts...)
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

func WithContext(ctx context.Context) requestOption {
	return func(req *http.Request) {
		*req = *req.WithContext(ctx)
	}
}
