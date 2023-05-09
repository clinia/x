package elasticx

import (
	"net/http"
	"regexp"
)

var IsLetter = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
var IsAlphanumeric = regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString

func IsError(r *http.Response) bool {
	return r.StatusCode > 299
}
