package httpx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HTTPClientTestSuite struct {
	suite.Suite
	testServer *httptest.Server
	client     *Client
}

func TestHTTPClientTestSuite(t *testing.T) {
	suite.Run(t, new(HTTPClientTestSuite))
}

func (s *HTTPClientTestSuite) SetupSuite() {
	s.testServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, values := range r.URL.Query() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		fmt.Fprint(w, "test server")
	}))

	s.client = NewHTTPClient()
}

func (s *HTTPClientTestSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *HTTPClientTestSuite) TestMakeHTTPRequest_InvalidRequest() {
	ctx := context.Background()

	_, err := s.client.MakeHTTPRequest(ctx, &Request{
		URL: s.testServer.URL,
	})

	s.Assert().Error(err)
}

func (s *HTTPClientTestSuite) TestMakeHTTPRequest_SuccessfulHTTPRequest() {
	ctx := context.Background()

	request := &Request{
		Method: http.MethodGet,
		URL:    s.testServer.URL,
		QueryParameters: map[string][]string{
			"test": {"test1", "test2"},
		},
	}

	response, err := s.client.MakeHTTPRequest(ctx, request)
	if err != nil {
		s.FailNow("unable to make http request to the test server: ", err)
		return
	}

	s.Assert().Equal(http.StatusOK, response.StatusCode)
	s.Assert().Equal("test server", string(response.Body))
	s.Assert().Equal(len(response.Headers.Get("test")), len(request.QueryParameters.Get("test")))
}
