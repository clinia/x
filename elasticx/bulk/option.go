package elasticxbulk

import (
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
)

type Option func(*bulk.Bulk)

// Refresh If `true`, Elasticsearch refreshes the affected shards to make this operation
// visible to search, if `wait_for` then wait for a refresh to make this
// operation visible to search, if `false` do nothing with refreshes.
// Valid values: `true`, `false`, `wait_for`.
// API name: refresh
func Refresh(r refresh.Refresh) Option {
	return func(b *bulk.Bulk) {
		b.Refresh(r)
	}
}

// Pipeline ID of the pipeline to use to preprocess incoming documents.
// If the index has a default ingest pipeline specified, then setting the value
// to `_none` disables the default ingest pipeline for this request.
// If a final pipeline is configured it will always run, regardless of the value
// of this parameter.
// API name: pipeline
func Pipeline(p string) Option {
	return func(b *bulk.Bulk) {
		b.Pipeline(p)
	}
}

// Routing Custom value used to route operations to a specific shard.
// API name: routing
func Routing(r string) Option {
	return func(b *bulk.Bulk) {
		b.Routing(r)
	}
}

// Source_ `true` or `false` to return the `_source` field or not, or a list of fields
// to return.
// API name: _source
func Source(s string) Option {
	return func(b *bulk.Bulk) {
		b.Source_(s)
	}
}

// SourceExcludes_ A comma-separated list of source fields to exclude from the response.
// API name: _source_excludes
func SourceExcludes(s ...string) Option {
	return func(b *bulk.Bulk) {
		b.SourceExcludes_(s...)
	}
}

// SourceIncludes_ A comma-separated list of source fields to include in the response.
// API name: _source_includes
func SourceIncludes(s ...string) Option {
	return func(b *bulk.Bulk) {
		b.SourceIncludes_(s...)
	}
}

// Timeout Explicit operation timeout.
// API name: timeout
func Timeout(t string) Option {
	return func(b *bulk.Bulk) {
		b.Timeout(t)
	}
}

// WaitForActiveShards The number of shard copies that must be active before proceeding with the
// operation.
// Set to all or any positive integer up to the total number of shards in the
// index (`number_of_replicas+1`).
// API name: wait_for_active_shards
func WaitForActiveShards(w string) Option {
	return func(b *bulk.Bulk) {
		b.WaitForActiveShards(w)
	}
}

// RequireAlias If true, requires destination to be an alias.
// API name: require_alias
func RequireAlias(r bool) Option {
	return func(b *bulk.Bulk) {
		b.RequireAlias(r)
	}
}

// ErrorTrace When set to `true` Elasticsearch will include the full stack trace of errors
// when they occur.
// API name: error_trace
func ErrorTrace(e bool) Option {
	return func(b *bulk.Bulk) {
		b.ErrorTrace(e)
	}
}

// FilterPath Comma-separated list of filters used to reduce the response.
// API name: filter_path
func FilterPath(f ...string) Option {
	return func(b *bulk.Bulk) {
		b.FilterPath(f...)
	}
}

// Human If true, the response will be prettified.
// API name: human
func Human(h bool) Option {
	return func(b *bulk.Bulk) {
		b.Human(h)
	}
}

// Pretty If true, the response will be prettified.
// API name: pretty
func Pretty(p bool) Option {
	return func(b *bulk.Bulk) {
		b.Pretty(p)
	}
}
