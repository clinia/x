package logrusx

import (
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

// NewLogFields converts a variadic list of attribute.KeyValue pairs into a logrus.Fields map.
// Each attribute.KeyValue's key is transformed by replacing all dots ('.') with double underscores ('__').
// The resulting map can be used with logrus for structured logging.
//
// Note: We are mainly using this since loki labels cannot contain dots. This also enables us to reuse the same attributes for both tracing and logging.
func NewLogFields(kvs ...attribute.KeyValue) logrus.Fields {
	f := logrus.Fields{}
	for _, kv := range kvs {
		k := strings.ReplaceAll(string(kv.Key), ".", "__")
		f[k] = kv.Value.AsInterface()
	}

	return f
}
