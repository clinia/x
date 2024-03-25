// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"net/http"
	"net/textproto"
	"strings"

	"github.com/clinia/x/otelhttpx"
)

// Header returns a Filter that returns true if the request
// includes a header k with a value equal to v.
func Header(k, v string) otelhttpx.Filter {
	return func(r *http.Request) bool {
		for _, hv := range r.Header[textproto.CanonicalMIMEHeaderKey(k)] {
			if v == hv {
				return true
			}
		}
		return false
	}
}

// HeaderContains returns a Filter that returns true if the request
// includes a header k with a value that contains v.
func HeaderContains(k, v string) otelhttpx.Filter {
	return func(r *http.Request) bool {
		for _, hv := range r.Header[textproto.CanonicalMIMEHeaderKey(k)] {
			if strings.Contains(hv, v) {
				return true
			}
		}
		return false
	}
}
