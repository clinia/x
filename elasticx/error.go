package elasticx

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func withElasticError(res *esapi.Response) error {
	return errors.New(res.String())
}

func withHttpError(r *http.Response) error {
	var (
		out = new(bytes.Buffer)
		b1  = bytes.NewBuffer([]byte{})
		b2  = bytes.NewBuffer([]byte{})
		tr  io.Reader
	)

	if r != nil && r.Body != nil {
		tr = io.TeeReader(r.Body, b1)
		defer r.Body.Close()

		if _, err := io.Copy(b2, tr); err != nil {
			out.WriteString(fmt.Sprintf("<error reading response body: %v>", err))
			return errors.New(out.String())
		}
		defer func() { r.Body = io.NopCloser(b1) }()
	}

	if r != nil {
		out.WriteString(fmt.Sprintf("[%d %s]", r.StatusCode, http.StatusText(r.StatusCode)))
		if r.StatusCode > 0 {
			out.WriteRune(' ')
		}
	} else {
		out.WriteString("[0 <nil>]")
	}

	if r != nil && r.Body != nil {
		//nolint:errcheck
		out.ReadFrom(b2)
	}

	return errors.New(out.String())
}
