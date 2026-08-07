package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds static user preferences stored in ~/.dgtr/config.json.
type Config struct {
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// Credentials holds machine-managed dynamic tokens stored in ~/.dgtr/credentials.json.
type Credentials struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// defaultConfigDir returns the default directory path (~/.dgtr).
func defaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home directory: %w", err)
	}
	return filepath.Join(home, ".dgtr"), nil
}

// defaultConfigPath returns the default config file path (~/.dgtr/config.json).
func defaultConfigPath() (string, error) {
	dir, err := defaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// defaultCredentialsPath returns the default credentials file path (~/.dgtr/credentials.json).
func defaultCredentialsPath() (string, error) {
	dir, err := defaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// loadConfig reads the static config file from path. If the file does not exist, an empty Config is returned.
func loadConfig(path string) (*Config, error) {
	if path == "" {
		p, err := defaultConfigPath()
		if err != nil {
			return nil, err
		}
		path = p
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config from %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config from %s: %w", path, err)
	}
	return &cfg, nil
}

// saveConfig writes the config file to path with 0600 file permissions and creates parent dir (0700) if needed.
func saveConfig(path string, cfg *Config) error {
	if path == "" {
		p, err := defaultConfigPath()
		if err != nil {
			return err
		}
		path = p
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config to %s: %w", path, err)
	}
	return nil
}

// loadCredentials reads the dynamic credentials file from path. If the file does not exist, empty Credentials is returned.
func loadCredentials(path string) (*Credentials, error) {
	if path == "" {
		p, err := defaultCredentialsPath()
		if err != nil {
			return nil, err
		}
		path = p
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Credentials{}, nil
		}
		return nil, fmt.Errorf("read credentials from %s: %w", path, err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("parse credentials from %s: %w", path, err)
	}
	return &creds, nil
}

// saveCredentials writes the credentials file to path with 0600 file permissions.
func saveCredentials(path string, creds *Credentials) error {
	if path == "" {
		p, err := defaultCredentialsPath()
		if err != nil {
			return err
		}
		path = p
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create credentials directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write credentials to %s: %w", path, err)
	}
	return nil
}
