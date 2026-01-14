package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFromYAML(t *testing.T) {
	t.Parallel()

	t.Run("loads valid YAML configuration", func(t *testing.T) {
		t.Parallel()

		// Create a temporary YAML file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `
addr: "0.0.0.0"
port: "8080"
admin_db:
  host: "localhost"
  port: 5432
  user: "testuser"
  password: "testpass"
  db_name: "testdb"
  initial_conn_retry: 3
auth:
  client_id: "github-client-id"
  client_secret: "github-client-secret"
  callback_url: "http://localhost:8080/callback"
  frontend_url: "http://localhost:3000"
  allowed_orgs:
    - "org1"
    - "org2"
  session_ttl: "24h"
redis:
  host: "localhost"
  port: 6379
  password: "redispass"
  db: 0
cors:
  allow_origins: "*"
  allow_methods: "GET,POST,PUT,DELETE"
  allow_headers: "Content-Type,Authorization"
  expose_headers: "X-Total-Count"
  allow_credentials: true
  max_age: 3600
tls:
  enabled: true
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
`

		if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		cfg, err := LoadFromYAML(configFile)
		if err != nil {
			t.Fatalf("LoadFromYAML() failed: %v", err)
		}

		// Verify loaded values
		if cfg.Addr != "0.0.0.0" {
			t.Errorf("Addr mismatch: got %s, want 0.0.0.0", cfg.Addr)
		}

		if cfg.Port != "8080" {
			t.Errorf("Port mismatch: got %s, want 8080", cfg.Port)
		}

		if cfg.AdminDBConfig.Host != "localhost" {
			t.Errorf("AdminDB Host mismatch: got %s, want localhost", cfg.AdminDBConfig.Host)
		}

		if cfg.AdminDBConfig.Port != 5432 {
			t.Errorf("AdminDB Port mismatch: got %d, want 5432", cfg.AdminDBConfig.Port)
		}

		if cfg.Auth.GitHubClientID != "github-client-id" {
			t.Errorf("GitHub Client ID mismatch: got %s, want github-client-id", cfg.Auth.GitHubClientID)
		}

		expectedTTL := 24 * time.Hour
		if cfg.Auth.SessionTTL != expectedTTL {
			t.Errorf("Session TTL mismatch: got %v, want %v", cfg.Auth.SessionTTL, expectedTTL)
		}

		expectedOrgs := []string{"org1", "org2"}
		if len(cfg.Auth.AllowedOrgs) != len(expectedOrgs) {
			t.Errorf("AllowedOrgs length mismatch: got %d, want %d", len(cfg.Auth.AllowedOrgs), len(expectedOrgs))
		}

		for i, org := range expectedOrgs {
			if cfg.Auth.AllowedOrgs[i] != org {
				t.Errorf("AllowedOrgs[%d] mismatch: got %s, want %s", i, cfg.Auth.AllowedOrgs[i], org)
			}
		}

		if cfg.Redis.Port != 6379 {
			t.Errorf("Redis Port mismatch: got %d, want 6379", cfg.Redis.Port)
		}

		if !cfg.CORS.AllowCredentials {
			t.Errorf("CORS AllowCredentials mismatch: got %t, want true", cfg.CORS.AllowCredentials)
		}

		if !cfg.TLS.Enabled {
			t.Errorf("TLS Enabled mismatch: got %t, want true", cfg.TLS.Enabled)
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		t.Parallel()

		_, err := LoadFromYAML("/non/existent/path.yaml")
		if err == nil {
			t.Error("LoadFromYAML() should return error for non-existent file")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "invalid.yaml")

		invalidYAML := `
addr: "0.0.0.0"
port: [invalid yaml structure
`

		if err := os.WriteFile(configFile, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to write invalid config file: %v", err)
		}

		_, err := LoadFromYAML(configFile)
		if err == nil {
			t.Error("LoadFromYAML() should return error for invalid YAML")
		}
	})
}

func TestLoadFromYAMLWithEnvOverride(t *testing.T) {
	t.Parallel()

	t.Run("overrides YAML with environment variables", func(t *testing.T) {
		// Create a temporary YAML file
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		yamlContent := `
addr: "0.0.0.0"
port: "8080"
admin_db:
  host: "localhost"
  port: 5432
auth:
  client_id: "yaml-client-id"
  session_ttl: "1h"
`

		if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}

		// Set environment variables to override YAML values
		originalPort := os.Getenv("PORT")
		originalClientID := os.Getenv("GITHUB_CLIENT_ID")
		originalTTL := os.Getenv("SESSION_TTL")

		defer func() {
			// Restore original values
			if err := os.Setenv("PORT", originalPort); err != nil {
				t.Errorf("Failed to restore PORT: %v", err)
			}
			if err := os.Setenv("GITHUB_CLIENT_ID", originalClientID); err != nil {
				t.Errorf("Failed to restore GITHUB_CLIENT_ID: %v", err)
			}
			if err := os.Setenv("SESSION_TTL", originalTTL); err != nil {
				t.Errorf("Failed to restore SESSION_TTL: %v", err)
			}
		}()

		if err := os.Setenv("PORT", "9090"); err != nil {
			t.Fatalf("Failed to set PORT: %v", err)
		}
		if err := os.Setenv("GITHUB_CLIENT_ID", "env-client-id"); err != nil {
			t.Fatalf("Failed to set GITHUB_CLIENT_ID: %v", err)
		}
		if err := os.Setenv("SESSION_TTL", "2h"); err != nil {
			t.Fatalf("Failed to set SESSION_TTL: %v", err)
		}

		cfg, err := LoadFromYAMLWithEnvOverride(configFile)
		if err != nil {
			t.Fatalf("LoadFromYAMLWithEnvOverride() failed: %v", err)
		}

		// Verify that environment variables override YAML values
		if cfg.Port != "9090" {
			t.Errorf("Port should be overridden by env: got %s, want 9090", cfg.Port)
		}

		if cfg.Auth.GitHubClientID != "env-client-id" {
			t.Errorf("GitHub Client ID should be overridden by env: got %s, want env-client-id", cfg.Auth.GitHubClientID)
		}

		expectedTTL := 2 * time.Hour
		if cfg.Auth.SessionTTL != expectedTTL {
			t.Errorf("Session TTL should be overridden by env: got %v, want %v", cfg.Auth.SessionTTL, expectedTTL)
		}

		// Verify that non-overridden values remain from YAML
		if cfg.Addr != "0.0.0.0" {
			t.Errorf("Addr should remain from YAML: got %s, want 0.0.0.0", cfg.Addr)
		}

		if cfg.AdminDBConfig.Port != 5432 {
			t.Errorf("AdminDB Port should remain from YAML: got %d, want 5432", cfg.AdminDBConfig.Port)
		}
	})
}
