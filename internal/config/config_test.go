package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("expected BaseURL %q, got %q", DefaultBaseURL, cfg.BaseURL)
	}
	if cfg.HasTokens() {
		t.Fatalf("expected HasTokens() to be false on fresh config")
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	// Create a temporary directory for isolation
	tempHome := t.TempDir()
	t.Setenv("USERPROFILE", tempHome) // Windows
	t.Setenv("HOME", tempHome)        // Unix/Linux

	cfg := &Config{
		BaseURL:      "http://test.conclave.local:8000",
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		User: UserProfile{
			ID:    "user-uuid-1234",
			Email: "alex@conclave.local",
			Name:  "Alex Mercer",
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify the file was created in the expected path
	expectedFile := filepath.Join(tempHome, ".conclave", "config.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("expected file to exist at %s, but not found", expectedFile)
	}

	// Load the config back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.BaseURL != cfg.BaseURL {
		t.Errorf("expected BaseURL %q, got %q", cfg.BaseURL, loaded.BaseURL)
	}
	if loaded.AccessToken != cfg.AccessToken {
		t.Errorf("expected AccessToken %q, got %q", cfg.AccessToken, loaded.AccessToken)
	}
	if loaded.User.Name != cfg.User.Name {
		t.Errorf("expected User.Name %q, got %q", cfg.User.Name, loaded.User.Name)
	}
	if !loaded.HasTokens() {
		t.Errorf("expected HasTokens() to be true")
	}

	// Test ClearAuth
	if err := loaded.ClearAuth(); err != nil {
		t.Fatalf("failed to clear auth: %v", err)
	}
	if loaded.HasTokens() {
		t.Errorf("expected HasTokens() to be false after ClearAuth")
	}
	if loaded.AccessToken != "" || loaded.RefreshToken != "" {
		t.Errorf("expected tokens to be empty after ClearAuth")
	}
}
