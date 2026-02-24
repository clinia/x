package arangox

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	arangoDriver "github.com/arangodb/go-driver"
	loggerxtest "github.com/clinia/x/loggerx/test"
	"github.com/clinia/x/otelx"
	"github.com/stretchr/testify/assert"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// obsFixture bundles the test dependencies for observability assertions.
type obsFixture struct {
	ctx      context.Context
	db       arangoDriver.Database
	recorder *tracetest.SpanRecorder
	logBuf   *bytes.Buffer
	tracer   *otelx.Tracer
}

func newObsFixture(t *testing.T, dbName string) *obsFixture {
	t.Helper()
	ctx, db := newFixture(t, dbName)
	recorder := tracetest.NewSpanRecorder()
	sdkTP := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	tracer := otelx.NewTracerFromProvider(sdkTP, migratorComponentName)
	_, buf := loggerxtest.NewTestLoggerWithJSONBuffer(t)
	return &obsFixture{ctx: ctx, db: db, recorder: recorder, logBuf: buf, tracer: tracer}
}

// newMigrator creates a Migrator wired with the recording tracer and a fresh JSON log buffer.
func (f *obsFixture) newMigrator(t *testing.T, pkg string, migrations Migrations, dryRun bool) *Migrator {
	t.Helper()
	l, buf := loggerxtest.NewTestLoggerWithJSONBuffer(t)
	f.logBuf = buf
	return NewMigrator(NewMigratorOptions{
		Database:   f.db,
		Package:    pkg,
		Migrations: migrations,
		DryRun:     dryRun,
		Tracer:     f.tracer,
		Logger:     l,
	})
}

// spansNamed returns all ended spans whose name ends with nameSuffix
// (e.g. "up", "up.apply", "arangox.migrator.version").
func spansNamed(recorder *tracetest.SpanRecorder, nameSuffix string) []sdktrace.ReadOnlySpan {
	var out []sdktrace.ReadOnlySpan
	for _, s := range recorder.Ended() {
		if strings.HasSuffix(s.Name(), nameSuffix) {
			out = append(out, s)
		}
	}
	return out
}

// hasAttr asserts that a span carries an attribute with the given key and value.
func hasAttr(t *testing.T, span sdktrace.ReadOnlySpan, key string, value interface{}) {
	t.Helper()
	for _, attr := range span.Attributes() {
		if string(attr.Key) != key {
			continue
		}
		switch v := value.(type) {
		case string:
			assert.Equal(t, v, attr.Value.AsString(), "attribute %q value mismatch", key)
		case int:
			assert.Equal(t, int64(v), attr.Value.AsInt64(), "attribute %q value mismatch", key)
		case bool:
			assert.Equal(t, v, attr.Value.AsBool(), "attribute %q value mismatch", key)
		}
		return
	}
	t.Errorf("span %q: attribute %q not found", span.Name(), key)
}

// spanHasAttrKey returns true if any attribute on the span has the given key.
func spanHasAttrKey(span sdktrace.ReadOnlySpan, key string) bool {
	for _, attr := range span.Attributes() {
		if string(attr.Key) == key {
			return true
		}
	}
	return false
}

// logLines parses a JSON-lines log buffer and returns all decoded entries.
func logLines(buf *bytes.Buffer) []map[string]interface{} {
	var entries []map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))
	for decoder.More() {
		var entry map[string]interface{}
		if err := decoder.Decode(&entry); err != nil {
			break
		}
		entries = append(entries, entry)
	}
	return entries
}

// hasLogMsg returns true if any log entry message contains the given substring.
func hasLogMsg(entries []map[string]interface{}, substring string) bool {
	for _, e := range entries {
		if msg, ok := e["msg"].(string); ok && strings.Contains(msg, substring) {
			return true
		}
	}
	return false
}
