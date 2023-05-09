package elasticx

import "net/url"

// Escape the given value for use in a URL path.
func pathEscape(s string) string {
	return url.QueryEscape(s)
}
