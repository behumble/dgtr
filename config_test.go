package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigAndCredentialsLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")
	credsPath := filepath.Join(tmpDir, "credentials.json")

	// Load non-existent config should return empty Config without error
	cfg, err := loadConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadConfig non-existent returned error: %v", err)
	}
	if cfg.ClientID != "" || cfg.ClientSecret != "" {
		t.Errorf("expected empty config, got %+v", cfg)
	}

	// Save static config
	cfg.ClientID = "test_client_id_456"
	cfg.ClientSecret = "test_client_secret_789"
	if err := saveConfig(cfgPath, cfg); err != nil {
		t.Fatalf("saveConfig returned error: %v", err)
	}

	// Reload config and verify contents
	loadedCfg, err := loadConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}
	if loadedCfg.ClientID != "test_client_id_456" || loadedCfg.ClientSecret != "test_client_secret_789" {
		t.Errorf("unexpected loaded config: %+v", loadedCfg)
	}

	// Load non-existent credentials
	creds, err := loadCredentials(credsPath)
	if err != nil {
		t.Fatalf("loadCredentials non-existent returned error: %v", err)
	}
	if creds.RefreshToken != "" {
		t.Errorf("expected empty refresh_token, got %q", creds.RefreshToken)
	}

	// Save dynamic credentials
	creds.RefreshToken = "test_refresh_token_123"
	if err := saveCredentials(credsPath, creds); err != nil {
		t.Fatalf("saveCredentials returned error: %v", err)
	}

	// Reload credentials and verify
	loadedCreds, err := loadCredentials(credsPath)
	if err != nil {
		t.Fatalf("loadCredentials returned error: %v", err)
	}
	if loadedCreds.RefreshToken != "test_refresh_token_123" {
		t.Errorf("expected 'test_refresh_token_123', got %q", loadedCreds.RefreshToken)
	}

	// Verify file permissions (0600)
	for _, p := range []string{cfgPath, credsPath} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat error for %s: %v", p, err)
		}
		if mode := info.Mode().Perm(); mode != 0o600 {
			t.Errorf("expected file mode 0600 for %s, got %o", p, mode)
		}
	}
}
