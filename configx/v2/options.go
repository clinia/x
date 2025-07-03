package configx

import (
	"log/slog"
	"strings"
)

// Option configures the config loading behavior.
type Option func(*options)

type options struct {
	// files are YAML (or JSON) config file paths to load in order.
	files []string

	// envPrefix is the prefix for environment variables (e.g. "MYAPP_").
	// Environment variables are loaded after files, so they override file values.
	envPrefix string

	// envDelimiter is the delimiter used to map env var names to nested keys.
	// Defaults to "_". For example, with prefix "MYAPP_" and delimiter "_",
	// MYAPP_DATABASE_HOST becomes "database.host".
	envDelimiter string

	// koanfDelimiter is the key path delimiter for koanf. Defaults to ".".
	koanfDelimiter string

	// logger is used for diagnostic logging during config loading.
	logger *slog.Logger

	// values are additional key-value overrides applied after all other sources.
	values map[string]any
}

func defaultOptions() *options {
	return &options{
		envDelimiter:   "_",
		koanfDelimiter: ".",
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
// For example, with prefix "MYAPP_", the env var MYAPP_DATABASE_HOST
// maps to config key "database.host".
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

// WithEnvDelimiter sets the delimiter used to split environment variable
// names into nested config key paths. Defaults to "_".
func WithEnvDelimiter(delim string) Option {
	return func(o *options) {
		o.envDelimiter = delim
	}
}

// WithDelimiter sets the key path delimiter for koanf. Defaults to ".".
func WithDelimiter(delim string) Option {
	return func(o *options) {
		o.koanfDelimiter = delim
	}
}

// WithLogger sets the logger for diagnostic output during config loading.
func WithLogger(logger *slog.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithValues sets additional key-value overrides applied after all other sources.
// Keys use the koanf delimiter (default ".") for nested paths.
func WithValues(values map[string]any) Option {
	return func(o *options) {
		o.values = values
	}
}
