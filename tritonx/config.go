package tritonx

import (
	"bytes"
	_ "embed"
	"io"
)

type Config struct {
	Insecure bool   `json:"insecure"`
	Host     string `json:"host"`
}

//go:embed config.schema.json
var ConfigSchema string

const ConfigSchemaID = "clinia://triton-config"

// AddConfigSchema adds the tracing schema to the compiler.
// The interface is specified instead of `jsonschema.Compiler` to allow the use of any jsonschema library fork or version.
func AddConfigSchema(c interface {
	AddResource(url string, r io.Reader) error
},
) error {
	return c.AddResource(ConfigSchemaID, bytes.NewBufferString(ConfigSchema))
}
