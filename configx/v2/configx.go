// Package configx provides a streamlined configuration loading pipeline
// built on koanf (for multi-source loading and merging) and gozod
// (for schema validation, defaults, and JSON Schema generation).
//
// # Usage
//
// Define your configuration as a Go struct with gozod tags for validation
// and json/yaml tags for serialization:
//
//	type Config struct {
//	    Server   ServerConfig   `json:"server"   yaml:"server"   gozod:"required"`
//	    Database DatabaseConfig `json:"database" yaml:"database" gozod:"required"`
//	}
//
//	type ServerConfig struct {
//	    Host string `json:"host" yaml:"host" gozod:"required"`
//	    Port int    `json:"port" yaml:"port" gozod:"required,min=1,max=65535"`
//	}
//
//	type DatabaseConfig struct {
//	    DSN string `json:"dsn" yaml:"dsn" gozod:"required"`
//	}
//
// Load configuration from YAML files and environment variables:
//
//	cfg, err := configx.Load[Config](
//	    configx.WithFiles("config.yaml"),
//	    configx.WithEnvPrefix("MYAPP"),
//	)
//
// Generate a JSON Schema for documentation:
//
//	schema, err := configx.Schema[Config]()
package configx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kaptinlin/gozod"
	koanfjson "github.com/knadh/koanf/parsers/json"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	koanfconfmap "github.com/knadh/koanf/providers/confmap"
	koanfenv "github.com/knadh/koanf/providers/env/v2"
	koanffile "github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Load loads configuration from the configured sources, merges them using koanf,
// and parses/validates the result into T using gozod.
//
// Sources are loaded in the following order (later sources override earlier ones):
//  1. Config files (YAML/JSON) in the order specified by WithFiles
//  2. Environment variables (if WithEnvPrefix is set)
//  3. Programmatic overrides (if WithValues is set)
//
// The merged configuration map is then parsed through the gozod schema derived
// from struct tags on T, which handles validation and default values.
func Load[T any](opts ...Option) (*T, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	k := koanf.New(o.koanfDelimiter)

	// 1. Load config files
	for _, f := range o.files {
		parser, err := parserForFile(f)
		if err != nil {
			return nil, fmt.Errorf("configx: %w", err)
		}
		if err := k.Load(koanffile.Provider(f), parser); err != nil {
			return nil, fmt.Errorf("configx: loading file %q: %w", f, err)
		}
		if o.logger != nil {
			o.logger.Info("loaded config file", "file", f)
		}
	}

	// 2. Load environment variables
	if o.envPrefix != "" {
		envProvider := koanfenv.Provider(o.koanfDelimiter, koanfenv.Opt{
			Prefix: o.envPrefix,
			TransformFunc: func(key, value string) (string, any) {
				key = strings.TrimPrefix(key, o.envPrefix)
				key = strings.ToLower(strings.ReplaceAll(key, o.envDelimiter, o.koanfDelimiter))
				return key, value
			},
		})

		if err := k.Load(envProvider, nil); err != nil {
			return nil, fmt.Errorf("configx: loading env vars: %w", err)
		}
		if o.logger != nil {
			o.logger.Info("loaded environment variables", "prefix", o.envPrefix)
		}
	}

	// 3. Apply programmatic overrides
	if len(o.values) > 0 {
		if err := k.Load(koanfconfmap.Provider(o.values, o.koanfDelimiter), nil); err != nil {
			return nil, fmt.Errorf("configx: loading values: %w", err)
		}
	}

	// 4. Get the merged config as a flat map
	raw := k.All()

	if o.logger != nil {
		o.logger.Debug("merged config", "keys", len(raw))
	}

	// 5. Parse and validate through gozod
	schema := gozod.FromStruct[T]()
	result, err := schema.Parse(mapToStruct[T](raw))
	if err != nil {
		return nil, fmt.Errorf("configx: validation failed: %w", err)
	}

	return &result, nil
}

// mapToStruct converts a map[string]any to a struct T by marshaling to JSON
// and unmarshaling back. This ensures proper type coercion for gozod parsing.
func mapToStruct[T any](m map[string]any) T {
	var result T
	b, err := json.Marshal(m)
	if err != nil {
		return result
	}
	_ = json.Unmarshal(b, &result)
	return result
}

// SchemaOption configures JSON Schema generation.
type SchemaOption func(*schemaOptions)

type schemaOptions struct {
	indent bool
}

// WithIndent produces indented JSON Schema output.
func WithIndent() SchemaOption {
	return func(o *schemaOptions) {
		o.indent = true
	}
}

// Schema generates a JSON Schema (Draft 2020-12) from the gozod schema
// derived from struct tags on T. This can be used to provide documentation
// and IDE support for YAML config files.
func Schema[T any](opts ...SchemaOption) ([]byte, error) {
	o := &schemaOptions{}
	for _, opt := range opts {
		opt(o)
	}

	schema := gozod.FromStruct[T]()

	jsonSchema, err := gozod.ToJSONSchema(schema, gozod.JSONSchemaOptions{
		IO: "input",
	})
	if err != nil {
		return nil, fmt.Errorf("configx: generating JSON schema: %w", err)
	}

	if o.indent {
		return json.MarshalIndent(jsonSchema, "", "  ")
	}
	return json.Marshal(jsonSchema)
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
