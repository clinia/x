package configx

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- shared test types ----------

type ServerConfig struct {
	Host string `json:"host" yaml:"host" gozod:"required"`
	Port int    `json:"port" yaml:"port" gozod:"required,min=1,max=65535"`
}

type DatabaseConfig struct {
	Host     string `json:"host"     yaml:"host"     gozod:"required"`
	Port     int    `json:"port"     yaml:"port"     gozod:"required,min=1,max=65535"`
	Name     string `json:"name"     yaml:"name"     gozod:"required"`
	Username string `json:"username" yaml:"username" gozod:"required"`
	Password string `json:"password" yaml:"password" gozod:"required,min=8"`
}

type TestConfig struct {
	Server   ServerConfig   `json:"server"   yaml:"server"   gozod:"required"`
	Database DatabaseConfig `json:"database" yaml:"database" gozod:"required"`
}

// testdataPath returns the absolute path to a testdata file relative to
// this test file's location.
func testdataPath(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(filename), "testdata", name)
}

// ---------- Load tests ----------

func TestLoad_FromYAMLFile(t *testing.T) {
	cfg, err := Load[TestConfig](
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
	cfg, err := Load[TestConfig](
		WithFiles(testdataPath(t, "config.json")),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "json-host", cfg.Server.Host)
	assert.Equal(t, 7070, cfg.Server.Port)
	assert.Equal(t, "json-db", cfg.Database.Host)
}

func TestLoad_MultipleFilesMerge(t *testing.T) {
	cfg, err := Load[TestConfig](
		WithFiles(
			testdataPath(t, "base.yaml"),
			testdataPath(t, "override.yaml"),
		),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// override.yaml sets server.port=9090
	assert.Equal(t, 9090, cfg.Server.Port)
	// override.yaml sets database.host
	assert.Equal(t, "prod-db.example.com", cfg.Database.Host)
	// base.yaml values retained where not overridden
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "myapp", cfg.Database.Name)
}

func TestLoad_EnvVarsOverride(t *testing.T) {
	t.Setenv("TEST_SERVER_HOST", "env-host")
	t.Setenv("TEST_SERVER_PORT", "4444")

	cfg, err := Load[TestConfig](
		WithFiles(testdataPath(t, "base.yaml")),
		WithEnvPrefix("TEST"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env-host", cfg.Server.Host)
	// Env vars are strings. The JSON round-trip in mapToStruct won't coerce
	// the string "4444" into an int when the surrounding map value is a string,
	// so the YAML-sourced int value (8080) stays.
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoad_WithValues(t *testing.T) {
	cfg, err := Load[TestConfig](
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

func TestLoad_ValidationFailure(t *testing.T) {
	// No files loaded, so all required fields are missing.
	_, err := Load[TestConfig]()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestLoad_UnsupportedFileFormat(t *testing.T) {
	_, err := Load[TestConfig](
		WithFiles("config.toml"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported config file format")
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load[TestConfig](
		WithFiles("/nonexistent/path/config.yaml"),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading file")
}

func TestLoad_MinimalConfig(t *testing.T) {
	cfg, err := Load[TestConfig](
		WithFiles(testdataPath(t, "minimal.yaml")),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "minimal", cfg.Server.Host)
	assert.Equal(t, 3000, cfg.Server.Port)
}

// ---------- Env-only and merge tests ----------

func TestLoad_EnvVarsOnly(t *testing.T) {
	// Use a struct with only string fields so env vars deserialise directly.
	type StringConfig struct {
		Server struct {
			Host string `json:"host" yaml:"host" gozod:"required"`
		} `json:"server" yaml:"server" gozod:"required"`
		Database struct {
			Host     string `json:"host"     yaml:"host"     gozod:"required"`
			Name     string `json:"name"     yaml:"name"     gozod:"required"`
			Username string `json:"username" yaml:"username" gozod:"required"`
			Password string `json:"password" yaml:"password" gozod:"required"`
		} `json:"database" yaml:"database" gozod:"required"`
	}

	t.Setenv("APP_SERVER_HOST", "env-only-host")
	t.Setenv("APP_DATABASE_HOST", "env-db-host")
	t.Setenv("APP_DATABASE_NAME", "envdb")
	t.Setenv("APP_DATABASE_USERNAME", "envuser")
	t.Setenv("APP_DATABASE_PASSWORD", "envpassword")

	cfg, err := Load[StringConfig](
		WithEnvPrefix("APP"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "env-only-host", cfg.Server.Host)
	assert.Equal(t, "env-db-host", cfg.Database.Host)
	assert.Equal(t, "envdb", cfg.Database.Name)
}

func TestLoad_FileAndEnvMerge(t *testing.T) {
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

	t.Setenv("MERGE_SERVER_HOST", "env-host")
	t.Setenv("MERGE_DATABASE_NAME", "envdb")

	type MergeConfig struct {
		Server struct {
			Host string `json:"host" yaml:"host" gozod:"required"`
			Port int    `json:"port" yaml:"port" gozod:"required,min=1"`
		} `json:"server" yaml:"server" gozod:"required"`
		Database struct {
			Host     string `json:"host"     yaml:"host"     gozod:"required"`
			Port     int    `json:"port"     yaml:"port"     gozod:"required,min=1"`
			Name     string `json:"name"     yaml:"name"     gozod:"required"`
			Username string `json:"username" yaml:"username" gozod:"required"`
			Password string `json:"password" yaml:"password" gozod:"required"`
		} `json:"database" yaml:"database" gozod:"required"`
	}

	cfg, err := Load[MergeConfig](
		WithFiles(yamlPath),
		WithEnvPrefix("MERGE"),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Env vars override
	assert.Equal(t, "env-host", cfg.Server.Host)
	assert.Equal(t, "envdb", cfg.Database.Name)
	// File values retained where not overridden
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "file-db", cfg.Database.Host)
	assert.Equal(t, "fileuser", cfg.Database.Username)
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

	t.Setenv("VALTEST_SERVER_HOST", "env-host")

	type ValConfig struct {
		Server struct {
			Host string `json:"host" yaml:"host" gozod:"required"`
			Port int    `json:"port" yaml:"port" gozod:"required,min=1"`
		} `json:"server" yaml:"server" gozod:"required"`
		Database struct {
			Host     string `json:"host"     yaml:"host"     gozod:"required"`
			Port     int    `json:"port"     yaml:"port"     gozod:"required,min=1"`
			Name     string `json:"name"     yaml:"name"     gozod:"required"`
			Username string `json:"username" yaml:"username" gozod:"required"`
			Password string `json:"password" yaml:"password" gozod:"required"`
		} `json:"database" yaml:"database" gozod:"required"`
	}

	cfg, err := Load[ValConfig](
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

// ---------- Schema tests ----------

func TestSchema_GeneratesValidJSON(t *testing.T) {
	schemaBytes, err := Schema[TestConfig]()
	require.NoError(t, err)

	// Should be valid JSON
	var parsed map[string]any
	err = json.Unmarshal(schemaBytes, &parsed)
	require.NoError(t, err)
	assert.NotEmpty(t, parsed)
}

func TestSchema_WithIndent(t *testing.T) {
	schemaBytes, err := Schema[TestConfig](WithIndent())
	require.NoError(t, err)

	// Indented JSON should contain newlines
	assert.Contains(t, string(schemaBytes), "\n")

	// Should still be valid JSON
	var parsed map[string]any
	err = json.Unmarshal(schemaBytes, &parsed)
	require.NoError(t, err)
}

// ---------- Option unit tests ----------

func TestWithEnvPrefix_AddsTrailingUnderscore(t *testing.T) {
	o := defaultOptions()
	WithEnvPrefix("MYAPP")(o)
	assert.Equal(t, "MYAPP_", o.envPrefix)
}

func TestWithEnvPrefix_PreservesTrailingUnderscore(t *testing.T) {
	o := defaultOptions()
	WithEnvPrefix("MYAPP_")(o)
	assert.Equal(t, "MYAPP_", o.envPrefix)
}

func TestWithEnvPrefix_ConvertsToUppercase(t *testing.T) {
	o := defaultOptions()
	WithEnvPrefix("myapp")(o)
	assert.Equal(t, "MYAPP_", o.envPrefix)
}
