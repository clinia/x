package elasticxmsearch

import (
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/msearch"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/searchtype"
)

type Option func(*msearch.Msearch)

// AllowNoIndices If false, the request returns an error if any wildcard expression, index
// alias, or _all value targets only missing or closed indices. This behavior
// applies even if the request targets other open indices. For example, a
// request targeting foo*,bar* returns an error if an index starts with foo but
// no index starts with bar.
// API name: allow_no_indices
func AllowNoIndices(b bool) Option {
	return func(m *msearch.Msearch) {
		m.AllowNoIndices(b)
	}
}

// CcsMinimizeRoundtrips If true, network roundtrips between the coordinating node and remote clusters
// are minimized for cross-cluster search requests.
// API name: ccs_minimize_roundtrips
func CcsMinimizeRoundtrips(b bool) Option {
	return func(m *msearch.Msearch) {
		m.CcsMinimizeRoundtrips(b)
	}
}

// ExpandWildcards Type of index that wildcard expressions can match. If the request can target
// data streams, this argument determines whether wildcard expressions match
// hidden data streams.
// API name: expand_wildcards
func ExpandWildcards(expandwildcards ...expandwildcard.ExpandWildcard) Option {
	return func(m *msearch.Msearch) {
		m.ExpandWildcards(expandwildcards...)
	}
}

// IgnoreThrottled If true, concrete, expanded or aliased indices are ignored when frozen.
// API name: ignore_throttled
func IgnoreThrottled(b bool) Option {
	return func(m *msearch.Msearch) {
		m.IgnoreThrottled(b)
	}
}

// IgnoreUnavailable If true, missing or closed indices are not included in the response.
// API name: ignore_unavailable
func IgnoreUnavailable(b bool) Option {
	return func(m *msearch.Msearch) {
		m.IgnoreUnavailable(b)
	}
}

// MaxConcurrentSearches The number of concurrent searches the multi search api can execute.
// API name: max_concurrent_searches
func MaxConcurrentSearches(s string) Option {
	return func(m *msearch.Msearch) {
		m.MaxConcurrentSearches(s)
	}
}

// MaxConcurrentShardRequests The number of concurrent shard requests the multi search api can execute per
// search.
// API name: max_concurrent_shard_requests
func MaxConcurrentShardRequests(s string) Option {
	return func(m *msearch.Msearch) {
		m.MaxConcurrentShardRequests(s)
	}
}

// PreFilterShardSize A threshold that enforces a pre-filter roundtrip to prefilter search shards based
// on query rewriting if the number of search shards the search request expands
// to exceeds the threshold. This filter roundtrip can limit the number of
// shards significantly if for instance a shard can not match any documents
// based on it's rewrite method ie. if date filters are mandatory to match.
// API name: pre_filter_shard_size
func PreFilterShardSize(s string) Option {
	return func(m *msearch.Msearch) {
		m.PreFilterShardSize(s)
	}
}

// RestTotalHitsAsInt Indicates whether hits.total should be rendered as an integer or an object in the
// rest search response.
// API name: rest_total_hits_as_int
func RestTotalHitsAsInt(b bool) Option {
	return func(m *msearch.Msearch) {
		m.RestTotalHitsAsInt(b)
	}
}

// Routing A comma-separated list of specific routing values.
// API name: routing
func Routing(r string) Option {
	return func(m *msearch.Msearch) {
		m.Routing(r)
	}
}

// SearchType Search operation type.
// API name: search_type
func SearchType(s searchtype.SearchType) Option {
	return func(m *msearch.Msearch) {
		m.SearchType(s)
	}
}

// TypedKeys Specify whether aggregation and suggester names should be prefixed by their respective
// types in the response.
// API name: typed_keys
func TypedKeys(b bool) Option {
	return func(m *msearch.Msearch) {
		m.TypedKeys(b)
	}
}

// ErrorTrace When set to `true` Elasticsearch will include the full stack trace of errors
// when they occur.
// API name: error_trace
func ErrorTrace(b bool) Option {
	return func(m *msearch.Msearch) {
		m.ErrorTrace(b)
	}
}

// FilterPath Comma-separated list of filters in dot notation which reduce the response
// returned by Elasticsearch.
// API name: filter_path
func FilterPath(s ...string) Option {
	return func(m *msearch.Msearch) {
		m.FilterPath(s...)
	}
}

// Human Whether to return time and byte values in human-readable format.
// API name: human
func Human(b bool) Option {
	return func(m *msearch.Msearch) {
		m.Human(b)
	}
}

// Pretty If set to `true` the returned JSON will be "pretty-formatted". Only use
// this option for debugging only.
// API name: pretty
func Pretty(b bool) Option {
	return func(m *msearch.Msearch) {
		m.Pretty(b)
	}
}
