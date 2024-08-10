package elasticxsearch

import (
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/expandwildcard"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/operator"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/searchtype"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/suggestmode"
)

type Option func(*search.Search)

// AllowNoIndices If `false`, the request returns an error if any wildcard expression, index
// alias, or `_all` value targets only missing or closed indices.
// This behavior applies even if the request targets other open indices.
// For example, a request targeting `foo*,bar*` returns an error if an index
// starts with `foo` but no index starts with `bar`.
// API name: allow_no_indices
func AllowNoIndices(b bool) Option {
	return func(s *search.Search) {
		s.AllowNoIndices(b)
	}
}

// AllowPartialSearchResults If true, returns partial results if there are shard request timeouts or shard
// failures. If false, returns an error with no partial results.
// API name: allow_partial_search_results
func AllowPartialSearchResults(b bool) Option {
	return func(s *search.Search) {
		s.AllowPartialSearchResults(b)
	}
}

// Analyzer Analyzer to use for the query string.
// This parameter can only be used when the q query string parameter is
// specified.
// API name: analyzer
func Analyzer(a string) Option {
	return func(s *search.Search) {
		s.Analyzer(a)
	}
}

// AnalyzeWildcard If true, wildcard and prefix queries are analyzed.
// This parameter can only be used when the q query string parameter is
// specified.
// API name: analyze_wildcard
func AnalyzeWildcard(b bool) Option {
	return func(s *search.Search) {
		s.AnalyzeWildcard(b)
	}
}

// BatchedReduceSize The number of shard results that should be reduced at once on the
// coordinating node.
// This value should be used as a protection mechanism to reduce the memory
// overhead per search request if the potential number of shards in the request
// can be large.
// API name: batched_reduce_size
func BatchedReduceSize(b string) Option {
	return func(s *search.Search) {
		s.BatchedReduceSize(b)
	}
}

// CcsMinimizeRoundtrips If true, network round-trips between the coordinating node and the remote
// clusters are minimized when executing cross-cluster search (CCS) requests.
// API name: ccs_minimize_roundtrips
func CcsMinimizeRoundtrips(b bool) Option {
	return func(s *search.Search) {
		s.CcsMinimizeRoundtrips(b)
	}
}

// DefaultOperator The default operator for query string query: AND or OR.
// This parameter can only be used when the `q` query string parameter is
// specified.
// API name: default_operator
func DefaultOperator(d operator.Operator) Option {
	return func(s *search.Search) {
		s.DefaultOperator(d)
	}
}

// Df Field to use as default where no field prefix is given in the query string.
// This parameter can only be used when the q query string parameter is
// specified.
// API name: df
func Df(d string) Option {
	return func(s *search.Search) {
		s.Df(d)
	}
}

// ExpandWildcards Type of index that wildcard patterns can match.
// If the request can target data streams, this argument determines whether
// wildcard expressions match hidden data streams.
// Supports comma-separated values, such as `open,hidden`.
// API name: expand_wildcards
func ExpandWildcards(expandwildcards ...expandwildcard.ExpandWildcard) Option {
	return func(s *search.Search) {
		s.ExpandWildcards(expandwildcards...)
	}
}

// IgnoreThrottled If `true`, concrete, expanded or aliased indices will be ignored when frozen.
// API name: ignore_throttled
func IgnoreThrottled(b bool) Option {
	return func(s *search.Search) {
		s.IgnoreThrottled(b)
	}
}

// IgnoreUnavailable If `false`, the request returns an error if it targets a missing or closed
// index.
// API name: ignore_unavailable
func IgnoreUnavailable(b bool) Option {
	return func(s *search.Search) {
		s.IgnoreUnavailable(b)
	}
}

// Lenient If `true`, format-based query failures (such as providing text to a numeric
// field) in the query string will be ignored.
// This parameter can only be used when the `q` query string parameter is
// specified.
// API name: lenient
func Lenient(b bool) Option {
	return func(s *search.Search) {
		s.Lenient(b)
	}
}

// MaxConcurrentShardRequests Defines the number of concurrent shard requests per node this search executes
// concurrently.
// This value should be used to limit the impact of the search on the cluster in
// order to limit the number of concurrent shard requests.
// API name: max_concurrent_shard_requests
func MaxConcurrentShardRequests(m string) Option {
	return func(s *search.Search) {
		s.MaxConcurrentShardRequests(m)
	}
}

// MinCompatibleShardNode The minimum version of the node that can handle the request
// Any handling node with a lower version will fail the request.
// API name: min_compatible_shard_node
func MinCompatibleShardNode(m string) Option {
	return func(s *search.Search) {
		s.MinCompatibleShardNode(m)
	}
}

// Preference Nodes and shards used for the search.
// By default, Elasticsearch selects from eligible nodes and shards using
// adaptive replica selection, accounting for allocation awareness. Valid values
// are:
// `_only_local` to run the search only on shards on the local node;
// `_local` to, if possible, run the search on shards on the local node, or if
// not, select shards using the default method;
// `_only_nodes:<node-id>,<node-id>` to run the search on only the specified
// nodes IDs, where, if suitable shards exist on more than one selected node,
// use shards on those nodes using the default method, or if none of the
// specified nodes are available, select shards from any available node using
// the default method;
// `_prefer_nodes:<node-id>,<node-id>` to if possible, run the search on the
// specified nodes IDs, or if not, select shards using the default method;
// `_shards:<shard>,<shard>` to run the search only on the specified shards;
// `<custom-string>` (any string that does not start with `_`) to route searches
// with the same `<custom-string>` to the same shards in the same order.
// API name: preference
func Preference(p string) Option {
	return func(s *search.Search) {
		s.Preference(p)
	}
}

// PreFilterShardSize Defines a threshold that enforces a pre-filter roundtrip to prefilter search
// shards based on query rewriting if the number of shards the search request
// expands to exceeds the threshold.
// This filter roundtrip can limit the number of shards significantly if for
// instance a shard can not match any documents based on its rewrite method (if
// date filters are mandatory to match but the shard bounds and the query are
// disjoint).
// When unspecified, the pre-filter phase is executed if any of these conditions
// is met:
// the request targets more than 128 shards;
// the request targets one or more read-only index;
// the primary sort of the query targets an indexed field.
// API name: pre_filter_shard_size
func PreFilterShardSize(p string) Option {
	return func(s *search.Search) {
		s.PreFilterShardSize(p)
	}
}

// RequestCache If `true`, the caching of search results is enabled for requests where `size`
// is `0`.
// Defaults to index level settings.
// API name: request_cache
func RequestCache(b bool) Option {
	return func(s *search.Search) {
		s.RequestCache(b)
	}
}

// Routing Custom value used to route operations to a specific shard.
// API name: routing
func Routing(r string) Option {
	return func(s *search.Search) {
		s.Routing(r)
	}
}

// Scroll Period to retain the search context for scrolling. See Scroll search results.
// By default, this value cannot exceed `1d` (24 hours).
// You can change this limit using the `search.max_keep_alive` cluster-level
// setting.
// API name: scroll
func Scroll(duration string) Option {
	return func(s *search.Search) {
		s.Scroll(duration)
	}
}

// SearchType How distributed term frequencies are calculated for relevance scoring.
// API name: search_type
func SearchType(searchtype searchtype.SearchType) Option {
	return func(s *search.Search) {
		s.SearchType(searchtype)
	}
}

// SuggestField Specifies which field to use for suggestions.
// API name: suggest_field
func SuggestField(field string) Option {
	return func(s *search.Search) {
		s.SuggestField(field)
	}
}

// SuggestMode Specifies the suggest mode.
// This parameter can only be used when the `suggest_field` and `suggest_text`
// query string parameters are specified.
// API name: suggest_mode
func SuggestMode(suggestmode suggestmode.SuggestMode) Option {
	return func(s *search.Search) {
		s.SuggestMode(suggestmode)
	}
}

// SuggestSize Number of suggestions to return.
// This parameter can only be used when the `suggest_field` and `suggest_text`
// query string parameters are specified.
// API name: suggest_size
func SuggestSize(suggestsize string) Option {
	return func(s *search.Search) {
		s.SuggestSize(suggestsize)
	}
}

// SuggestText The source text for which the suggestions should be returned.
// This parameter can only be used when the `suggest_field` and `suggest_text`
// query string parameters are specified.
// API name: suggest_text
func SuggestText(suggesttext string) Option {
	return func(s *search.Search) {
		s.SuggestText(suggesttext)
	}
}

// TypedKeys If `true`, aggregation and suggester names are be prefixed by their
// respective types in the response.
// API name: typed_keys
func TypedKeys(typedkeys bool) Option {
	return func(s *search.Search) {
		s.TypedKeys(typedkeys)
	}
}

// RestTotalHitsAsInt Indicates whether `hits.total` should be rendered as an integer or an object
// in the rest search response.
// API name: rest_total_hits_as_int
func RestTotalHitsAsInt(b bool) Option {
	return func(s *search.Search) {
		s.RestTotalHitsAsInt(b)
	}
}

// SourceExcludes_ A comma-separated list of source fields to exclude from the response.
// You can also use this parameter to exclude fields from the subset specified
// in `_source_includes` query parameter.
// If the `_source` parameter is `false`, this parameter is ignored.
// API name: _source_excludes
func SourceExcludes_(sourceexcludes ...string) Option {
	return func(s *search.Search) {
		s.SourceExcludes_(sourceexcludes...)
	}
}

// SourceIncludes_ A comma-separated list of source fields to include in the response.
// If this parameter is specified, only these source fields are returned.
// You can exclude fields from this subset using the `_source_excludes` query
// parameter.
// If the `_source` parameter is `false`, this parameter is ignored.
// API name: _source_includes
func SourceIncludes_(sourceincludes ...string) Option {
	return func(s *search.Search) {
		s.SourceIncludes_(sourceincludes...)
	}
}

// Q Query in the Lucene query string syntax using query parameter search.
// Query parameter searches do not support the full Elasticsearch Query DSL but
// are handy for testing.
// API name: q
func Q(q string) Option {
	return func(s *search.Search) {
		s.Q(q)
	}
}

// ForceSyntheticSource Should this request force synthetic _source?
// Use this to test if the mapping supports synthetic _source and to get a sense
// of the worst case performance.
// Fetches with this enabled will be slower the enabling synthetic source
// natively in the index.
// API name: force_synthetic_source
func ForceSyntheticSource(b bool) Option {
	return func(s *search.Search) {
		s.ForceSyntheticSource(b)
	}
}

// ErrorTrace When set to `true` Elasticsearch will include the full stack trace of errors
// when they occur.
// API name: error_trace
func ErrorTrace(b bool) Option {
	return func(s *search.Search) {
		s.ErrorTrace(b)
	}
}

// FilterPath Comma-separated list of filters in dot notation which reduce the response
// returned by Elasticsearch.
// API name: filter_path
func FilterPath(filterpath ...string) Option {
	return func(s *search.Search) {
		s.FilterPath(filterpath...)
	}
}

// Human When set to `true` will return statistics in a format suitable for humans.
// For example `"exists_time": "1h"` for humans and
// `"eixsts_time_in_millis": 3600000` for computers. When disabled the human
// readable values will be omitted. This makes sense for responses being
// consumed
// only by machines.
// API name: human
func Human(b bool) Option {
	return func(s *search.Search) {
		s.Human(b)
	}
}

// Pretty If set to `true` the returned JSON will be "pretty-formatted". Only use
// this option for debugging only.
// API name: pretty
func Pretty(b bool) Option {
	return func(s *search.Search) {
		s.Pretty(b)
	}
}
