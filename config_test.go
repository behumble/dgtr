package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	// Load non-existent config should return empty Config without error
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadConfig non-existent returned error: %v", err)
	}
	if cfg.RefreshToken != "" {
		t.Errorf("expected empty refresh_token, got %q", cfg.RefreshToken)
	}

	// Save config
	cfg.ClientID = "test_client_id_456"
	cfg.ClientSecret = "test_client_secret_789"
	cfg.RefreshToken = "test_refresh_token_123"
	if err := saveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("saveConfig returned error: %v", err)
	}

	// Reload config and verify contents
	loaded, err := loadConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if loaded.ClientID != "test_client_id_456" {
		t.Errorf("expected 'test_client_id_456', got %q", loaded.ClientID)
	}
	if loaded.ClientSecret != "test_client_secret_789" {
		t.Errorf("expected 'test_client_secret_789', got %q", loaded.ClientSecret)
	}
	if loaded.RefreshToken != "test_refresh_token_123" {
		t.Errorf("expected 'test_refresh_token_123', got %q", loaded.RefreshToken)
	}

	// Verify file permissions (0600)
	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("stat config file error: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != 0o600 {
		t.Errorf("expected file mode 0600, got %o", mode)
	}
}
