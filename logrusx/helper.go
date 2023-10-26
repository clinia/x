// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package logrusx

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	*logrus.Entry
	leakSensitive bool
	redactionText string
	opts          []Option
	name          string
	version       string
}

var opts = otelhttptrace.WithPropagators(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

func (l *Logger) LeakSensitiveData() bool {
	return l.leakSensitive
}

func (l *Logger) Logrus() *logrus.Logger {
	return l.Entry.Logger
}

func (l *Logger) NewEntry() *Logger {
	ll := *l
	ll.Entry = logrus.NewEntry(l.Logger)
	return &ll
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	ll := *l
	ll.Entry = l.Logger.WithContext(ctx)
	return &ll
}

func (l *Logger) HTTPHeadersRedacted(h http.Header) map[string]interface{} {
	headers := map[string]interface{}{}

	for key, value := range h {
		keyLower := strings.ToLower(key)
		if keyLower == "authorization" || keyLower == "cookie" || keyLower == "set-cookie" {
			headers[keyLower] = l.maybeRedact(value)
		} else {
			headers[keyLower] = h.Get(key)
		}
	}

	return headers
}

func (l *Logger) WithRequest(r *http.Request) *Logger {
	headers := l.HTTPHeadersRedacted(r.Header)
	if ua := r.UserAgent(); len(ua) > 0 {
		headers["user-agent"] = ua
	}

	remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIP = r.RemoteAddr
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}

	ll := l.WithField("http_request", map[string]interface{}{
		"remote":  remoteIP,
		"proto":   r.Proto,
		"method":  r.Method,
		"path":    r.URL.EscapedPath(),
		"query":   l.maybeRedact(r.URL.RawQuery),
		"scheme":  scheme,
		"host":    r.Host,
		"headers": headers,
	})

	spanCtx := trace.SpanContextFromContext(r.Context())
	if !spanCtx.IsValid() {
		_, _, spanCtx = otelhttptrace.Extract(r.Context(), r, opts)
	}
	if spanCtx.IsValid() {
		if spanCtx.HasTraceID() {
			ll = ll.WithField("TraceID", spanCtx.TraceID().String())
		}
		if spanCtx.HasSpanID() {
			ll = ll.WithField("SpanID", spanCtx.SpanID().String())
		}
	}
	return ll
}

func (l *Logger) Logf(level logrus.Level, format string, args ...interface{}) {
	// Add traces information if available in context
	if l.Context != nil {
		spanCtx := trace.SpanContextFromContext(l.Context)
		if spanCtx.IsValid() {
			if spanCtx.HasTraceID() {
				l = l.WithField("TraceID", spanCtx.TraceID().String())
			}
			if spanCtx.HasSpanID() {
				l = l.WithField("SpanID", spanCtx.SpanID().String())
			}
		}
	}
	if !l.leakSensitive {
		for i, arg := range args {
			switch urlArg := arg.(type) {
			case url.URL:
				urlCopy := url.URL{Scheme: urlArg.Scheme, Host: urlArg.Host, Path: urlArg.Path}
				args[i] = urlCopy
			case *url.URL:
				urlCopy := url.URL{Scheme: urlArg.Scheme, Host: urlArg.Host, Path: urlArg.Path}
				args[i] = &urlCopy
			default:
				continue
			}
		}
	}
	l.Entry.Logf(level, format, args...)
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Logf(logrus.TraceLevel, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logf(logrus.DebugLevel, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logf(logrus.InfoLevel, format, args...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logf(logrus.WarnLevel, format, args...)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logf(logrus.ErrorLevel, format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logf(logrus.FatalLevel, format, args...)
	l.Entry.Logger.Exit(1)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Logf(logrus.PanicLevel, format, args...)
}

func (l *Logger) WithFields(f logrus.Fields) *Logger {
	ll := *l
	ll.Entry = l.Entry.WithFields(f)
	return &ll
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	ll := *l
	ll.Entry = l.Entry.WithField(key, value)
	return &ll
}

func (l *Logger) maybeRedact(value interface{}) interface{} {
	if fmt.Sprintf("%v", value) == "" || value == nil {
		return nil
	}
	if !l.leakSensitive {
		return l.redactionText
	}
	return value
}

func (l *Logger) WithSensitiveField(key string, value interface{}) *Logger {
	return l.WithField(key, l.maybeRedact(value))
}

func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}

	ctx := map[string]interface{}{"message": err.Error()}
	// if l.Entry.Logger.IsLevelEnabled(logrus.DebugLevel) {
	// 	if e, ok := err.(errorsx.StackTracer); ok {
	// 		ctx["stack_trace"] = fmt.Sprintf("%+v", e.StackTrace())
	// 	} else {
	// 		ctx["stack_trace"] = fmt.Sprintf("stack trace could not be recovered from error type %s", reflect.TypeOf(err))
	// 	}
	// }
	// if c := errorsx.ReasonCarrier(nil); errors.As(err, &c) {
	// 	ctx["reason"] = c.Reason()
	// }
	// if c := errorsx.RequestIDCarrier(nil); errors.As(err, &c) && c.RequestID() != "" {
	// 	ctx["request_id"] = c.RequestID()
	// }
	// if c := errorsx.DetailsCarrier(nil); errors.As(err, &c) && c.Details() != nil {
	// 	ctx["details"] = c.Details()
	// }
	// if c := errorsx.StatusCarrier(nil); errors.As(err, &c) && c.Status() != "" {
	// 	ctx["status"] = c.Status()
	// }
	// if c := errorsx.StatusCodeCarrier(nil); errors.As(err, &c) && c.StatusCode() != 0 {
	// 	ctx["status_code"] = c.StatusCode()
	// }
	// if c := errorsx.DebugCarrier(nil); errors.As(err, &c) {
	// 	ctx["debug"] = c.Debug()
	// }

	return l.WithField("error", ctx)
}
