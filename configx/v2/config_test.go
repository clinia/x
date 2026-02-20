package configx

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/clinia/x/errorx"
	"github.com/kaptinlin/gozod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- shared test types & schemas ----------

type SimpleConfig struct {
	Host  string `json:"host"  yaml:"host"`
	Port  int    `json:"port"  yaml:"port"`
	Debug bool   `json:"debug" yaml:"debug"`
}

type ServerConfig struct {
	Host string `json:"host" yaml:"host"`
	Port int    `json:"port" yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host"     yaml:"host"`
	Port     int    `json:"port"     yaml:"port"`
	Name     string `json:"name"     yaml:"name"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type FullConfig struct {
	Server   ServerConfig   `json:"server"   yaml:"server"`
	Database DatabaseConfig `json:"database" yaml:"database"`
}

func simpleSchema() gozod.ZodType[SimpleConfig] {
	return gozod.Struct[SimpleConfig](gozod.StructSchema{
		"host":  gozod.String().Default("localhost"),
		"port":  gozod.Int().Default(8080),
		"debug": gozod.Bool().Default(false),
	})
}

func fullSchema() gozod.ZodType[FullConfig] {
	return gozod.Struct[FullConfig](gozod.StructSchema{
		"server": gozod.Struct[ServerConfig](gozod.StructSchema{
			"host": gozod.String().Default("localhost"),
			"port": gozod.Int().Default(8080),
		}),
		"database": gozod.Struct[DatabaseConfig](gozod.StructSchema{
			"host":     gozod.String(),
			"port":     gozod.Int().Default(5432),
			"name":     gozod.String(),
			"username": gozod.String(),
			"password": gozod.String(),
		}),
	})
}

// testdataPath returns the absolute path to a testdata file relative to
// this test file's location.
func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(filename), "testdata", name)
}

// ================== Load tests ==================

func TestLoad_FromYAMLFile(t *testing.T) {
	cfg, err := Load(fullSchema(),
		WithFiles(testdataPath(t, "base.yaml")),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "myapp", cfg.Database.Name)
	assert.Equal(t, "admin", cfg.Database.Username)
	assert.Equal(t, "secret123", cfg.Database.Password)
}

func TestLoad_FromJSONFile(t *testing.T) {
	cfg, err := Load(fullSchema(),
		WithFiles(testdataPath(t, "config.json")),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "json-host", cfg.Server.Host)
	assert.Equal(t, 7070, cfg.Server.Port)
	assert.Equal(t, "json-db", cfg.Database.Host)
}

func TestLoad_MultipleFilesMerge(t *testing.T) {
	cfg, err := Load(fullSchema(),
		WithFiles(
			testdataPath(t, "base.yaml"),
			testdataPath(t, "override.yaml"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// override.yaml overrides server.port=9090, database.host
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "prod-db.example.com", cfg.Database.Host)
	// base.yaml values retained where not overridden
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "myapp", cfg.Database.Name)
}

func TestLoad_SimpleSchema(t *testing.T) {
	cfg, err := Load(simpleSchema(),
		WithFiles(testdataPath(t, "simple.yaml")),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "simple-host", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.True(t, cfg.Debug)
}

func TestLoad_Defaults(t *testing.T) {
	// No sources at all — should apply defaults from schema.
	cfg, err := Load(simpleSchema())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.False(t, cfg.Debug)
}

// ================== Env var tests ==================

func TestLoad_EnvVarsOverride(t *testing.T) {
	t.Setenv("TEST_SERVER__HOST", "env-host")
	t.Setenv("TEST_SERVER__PORT", "4444")

	cfg, err := Load(fullSchema(),
		WithFiles(testdataPath(t, "base.yaml")),
		WithEnvPrefix("TEST"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env-host", cfg.Server.Host)
	assert.Equal(t, 4444, cfg.Server.Port) // coerced from string to int
}

func TestLoad_EnvVarsDoubleUnderscore(t *testing.T) {
	// Test that __ maps to nesting and single _ is preserved.
	type DBConfig struct {
		MaxConns int `json:"max_conns" yaml:"max_conns"`
	}
	type AppConfig struct {
		Database DBConfig `json:"database" yaml:"database"`
	}
	schema := gozod.Struct[AppConfig](gozod.StructSchema{
		"database": gozod.Struct[DBConfig](gozod.StructSchema{
			"max_conns": gozod.Int().Default(10),
		}),
	})

	t.Setenv("APP_DATABASE__MAX_CONNS", "25")

	cfg, err := Load(schema, WithEnvPrefix("APP"))
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 25, cfg.Database.MaxConns)
}

func TestLoad_EnvVarsOnly(t *testing.T) {
	t.Setenv("APP_HOST", "env-only-host")
	t.Setenv("APP_PORT", "3000")
	t.Setenv("APP_DEBUG", "true")

	// For flat config, env vars use single _ (no nesting needed)
	schema := gozod.Struct[SimpleConfig](gozod.StructSchema{
		"host":  gozod.String(),
		"port":  gozod.Int(),
		"debug": gozod.Bool(),
	})

	cfg, err := Load(schema, WithEnvPrefix("APP"))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env-only-host", cfg.Host)
	assert.Equal(t, 3000, cfg.Port)
	assert.True(t, cfg.Debug)
}

// ================== WithValues tests ==================

func TestLoad_WithValues(t *testing.T) {
	cfg, err := Load(fullSchema(),
		WithFiles(testdataPath(t, "base.yaml")),
		WithValues(map[string]any{
			"server.host": "override-host",
			"server.port": 1234,
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "override-host", cfg.Server.Host)
	assert.Equal(t, 1234, cfg.Server.Port)
}

func TestLoad_ValuesOverrideEverything(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(yamlPath, []byte(`
server:
  host: "file-host"
  port: 8080
database:
  host: "file-db"
  port: 5432
  name: "filedb"
  username: "fileuser"
  password: "filepasswd"
`), 0o644)
	require.NoError(t, err)

	t.Setenv("VALTEST_SERVER__HOST", "env-host")

	cfg, err := Load(fullSchema(),
		WithFiles(yamlPath),
		WithEnvPrefix("VALTEST"),
		WithValues(map[string]any{
			"server.host": "values-host",
		}),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// WithValues overrides both file and env
	assert.Equal(t, "values-host", cfg.Server.Host)
}

// ================== Error cases ==================

func TestLoad_UnsupportedFileFormat(t *testing.T) {
	_, err := Load(simpleSchema(),
		WithFiles("config.toml"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config file format")
	assert.True(t, errorx.IsInvalidArgumentError(err))
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(simpleSchema(),
		WithFiles("/nonexistent/path/config.yaml"),
	)
	require.Error(t, err)
	assert.True(t, errorx.IsInvalidArgumentError(err))
	assert.Contains(t, err.Error(), "/nonexistent/path/config.yaml")
}

func TestLoad_ValidationFailure(t *testing.T) {
	// Schema with required field and no default, no sources.
	type RequiredConfig struct {
		Name string `json:"name" yaml:"name"`
	}
	schema := gozod.Struct[RequiredConfig](gozod.StructSchema{
		"name": gozod.String().Min(1),
	})

	_, err := Load(schema)
	require.Error(t, err)
	assert.True(t, errorx.IsInvalidArgumentError(err))
	assert.Contains(t, err.Error(), "config validation failed")
}

// ================== Env transform unit tests ==================

func TestEnvTransformFunc_StripsPrefixAndNests(t *testing.T) {
	transform := envTransformFunc("APP_", ".")

	key, val := transform("APP_SERVER__HOST", "localhost")
	assert.Equal(t, "server.host", key)
	assert.Equal(t, "localhost", val)
}

func TestEnvTransformFunc_PreservesSingleUnderscore(t *testing.T) {
	transform := envTransformFunc("APP_", ".")

	key, _ := transform("APP_DATABASE__MAX_CONNS", "10")
	assert.Equal(t, "database.max_conns", key)
}

func TestEnvTransformFunc_FlatKey(t *testing.T) {
	transform := envTransformFunc("APP_", ".")

	key, _ := transform("APP_DEBUG", "true")
	assert.Equal(t, "debug", key)
}

// ================== Coerce value tests ==================

func TestCoerceValue_Int(t *testing.T) {
	assert.Equal(t, 42, coerceValue("42"))
	assert.Equal(t, -1, coerceValue("-1"))
	assert.Equal(t, 0, coerceValue("0"))
}

func TestCoerceValue_Float(t *testing.T) {
	assert.Equal(t, 3.14, coerceValue("3.14"))
	assert.Equal(t, 0.5, coerceValue("0.5"))
}

func TestCoerceValue_Bool(t *testing.T) {
	assert.Equal(t, true, coerceValue("true"))
	assert.Equal(t, false, coerceValue("false"))
	assert.Equal(t, true, coerceValue("TRUE"))
}

func TestCoerceValue_JSON(t *testing.T) {
	result := coerceValue(`{"key":"value"}`)
	m, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", m["key"])

	result = coerceValue(`[1,2,3]`)
	arr, ok := result.([]any)
	require.True(t, ok)
	assert.Len(t, arr, 3)
}

func TestCoerceValue_String(t *testing.T) {
	assert.Equal(t, "hello", coerceValue("hello"))
	assert.Equal(t, "", coerceValue(""))
	assert.Equal(t, "192.168.1.1", coerceValue("192.168.1.1"))
}
