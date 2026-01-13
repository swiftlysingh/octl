package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "octl-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("Load returns empty config when file doesn't exist", func(t *testing.T) {
		cfg, err := Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}
		if cfg.ClientID != "" {
			t.Errorf("Load() ClientID = %q, want empty", cfg.ClientID)
		}
	})

	t.Run("Save and Load roundtrip", func(t *testing.T) {
		cfg := &Config{ClientID: "test-client-id"}
		if err := Save(cfg); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if loaded.ClientID != cfg.ClientID {
			t.Errorf("Load() ClientID = %q, want %q", loaded.ClientID, cfg.ClientID)
		}
	})

	t.Run("SetClientID saves to config", func(t *testing.T) {
		if err := SetClientID("new-client-id"); err != nil {
			t.Fatalf("SetClientID() error = %v", err)
		}

		loaded, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if loaded.ClientID != "new-client-id" {
			t.Errorf("ClientID = %q, want %q", loaded.ClientID, "new-client-id")
		}
	})

	t.Run("GetClientID returns from config", func(t *testing.T) {
		// Clear environment
		os.Unsetenv("OCTL_CLIENT_ID")

		if err := SetClientID("config-client-id"); err != nil {
			t.Fatalf("SetClientID() error = %v", err)
		}

		if got := GetClientID(); got != "config-client-id" {
			t.Errorf("GetClientID() = %q, want %q", got, "config-client-id")
		}
	})

	t.Run("GetClientID prefers environment variable", func(t *testing.T) {
		os.Setenv("OCTL_CLIENT_ID", "env-client-id")
		defer os.Unsetenv("OCTL_CLIENT_ID")

		if got := GetClientID(); got != "env-client-id" {
			t.Errorf("GetClientID() = %q, want %q", got, "env-client-id")
		}
	})

	t.Run("ConfigDir returns correct path", func(t *testing.T) {
		dir, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		expected := filepath.Join(tmpDir, ".config", "octl")
		if dir != expected {
			t.Errorf("ConfigDir() = %q, want %q", dir, expected)
		}
	})
}
