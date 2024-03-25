// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otelhttpx

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
)

// Labeler is used to allow instrumented HTTP handlers to add custom attributes to
// the metrics recorded by the net/http instrumentation.
type Labeler struct {
	mu         sync.Mutex
	attributes []attribute.KeyValue
}

// Add attributes to a Labeler.
func (l *Labeler) Add(ls ...attribute.KeyValue) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attributes = append(l.attributes, ls...)
}

// Get returns a copy of the attributes added to the Labeler.
func (l *Labeler) Get() []attribute.KeyValue {
	l.mu.Lock()
	defer l.mu.Unlock()
	ret := make([]attribute.KeyValue, len(l.attributes))
	copy(ret, l.attributes)
	return ret
}

type labelerContextKeyType int

const lablelerContextKey labelerContextKeyType = 0

func injectLabeler(ctx context.Context, l *Labeler) context.Context {
	return context.WithValue(ctx, lablelerContextKey, l)
}

// LabelerFromContext retrieves a Labeler instance from the provided context if
// one is available.  If no Labeler was found in the provided context a new, empty
// Labeler is returned and the second return value is false.  In this case it is
// safe to use the Labeler but any attributes added to it will not be used.
func LabelerFromContext(ctx context.Context) (*Labeler, bool) {
	l, ok := ctx.Value(lablelerContextKey).(*Labeler)
	if !ok {
		l = &Labeler{}
	}
	return l, ok
}
