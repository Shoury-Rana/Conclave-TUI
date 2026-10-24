package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultBaseURL is the fallback backend address for local development.
// Conclave uses root-domain APIs (e.g. localhost:8000/auth/...) for global auth,
// and subdomain-based routing (e.g. <tenant>.lvh.me:8000) for workspace-specific routes.
const DefaultBaseURL = "http://127.0.0.1:8000"

// UserProfile stores basic identity information about the logged-in user.
type UserProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Config holds the persistent settings and session tokens for Conclave TUI.
// It is serialized to JSON and stored in the user's home directory (~/.conclave/config.json).
type Config struct {
	BaseURL      string      `json:"base_url"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserProfile `json:"user"`
	ActiveTenant string      `json:"active_tenant,omitempty"`
}

// DefaultConfig returns a new Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		BaseURL: DefaultBaseURL,
	}
}

// configFilePath resolves the full path to the configuration file:
// On Windows: C:\Users\<Username>\.conclave\config.json
// On Linux/macOS: /home/<user>/.conclave/config.json
func configFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".conclave", "config.json"), nil
}

// Load reads the configuration from ~/.conclave/config.json.
// If the file does not exist yet, it returns DefaultConfig() without an error.
func Load() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return DefaultConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// First-time run: Return defaults gracefully.
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file at %s: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Ensure BaseURL is never left empty.
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	return cfg, nil
}

// Save writes the current configuration to ~/.conclave/config.json.
// It automatically creates the parent directory (~/.conclave) if it doesn't exist.
func (c *Config) Save() error {
	path, err := configFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", path, err)
	}

	return nil
}

// ClearAuth removes session tokens and user profile information (used on Logout).
func (c *Config) ClearAuth() error {
	c.AccessToken = ""
	c.RefreshToken = ""
	c.User = UserProfile{}
	c.ActiveTenant = ""
	return c.Save()
}

// HasTokens returns true if both access and refresh tokens are stored.
func (c *Config) HasTokens() bool {
	return c.AccessToken != "" && c.RefreshToken != ""
}
