package httpx

import (
	"net/http"

	"github.com/clinia/x/errorx"
)

const (
	CliniaHealthyHeaderKey = "X-Clinia-Healthy"
	CliniaHealthyValue     = "true"
	CliniaUnHealthyValue   = "false"
)

func SetCliniaHealthyHeader(w http.ResponseWriter) error {
	if w == nil {
		return errorx.InternalErrorf("response writer cannot be nil")
	}
	w.Header().Add(CliniaHealthyHeaderKey, CliniaHealthyValue)
	return nil
}

func SetCliniaUnHealthyHeader(w http.ResponseWriter) error {
	if w == nil {
		return errorx.InternalErrorf("response writer cannot be nil")
	}
	w.Header().Add(CliniaHealthyHeaderKey, CliniaUnHealthyValue)
	return nil
}
