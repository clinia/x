package httpx

import (
	"net/http"
	"testing"
)

func TestResponse_Validate(t *testing.T) {
	testCases := []struct {
		testCaseName string
		input        *Response
		errExpected  bool
	}{
		{
			testCaseName: "All required values are present",
			input: &Response{
				StatusCode: http.StatusOK,
			},
			errExpected: false,
		},
		{
			testCaseName: "Status code not present",
			input:        &Response{},
			errExpected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testCaseName, func(t *testing.T) {
			err := tc.input.Validate()
			if err != nil && !tc.errExpected {
				t.Errorf("no error expected but got %v", err)
				return
			}
		})
	}
}

func TestRequest_Validate(t *testing.T) {
	testCases := []struct {
		testCaseName string
		input        *Request
		errExpected  bool
	}{
		{
			testCaseName: "All required values are present",
			input: &Request{
				Method: http.MethodGet,
				URL:    "https://example.com",
			},
			errExpected: false,
		},
		{
			testCaseName: "URL not present",
			input: &Request{
				Method: http.MethodGet,
			},
			errExpected: true,
		},
		{
			testCaseName: "Method not present",
			input: &Request{
				URL: "https://example.com",
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testCaseName, func(t *testing.T) {
			err := tc.input.Validate()
			if err != nil && !tc.errExpected {
				t.Errorf("no error expected but got %v", err)
				return
			}
		})
	}
}
