package configx

import (
	"encoding/json"
	"fmt"

	"github.com/kaptinlin/gozod"
)

// GenerateJSONSchema converts a gozod schema into a JSON Schema (Draft 2020-12)
// byte slice. The schema parameter can be any gozod schema type (e.g. a
// *gozod.ZodStruct, *gozod.ZodObject, etc.).
//
// The generated schema uses IO: "input" mode so that defaults and optionality
// reflect the input/parsing perspective.
//
// Example:
//
//	schema := gozod.Struct[Config](gozod.StructSchema{
//	    "host": gozod.String().Default("localhost"),
//	    "port": gozod.Int().Default(8080),
//	})
//	jsonSchemaBytes, err := configx.GenerateJSONSchema(schema)
func GenerateJSONSchema(schema any, opts ...gozod.JSONSchemaOptions) ([]byte, error) {
	var opt gozod.JSONSchemaOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.IO == "" {
		opt.IO = "input"
	}

	jsonSchema, err := gozod.ToJSONSchema(schema, opt)
	if err != nil {
		return nil, fmt.Errorf("configx: generating JSON schema: %w", err)
	}

	return json.MarshalIndent(jsonSchema, "", "  ")
}
