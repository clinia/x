package slogx

import (
	"context"
	"log/slog"
	"time"

	slogctx "github.com/veqryn/slog-context"
)

func NewRequestIDExtractor(requestIDContextKey interface{}, requestIDFieldKey string) slogctx.AttrExtractor {
	return func(ctx context.Context, recordT time.Time, recordLvl slog.Level, recordMsg string) []slog.Attr {
		defer func() {
			// Nullify panic to prevent having this hook break a request
			recover()
		}()

		requestID := ctx.Value(requestIDContextKey)
		if requestID == nil {
			return nil
		}
		return []slog.Attr{slog.Any(requestIDFieldKey, requestID)}
	}
}
