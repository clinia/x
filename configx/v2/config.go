// Package configx provides a streamlined configuration loading pipeline
// built on koanf (for multi-source loading and merging) and gozod
// (for schema validation, defaults, and JSON Schema generation).
//
// # Usage
//
// Define your configuration as a Go struct and a programmatic gozod schema:
//
//	type Config struct {
//	    Host string `json:"host"`
//	    Port int    `json:"port"`
//	}
//
//	var configSchema = gozod.Struct[Config](gozod.StructSchema{
//	    "host": gozod.String().Default("localhost"),
//	    "port": gozod.Int().Default(8080).Min(1).Max(65535),
//	})
//
// Load configuration from YAML files and environment variables:
//
//	cfg, err := configx.Load(configSchema,
//	    configx.WithFiles("config.yaml"),
//	    configx.WithEnvPrefix("MYAPP"),
//	)
//
// Generate a JSON Schema for documentation:
//
//	schema, err := configx.GenerateJSONSchema(configSchema)
package configx

import (
	"fmt"
	"strings"

	"github.com/clinia/x/errorx"
	"github.com/kaptinlin/gozod"
	koanfjson "github.com/knadh/koanf/parsers/json"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	koanfconfmap "github.com/knadh/koanf/providers/confmap"
	koanfenv "github.com/knadh/koanf/providers/env/v2"
	koanffile "github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Load loads configuration from the configured sources, merges them using koanf,
// and parses/validates the result into T using the provided gozod schema.
//
// Sources are loaded in the following order (later sources override earlier ones):
//  1. Config files (YAML/JSON) in the order specified by WithFiles
//  2. Environment variables (if WithEnvPrefix is set)
//  3. Programmatic overrides (if WithValues is set)
//
// The merged configuration map is then passed to schema.Parse, which handles
// type coercion, default values, and validation.
func Load[T any](schema gozod.ZodType[T], opts ...Option) (*T, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	k := koanf.New(o.delimiter)

	// 1. Load config files in order.
	for _, f := range o.files {
		parser, err := parserForFile(f)
		if err != nil {
			return nil, errorx.InvalidArgumentErrorf("failed to load config file %q: %s", f, err)
		}
		if err := k.Load(koanffile.Provider(f), parser); err != nil {
			return nil, errorx.InvalidArgumentErrorf("failed to load config file %q: %s", f, err)
		}
	}

	// 2. Load environment variables.
	if o.envPrefix != "" {
		provider := koanfenv.Provider(o.delimiter, koanfenv.Opt{
			Prefix:        o.envPrefix,
			TransformFunc: envTransformFunc(o.envPrefix, o.delimiter),
		})
		if err := k.Load(provider, nil); err != nil {
			return nil, errorx.InternalErrorf("failed to load environment variables: %s", err)
		}
	}

	// 3. Apply programmatic overrides.
	if len(o.values) > 0 {
		if err := k.Load(koanfconfmap.Provider(o.values, o.delimiter), nil); err != nil {
			return nil, errorx.InternalErrorf("failed to apply config values: %s", err)
		}
	}

	// 4. Get the merged nested config map and parse through gozod schema.
	raw := k.Raw()

	result, err := schema.Parse(raw)
	if err != nil {
		return nil, errorx.InvalidArgumentErrorf("config validation failed: %s", err)
	}

	return &result, nil
}

// parserForFile returns the appropriate koanf parser for the given file path.
func parserForFile(path string) (koanf.Parser, error) {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"):
		return koanfyaml.Parser(), nil
	case strings.HasSuffix(lower, ".json"):
		return koanfjson.Parser(), nil
	default:
		return nil, fmt.Errorf("unsupported config file format: %q (supported: .yaml, .yml, .json)", path)
	}
}
