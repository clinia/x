package kgox

import (
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

// TODO: [ENG-1359] Move me to logrusx
func newLogFields(kvs ...attribute.KeyValue) logrus.Fields {
	f := logrus.Fields{}
	for _, kv := range kvs {
		k := strings.ReplaceAll(string(kv.Key), ".", "__")
		f[k] = kv.Value.AsInterface()
	}

	return f
}
