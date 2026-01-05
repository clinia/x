package slogx

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
)

var (
	sensitiveHeadersMap = make(map[string]bool, 0)
	redactionText       = "**[REDACTED]**"
	includeQuery        = false
)

// ConfigureSensitiveHeaders sets the headers that should be redacted in the logs.
// Note that this will be applied globally to all loggers using slogx.
func ConfigureSensitiveHeaders(sensitiveHeaders ...string) {
	for _, header := range sensitiveHeaders {
		sensitiveHeadersMap[strings.ToLower(header)] = true
	}
}

// ConfigureRedactionText sets the text that will be used to redact sensitive headers in the logs.
// Default is "**[REDACTED]**"
// Note that this will be applied globally to all loggers using slogx.
func ConfigureRedactionText(text string) {
	redactionText = text
}

// ConfigureIncludeQuery sets whether to include query parameters in the logs. Defaults to false
func ConfigureIncludeQuery(include bool) {
	includeQuery = include
}

// ConfigureDefaultSensitiveHeaders configures a default set of sensitive headers to be redacted in the logs.
func ConfigureDefaultSensitiveHeaders() {
	ConfigureSensitiveHeaders("Authorization")
}

func RedactHeaders(headers http.Header) slog.Attr {
	headerMap := make(map[string][]string, 0)
	for key, values := range headers {
		if sensitiveHeadersMap[strings.ToLower(key)] {
			headerMap[key] = []string{redactionText}
		} else {
			headerMap[key] = values
		}
	}

	return slog.Any("headers", headerMap)
}

func WithRequest(sl *slog.Logger, r *http.Request) *slog.Logger {
	attrs := []slog.Attr{
		RedactHeaders(r.Header),
		slog.String("proto", r.Proto),
		slog.String("method", r.Method),
		slog.String("path", r.URL.EscapedPath()),
		slog.String("host", r.Host),
	}

	if len(r.URL.RawQuery) > 0 {
		if includeQuery {
			attrs = append(attrs, slog.String("query", r.URL.RawQuery))
		} else {
			attrs = append(attrs, slog.String("query", redactionText))
		}
	}

	if ua := r.UserAgent(); len(ua) > 0 {
		attrs = append(attrs, slog.String("user-agent", ua))
	}

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}

	attrs = append(attrs, slog.String("scheme", scheme))
	attrs = append(attrs, slog.String("remote", remoteIP))

	return sl.With(slog.GroupAttrs("http_request", attrs...))
}
