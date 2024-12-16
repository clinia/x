package httpx

import (
	"net/http"

	"github.com/clinia/x/errorx"
)

const (
	CliniaHealthyHeaderKey = "X-Clinia-Healthy"
	CliniaHealthyValue     = "true"
	CliniaUnHeahlthyValue  = "false"
)

func SetCliniaHealthyHeader(r *http.Request) error {
	if r == nil {
		return errorx.InternalErrorf("request can not be nil")
	}
	r.Header.Add(CliniaHealthyHeaderKey, CliniaHealthyValue)
	return nil
}

func SetCliniaUnHealthyHeader(r *http.Request) error {
	if r == nil {
		return errorx.InternalErrorf("request can not be nil")
	}
	r.Header.Add(CliniaHealthyHeaderKey, CliniaUnHeahlthyValue)
	return nil
}
