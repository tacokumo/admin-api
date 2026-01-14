package auth

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

// TokenStorage handles persistent storage of authentication tokens
type TokenStorage struct {
	configDir string
}

type StoredToken struct {
	BearerToken string             `json:"bearer_token"`
	User        *AuthenticatedUser `json:"user,omitempty"`
}

// NewTokenStorage creates a new token storage instance
func NewTokenStorage() (*TokenStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user home directory")
	}

	configDir := filepath.Join(homeDir, ".admin-api")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, errors.Wrap(err, "failed to create config directory")
	}

	return &TokenStorage{
		configDir: configDir,
	}, nil
}

// GetTokenPath returns the path where the token is stored
func (ts *TokenStorage) GetTokenPath() string {
	return filepath.Join(ts.configDir, "token.json")
}

// SaveToken saves a token to persistent storage
func (ts *TokenStorage) SaveToken(token string, user *AuthenticatedUser) error {
	storedToken := StoredToken{
		BearerToken: token,
		User:        user,
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal token data")
	}

	tokenPath := ts.GetTokenPath()
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		return errors.Wrap(err, "failed to write token file")
	}

	return nil
}

// LoadToken loads a token from persistent storage
func (ts *TokenStorage) LoadToken() (*StoredToken, error) {
	tokenPath := ts.GetTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token file exists
		}
		return nil, errors.Wrap(err, "failed to read token file")
	}

	var storedToken StoredToken
	if err := json.Unmarshal(data, &storedToken); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal token data")
	}

	return &storedToken, nil
}

// RemoveToken removes the stored token
func (ts *TokenStorage) RemoveToken() error {
	tokenPath := ts.GetTokenPath()
	if err := os.Remove(tokenPath); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove token file")
	}
	return nil
}

// HasToken checks if a token is stored
func (ts *TokenStorage) HasToken() bool {
	tokenPath := ts.GetTokenPath()
	_, err := os.Stat(tokenPath)
	return err == nil
}
