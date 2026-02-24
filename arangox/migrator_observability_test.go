package arangox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	arangoDriver "github.com/arangodb/go-driver"
	loggerxtest "github.com/clinia/x/loggerx/test"
	"github.com/clinia/x/otelx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
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
	tracer := otelx.NewTracerFromProvider(sdkTP, "arangox.migrator")
	l, buf := loggerxtest.NewTestLoggerWithJSONBuffer(t)
	_ = l // stored via NewMigrator options; returned separately so callers can build the migrator
	return &obsFixture{ctx: ctx, db: db, recorder: recorder, logBuf: buf, tracer: tracer}
}

// newObsMigrator creates a Migrator wired with the recording tracer and JSON logger.
func (f *obsFixture) newMigrator(t *testing.T, pkg string, migrations Migrations, dryRun bool) *Migrator {
	t.Helper()
	l, _ := loggerxtest.NewTestLoggerWithJSONBuffer(t)
	// Replace the log buffer with the fixture's shared buffer so we capture all logs.
	l2, buf2 := loggerxtest.NewTestLoggerWithJSONBuffer(t)
	f.logBuf = buf2
	_ = l
	return NewMigrator(NewMigratorOptions{
		Database:   f.db,
		Package:    pkg,
		Migrations: migrations,
		DryRun:     dryRun,
		Tracer:     f.tracer,
		Logger:     l2,
	})
}

// spansNamed returns all ended spans whose name matches the given suffix (e.g. "up", "up.apply").
func spansNamed(recorder *tracetest.SpanRecorder, nameSuffix string) []sdktrace.ReadOnlySpan {
	var out []sdktrace.ReadOnlySpan
	for _, s := range recorder.Ended() {
		if strings.HasSuffix(s.Name(), nameSuffix) {
			out = append(out, s)
		}
	}
	return out
}

// hasAttr checks that a span carries an attribute with the given key and value.
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

// logLines parses the JSON log buffer and returns all log entries.
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

// hasLogMsg returns true if any log entry has the given message substring.
func hasLogMsg(entries []map[string]interface{}, substring string) bool {
	for _, e := range entries {
		if msg, ok := e["msg"].(string); ok && strings.Contains(msg, substring) {
			return true
		}
	}
	return false
}

// -- Tests -------------------------------------------------------------------

func TestMigratorTracing_Up(t *testing.T) {
	f := newObsFixture(t, "test_obs_up")

	migrations := Migrations{
		{
			Version:     1,
			Description: "create users collection",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
			Down:        func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
		{
			Version:     2,
			Description: "add index",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
			Down:        func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
	}
	m := f.newMigrator(t, "obs-pkg", migrations, false)

	err := m.Up(f.ctx, 0)
	require.NoError(t, err)

	// --- span names ---
	require.NotEmpty(t, spansNamed(f.recorder, "arangox.migrator.createCollectionIfNotExist"), "expected createCollectionIfNotExist span")
	require.NotEmpty(t, spansNamed(f.recorder, "arangox.migrator.version"), "expected version span")
	upSpans := spansNamed(f.recorder, "arangox.migrator.up")
	require.Len(t, upSpans, 1, "expected exactly one up span")
	applySpans := spansNamed(f.recorder, "arangox.migrator.up.apply")
	assert.Len(t, applySpans, 2, "expected one up.apply span per migration")

	// --- outer up span attributes ---
	upSpan := upSpans[0]
	hasAttr(t, upSpan, "db.system", "arangodb")
	hasAttr(t, upSpan, "migration.direction", "up")
	hasAttr(t, upSpan, "migration.dry_run", false)
	hasAttr(t, upSpan, "migration.target_version", 0)
	assert.Equal(t, codes.Unset, upSpan.Status().Code, "up span should have no error status")

	// --- per-migration apply span attributes ---
	for i, applySpan := range applySpans {
		expectedVersion := i + 1
		hasAttr(t, applySpan, "migration.version", expectedVersion)
		assert.Equal(t, codes.Unset, applySpan.Status().Code)
	}

	// --- logs ---
	entries := logLines(f.logBuf)
	assert.True(t, hasLogMsg(entries, "running up migrations"), "expected 'running up migrations' log")
	assert.True(t, hasLogMsg(entries, "applying up migration"), "expected 'applying up migration' log")
	assert.True(t, hasLogMsg(entries, "applied"), "expected migration applied confirmation log")
}

func TestMigratorTracing_Down(t *testing.T) {
	f := newObsFixture(t, "test_obs_down")

	migrations := Migrations{
		{
			Version:     1,
			Description: "create users collection",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
			Down:        func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
		{
			Version:     2,
			Description: "add index",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
			Down:        func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
	}
	m := f.newMigrator(t, "obs-pkg", migrations, false)

	err := m.Up(f.ctx, 0)
	require.NoError(t, err)

	// Reset recorder to isolate Down spans.
	f.recorder = tracetest.NewSpanRecorder()
	sdkTP := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(f.recorder))
	m.t = otelx.NewTracerFromProvider(sdkTP, "arangox.migrator")

	err = m.Down(f.ctx, 0)
	require.NoError(t, err)

	// --- span names ---
	downSpans := spansNamed(f.recorder, "arangox.migrator.down")
	require.Len(t, downSpans, 1, "expected exactly one down span")
	applySpans := spansNamed(f.recorder, "arangox.migrator.down.apply")
	assert.Len(t, applySpans, 2, "expected one down.apply span per migration")

	// --- outer down span attributes ---
	downSpan := downSpans[0]
	hasAttr(t, downSpan, "db.system", "arangodb")
	hasAttr(t, downSpan, "migration.direction", "down")
	hasAttr(t, downSpan, "migration.dry_run", false)
	assert.Equal(t, codes.Unset, downSpan.Status().Code)

	// --- per-migration apply span attributes ---
	for _, applySpan := range applySpans {
		assert.True(t,
			spanHasAttrKey(applySpan, "migration.version"),
			"down.apply span should have migration.version attribute",
		)
		assert.True(t,
			spanHasAttrKey(applySpan, "migration.description"),
			"down.apply span should have migration.description attribute",
		)
		assert.Equal(t, codes.Unset, applySpan.Status().Code)
	}

	// --- logs ---
	entries := logLines(f.logBuf)
	assert.True(t, hasLogMsg(entries, "running down migrations"), "expected 'running down migrations' log")
}

func TestMigratorTracing_ErrorRecordedOnSpan(t *testing.T) {
	f := newObsFixture(t, "test_obs_error")
	migrationErr := errors.New("intentional migration error")

	migrations := Migrations{
		{
			Version:     1,
			Description: "failing migration",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return migrationErr },
		},
	}
	m := f.newMigrator(t, "obs-err-pkg", migrations, false)

	err := m.Up(f.ctx, 0)
	assert.ErrorIs(t, err, migrationErr)

	// The up.apply span should have error status.
	applySpans := spansNamed(f.recorder, "arangox.migrator.up.apply")
	require.Len(t, applySpans, 1)
	assert.Equal(t, codes.Error, applySpans[0].Status().Code)
	assert.Equal(t, migrationErr.Error(), applySpans[0].Status().Description)

	// The outer up span should also have error status.
	upSpans := spansNamed(f.recorder, "arangox.migrator.up")
	require.Len(t, upSpans, 1)
	assert.Equal(t, codes.Error, upSpans[0].Status().Code)
}

func TestMigratorTracing_Version(t *testing.T) {
	f := newObsFixture(t, "test_obs_version")

	migrations := Migrations{
		{
			Version:     1,
			Description: "initial",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
	}
	m := f.newMigrator(t, "obs-ver-pkg", migrations, false)

	current, latest, _, err := m.Version(f.ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(0), current)
	assert.Equal(t, uint(1), latest)

	versionSpans := spansNamed(f.recorder, "arangox.migrator.version")
	require.Len(t, versionSpans, 1)

	versionSpan := versionSpans[0]
	hasAttr(t, versionSpan, "db.system", "arangodb")
	hasAttr(t, versionSpan, "migration.current_version", 0)
	hasAttr(t, versionSpan, "migration.latest_version", 1)
	assert.Equal(t, codes.Unset, versionSpan.Status().Code)
}

func TestMigratorTracing_DryRun(t *testing.T) {
	f := newObsFixture(t, "test_obs_dryrun")

	migrations := Migrations{
		{
			Version:     1,
			Description: "dry-run migration",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
			Down:        func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
	}

	// First apply migrations normally so the collection exists, then test dry-run.
	prep := NewMigrator(NewMigratorOptions{
		Database:   f.db,
		Package:    "obs-dry-pkg",
		Migrations: migrations,
	})
	require.NoError(t, prep.Up(f.ctx, 1))

	// Now create a fresh dry-run migrator with the recording tracer.
	m := f.newMigrator(t, "obs-dry-pkg", Migrations{
		{
			Version:     2,
			Description: "second dry-run migration",
			Up:          func(ctx context.Context, db arangoDriver.Database) error { return nil },
			Down:        func(ctx context.Context, db arangoDriver.Database) error { return nil },
		},
	}, true)

	err := m.Up(f.ctx, 0)
	require.NoError(t, err)

	// Outer span should still be produced even in dry-run.
	upSpans := spansNamed(f.recorder, "arangox.migrator.up")
	require.Len(t, upSpans, 1)
	hasAttr(t, upSpans[0], "migration.dry_run", true)

	// No up.apply child spans should be produced in dry-run (migrations are not actually applied).
	applySpans := spansNamed(f.recorder, "arangox.migrator.up.apply")
	assert.Empty(t, applySpans, "dry-run should not produce apply spans")

	// Dry-run log warning should be present.
	entries := logLines(f.logBuf)
	assert.True(t, hasLogMsg(entries, "[dry-run]"), "expected dry-run warning log")
}

func TestMigratorTracing_CreateCollectionIfNotExist(t *testing.T) {
	f := newObsFixture(t, "test_obs_collection")

	m := f.newMigrator(t, "obs-col-pkg", Migrations{}, false)

	err := m.createCollectionIfNotExist(f.ctx, "test_col")
	require.NoError(t, err)

	spans := spansNamed(f.recorder, "arangox.migrator.createCollectionIfNotExist")
	require.Len(t, spans, 1)
	hasAttr(t, spans[0], "db.system", "arangodb")
	hasAttr(t, spans[0], "db.collection.name", "test_col")
	assert.Equal(t, codes.Unset, spans[0].Status().Code)

	// calling again on the same collection should still produce a span (collection exists path)
	f.recorder = tracetest.NewSpanRecorder()
	sdkTP := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(f.recorder))
	m.t = otelx.NewTracerFromProvider(sdkTP, "arangox.migrator")

	err = m.createCollectionIfNotExist(f.ctx, "test_col")
	require.NoError(t, err)
	spans = spansNamed(f.recorder, "arangox.migrator.createCollectionIfNotExist")
	require.Len(t, spans, 1, "should produce a span even when collection already exists")
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
