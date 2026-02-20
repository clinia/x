package configx

import (
	"encoding/json"
	"strconv"
	"strings"
)

// envTransformFunc returns a koanf env TransformFunc that:
//  1. Strips the prefix from the environment variable key.
//  2. Replaces double underscores (__) with the koanf delimiter to represent nesting.
//     Single underscores are preserved as-is (for multi-word keys like max_conns).
//  3. Lowercases the entire key.
//  4. Coerces the string value to the most specific Go type (int, float64, bool, JSON).
//
// Example with prefix "APP_" and delimiter ".":
//
//	APP_SERVER__HOST        → key="server.host",   value="0.0.0.0"
//	APP_DATABASE__MAX_CONNS → key="database.max_conns", value=10  (coerced to int)
//	APP_DEBUG               → key="debug",          value=true    (coerced to bool)
func envTransformFunc(prefix, delimiter string) func(string, string) (string, any) {
	return func(key, value string) (string, any) {
		key = strings.TrimPrefix(key, prefix)
		key = strings.ReplaceAll(key, "__", delimiter)
		key = strings.ToLower(key)
		return key, coerceValue(value)
	}
}

// coerceValue attempts to parse a string environment variable value into the
// most specific Go type. Parsing is attempted in the following order:
//
//  1. int (via strconv.Atoi)
//  2. float64 (via strconv.ParseFloat)
//  3. bool (via strconv.ParseBool: "true", "false", "1", "0", etc.)
//  4. JSON object/array (if the string starts with '{' or '[')
//  5. Falls back to the original string value.
//
// This ordering ensures that "42" becomes int(42) rather than float64(42),
// and that "1"/"0" become int rather than bool.
func coerceValue(v string) any {
	// Try integer first.
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}

	// Try float.
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}

	// Try boolean.
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}

	// Try JSON for objects and arrays.
	if len(v) > 0 && (v[0] == '{' || v[0] == '[') {
		var parsed any
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			return parsed
		}
	}

	return v
}
