package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

func (c *Client) MakeHTTPRequest(ctx context.Context, input *Request) (*Response, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var (
		httpRequest *http.Request
		err         error
	)

	if input.Body != nil {
		requestBodyBytes, err := json.Marshal(input.Body)
		if err != nil {
			return nil, err
		}

		httpRequest, err = http.NewRequestWithContext(ctx, input.Method, input.URL, bytes.NewBuffer(requestBodyBytes))
		if err != nil {
			return nil, err
		}
	} else {
		httpRequest, err = http.NewRequestWithContext(ctx, input.Method, input.URL, nil)
		if err != nil {
			return nil, err
		}
	}

	buildQueryParams(httpRequest, input.QueryParameters)

	httpRequest.Header = input.Headers

	startTime := time.Now()

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

	endTime := time.Since(startTime)

	body, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: httpResponse.StatusCode,
		Body:       body,
		Headers:    httpResponse.Header,
		Duration:   endTime,
	}, nil
}

func buildQueryParams(httpRequest *http.Request, params url.Values) {
	if len(params) > 0 {
		requestQueryParams := httpRequest.URL.Query()

		for queryParamKey, queryParamValues := range params {
			for _, queryParamValue := range queryParamValues {
				requestQueryParams.Add(queryParamKey, queryParamValue)
			}
		}

		httpRequest.URL.RawQuery = requestQueryParams.Encode()
	}
}
