package config

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestParseIntEnvVal(t *testing.T) {
	t.Run("empty env", func(t *testing.T) {
		val, err := ParseIntEnvVal("NON_EXISTENT_ENV_VAR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != nil {
			t.Errorf("expected nil, got %v", *val)
		}
	})

	t.Run("valid int env", func(t *testing.T) {
		t.Setenv("TEST_INT_ENV", "42")
		val, err := ParseIntEnvVal("TEST_INT_ENV")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val == nil || *val != 42 {
			t.Errorf("expected 42, got %v", val)
		}
	})

	t.Run("invalid int env", func(t *testing.T) {
		t.Setenv("TEST_INVALID_ENV", "not-a-number")
		_, err := ParseIntEnvVal("TEST_INVALID_ENV")
		if err == nil {
			t.Error("expected error for invalid int, got nil")
		}
	})
}

func TestEnvOrStr(t *testing.T) {
	t.Run("env exists", func(t *testing.T) {
		t.Setenv("TEST_STR_ENV", "env-value")
		res := EnvOrStr("TEST_STR_ENV", "default")
		if res != "env-value" {
			t.Errorf("expected 'env-value', got '%s'", res)
		}
	})

	t.Run("env empty", func(t *testing.T) {
		res := EnvOrStr("NON_EXISTENT_STR_ENV", "default")
		if res != "default" {
			t.Errorf("expected 'default', got '%s'", res)
		}
	})
}

func TestDurationSeconds_UnmarshalJSON(t *testing.T) {
	t.Run("parse string duration 1s", func(t *testing.T) {
		var d DurationSeconds
		err := d.UnmarshalJSON([]byte(`"1s"`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d != 1 {
			t.Errorf("expected 1 second, got %d", d)
		}
	})

	t.Run("parse string duration 5m", func(t *testing.T) {
		var d DurationSeconds
		err := d.UnmarshalJSON([]byte(`"5m"`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d != 300 {
			t.Errorf("expected 300 seconds, got %d", d)
		}
	})

	t.Run("parse raw int 42", func(t *testing.T) {
		var d DurationSeconds
		err := d.UnmarshalJSON([]byte(`42`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d != 42 {
			t.Errorf("expected 42 seconds, got %d", d)
		}
	})

	t.Run("parse invalid string", func(t *testing.T) {
		var d DurationSeconds
		err := d.UnmarshalJSON([]byte(`"invalid"`))
		if err == nil {
			t.Error("expected error for invalid duration string")
		}
	})
}

type TestConfig struct {
	Addr           string          `json:"address" flag:"a"`
	ReportInterval DurationSeconds `json:"report_interval" flag:"r"`
	PollInterval   int             `json:"poll_interval" flag:"p"`
	Secret         string          `json:"-" flag:"s"`
}


func TestLoadAndMerge(t *testing.T) {	
	tmpDir := t.TempDir()
	configFilePath := filepath.Join(tmpDir, "config.json")
	jsonContent := `{
		"address": "json-host:8080",
		"report_interval": "5s",
		"poll_interval": 10
	}`
	if err := os.WriteFile(configFilePath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	t.Run("JSON values applied when no CLI flags provided", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard) 

		var cfg TestConfig
		fs.StringVar(&cfg.Addr, "a", "default:8080", "")
		fs.IntVar((*int)(&cfg.ReportInterval), "r", 10, "")
		fs.IntVar(&cfg.PollInterval, "p", 2, "")
		fs.StringVar(&cfg.Secret, "s", "", "")

		// no flags!
		err := fs.Parse([]string{})
		if err != nil {
			t.Fatalf("failed to parse flags: %v", err)
		}

		err = LoadAndMerge(fs, configFilePath, &cfg)
		if err != nil {
			t.Fatalf("unexpected LoadAndMerge error: %v", err)
		}

		// values must be from json!
		if cfg.Addr != "json-host:8080" {
			t.Errorf("expected 'json-host:8080' from JSON, got '%s'", cfg.Addr)
		}
		if cfg.ReportInterval != 5 {
			t.Errorf("expected 5 from JSON, got %d", cfg.ReportInterval)
		}
		if cfg.PollInterval != 10 {
			t.Errorf("expected 10 from JSON, got %d", cfg.PollInterval)
		}
	})

	t.Run("CLI flags override JSON values", func(t *testing.T) {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		var cfg TestConfig
		fs.StringVar(&cfg.Addr, "a", "default:8080", "")
		fs.IntVar((*int)(&cfg.ReportInterval), "r", 10, "")
		fs.IntVar(&cfg.PollInterval, "p", 2, "")
		fs.StringVar(&cfg.Secret, "s", "", "")

		// flag -a with CLI. Flags -r и -p must be from json
		err := fs.Parse([]string{"-a", "cli-host:9090"})
		if err != nil {
			t.Fatalf("failed to parse flags: %v", err)
		}

		err = LoadAndMerge(fs, configFilePath, &cfg)
		if err != nil {
			t.Fatalf("unexpected LoadAndMerge error: %v", err)
		}

		// Flag 'a' must be from CLI, rest from JSON
		if cfg.Addr != "cli-host:9090" {
			t.Errorf("expected 'cli-host:9090' from CLI, got '%s'", cfg.Addr)
		}
		if cfg.ReportInterval != 5 {
			t.Errorf("expected 5 from JSON, got %d", cfg.ReportInterval)
		}
		if cfg.PollInterval != 10 {
			t.Errorf("expected 10 from JSON, got %d", cfg.PollInterval)
		}
	})

	t.Run("CONFIG environment variable overrides config path parameter", func(t *testing.T) {
		// file for ENV test
		envConfigPath := filepath.Join(tmpDir, "env_config.json")
		envJSONContent := `{"address": "env-file-host:1111"}`
		if err := os.WriteFile(envConfigPath, []byte(envJSONContent), 0644); err != nil {
			t.Fatalf("failed to create temp env config file: %v", err)
		}

		t.Setenv("CONFIG", envConfigPath)

		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.SetOutput(io.Discard)

		var cfg TestConfig
		fs.StringVar(&cfg.Addr, "a", "default:8080", "")

		err := fs.Parse([]string{})
		if err != nil {
			t.Fatalf("failed to parse flags: %v", err)
		}

		err = LoadAndMerge(fs, configFilePath, &cfg)
		if err != nil {
			t.Fatalf("unexpected LoadAndMerge error: %v", err)
		}

		if cfg.Addr != "env-file-host:1111" {
			t.Errorf("expected 'env-file-host:1111' from CONFIG env file, got '%s'", cfg.Addr)
		}
	})
}