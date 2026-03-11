package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	if err := os.Unsetenv("MEMEX_DATABASE_PATH"); err != nil {
		t.Fatalf("Failed to unset MEMEX_DATABASE_PATH: %v", err)
	}
	if err := os.Unsetenv("MEMEX_MODE"); err != nil {
		t.Fatalf("Failed to unset MEMEX_MODE: %v", err)
	}
	if err := os.Unsetenv("MEMEX_LOG_LEVEL"); err != nil {
		t.Fatalf("Failed to unset MEMEX_LOG_LEVEL: %v", err)
	}

	cfg := Load()

	if cfg.Mode != "local" {
		t.Errorf("Expected Mode='local', got '%s'", cfg.Mode)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel='info', got '%s'", cfg.LogLevel)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home dir: %v", err)
	}
	expectedPath := filepath.Join(home, ".memex", "memex.db")
	if cfg.DatabasePath != expectedPath {
		t.Errorf("Expected DatabasePath='%s', got '%s'", expectedPath, cfg.DatabasePath)
	}
}

func TestConfigFromEnv(t *testing.T) {
	if err := os.Setenv("MEMEX_DATABASE_PATH", "/tmp/test.db"); err != nil {
		t.Fatalf("Failed to set MEMEX_DATABASE_PATH: %v", err)
	}
	if err := os.Setenv("MEMEX_MODE", "cloud"); err != nil {
		t.Fatalf("Failed to set MEMEX_MODE: %v", err)
	}
	if err := os.Setenv("MEMEX_LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("Failed to set MEMEX_LOG_LEVEL: %v", err)
	}

	cfg := Load()

	if cfg.DatabasePath != "/tmp/test.db" {
		t.Errorf("Expected DatabasePath='/tmp/test.db', got '%s'", cfg.DatabasePath)
	}

	if cfg.Mode != "cloud" {
		t.Errorf("Expected Mode='cloud', got '%s'", cfg.Mode)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel='debug', got '%s'", cfg.LogLevel)
	}

	if err := os.Unsetenv("MEMEX_DATABASE_PATH"); err != nil {
		t.Errorf("Failed to unset MEMEX_DATABASE_PATH: %v", err)
	}
	if err := os.Unsetenv("MEMEX_MODE"); err != nil {
		t.Errorf("Failed to unset MEMEX_MODE: %v", err)
	}
	if err := os.Unsetenv("MEMEX_LOG_LEVEL"); err != nil {
		t.Errorf("Failed to unset MEMEX_LOG_LEVEL: %v", err)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		config  *Config
		name    string
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "local",
				LogLevel:     "info",
			},
			wantErr: false,
		},
		{
			name: "empty database path",
			config: &Config{
				DatabasePath: "",
				Mode:         "local",
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "local",
				LogLevel:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "invalid",
				LogLevel:     "info",
			},
			wantErr: true,
		},
		{
			name: "debug log level is valid",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "local",
				LogLevel:     "debug",
			},
			wantErr: false,
		},
		{
			name: "warn log level is valid",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "local",
				LogLevel:     "warn",
			},
			wantErr: false,
		},
		{
			name: "error log level is valid",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "local",
				LogLevel:     "error",
			},
			wantErr: false,
		},
		{
			name: "cloud mode is valid",
			config: &Config{
				DatabasePath: "/tmp/test.db",
				Mode:         "cloud",
				LogLevel:     "info",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
