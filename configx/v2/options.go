package configx

import "strings"

// Option configures the config loading behavior.
type Option func(*options)

type options struct {
	// files are YAML (or JSON) config file paths to load in order.
	files []string

	// envPrefix is the prefix for environment variables (e.g. "MYAPP_").
	// Environment variables are loaded after files, so they override file values.
	envPrefix string

	// delimiter is the key path delimiter for koanf. Defaults to ".".
	delimiter string

	// values are additional key-value overrides applied after all other sources.
	// Keys use the delimiter (default ".") for nested paths.
	values map[string]any
}

func defaultOptions() *options {
	return &options{
		delimiter: ".",
	}
}

// WithFiles specifies YAML/JSON config files to load in order.
// Later files override values from earlier files.
func WithFiles(files ...string) Option {
	return func(o *options) {
		o.files = append(o.files, files...)
	}
}

// WithEnvPrefix sets the prefix for environment variable loading.
// The prefix is stripped from env var names before mapping to config keys.
// A trailing underscore is added automatically if not present.
//
// Nesting is indicated by double underscores (__) in the env var name.
// Single underscores are preserved as literal underscores in the key.
// For example, with prefix "APP":
//
//	APP_SERVER__HOST       → server.host
//	APP_DATABASE__MAX_CONNS → database.max_conns
//
// If not set, environment variables are not loaded.
func WithEnvPrefix(prefix string) Option {
	return func(o *options) {
		o.envPrefix = strings.ToUpper(prefix)
		if o.envPrefix != "" && !strings.HasSuffix(o.envPrefix, "_") {
			o.envPrefix += "_"
		}
	}
}

// WithDelimiter sets the key path delimiter for koanf. Defaults to ".".
func WithDelimiter(delim string) Option {
	return func(o *options) {
		o.delimiter = delim
	}
}

// WithValues sets additional key-value overrides applied after all other sources.
// Keys use the koanf delimiter (default ".") for nested paths.
// For example: map[string]any{"server.host": "localhost", "server.port": 8080}
func WithValues(values map[string]any) Option {
	return func(o *options) {
		o.values = values
	}
}
