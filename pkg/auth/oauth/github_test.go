package oauth

import (
	"testing"

	"golang.org/x/oauth2/github"
)

func TestNewGitHubClient(t *testing.T) {
	t.Parallel()

	t.Run("creates client with correct configuration", func(t *testing.T) {
		t.Parallel()

		clientID := "test-client-id"
		clientSecret := "test-client-secret"
		callbackURL := "http://localhost:8080/callback"
		allowedOrgs := []string{"org1", "org2"}

		client := NewGitHubClient(clientID, clientSecret, callbackURL, allowedOrgs)

		if client == nil {
			t.Fatal("NewGitHubClient() returned nil")
		}

		// Check OAuth2 config
		if client.config.ClientID != clientID {
			t.Errorf("ClientID mismatch: got %s, want %s", client.config.ClientID, clientID)
		}

		if client.config.ClientSecret != clientSecret {
			t.Errorf("ClientSecret mismatch: got %s, want %s", client.config.ClientSecret, clientSecret)
		}

		if client.config.RedirectURL != callbackURL {
			t.Errorf("RedirectURL mismatch: got %s, want %s", client.config.RedirectURL, callbackURL)
		}

		expectedScopes := []string{"user:email", "read:org"}
		if len(client.config.Scopes) != len(expectedScopes) {
			t.Errorf("Scopes length mismatch: got %d, want %d", len(client.config.Scopes), len(expectedScopes))
		}

		for i, scope := range expectedScopes {
			if client.config.Scopes[i] != scope {
				t.Errorf("Scope[%d] mismatch: got %s, want %s", i, client.config.Scopes[i], scope)
			}
		}

		if client.config.Endpoint != github.Endpoint {
			t.Errorf("Endpoint mismatch: got %v, want %v", client.config.Endpoint, github.Endpoint)
		}

		// Check allowed orgs
		if len(client.allowedOrgs) != len(allowedOrgs) {
			t.Errorf("AllowedOrgs length mismatch: got %d, want %d", len(client.allowedOrgs), len(allowedOrgs))
		}

		for i, org := range allowedOrgs {
			if client.allowedOrgs[i] != org {
				t.Errorf("AllowedOrgs[%d] mismatch: got %s, want %s", i, client.allowedOrgs[i], org)
			}
		}

		// Check HTTP client timeout
		if client.httpClient.Timeout.Seconds() != 30 {
			t.Errorf("HTTP client timeout mismatch: got %v, want 30s", client.httpClient.Timeout)
		}
	})
}

func TestGitHubClient_GetAuthURL(t *testing.T) {
	t.Parallel()

	t.Run("generates auth URL with state", func(t *testing.T) {
		t.Parallel()

		client := NewGitHubClient("client-id", "client-secret", "callback-url", nil)
		state := "test-state-123"

		authURL := client.GetAuthURL(state)

		// Should contain the state parameter
		if authURL == "" {
			t.Error("GetAuthURL() returned empty string")
		}

		// Basic check that it contains GitHub OAuth URL
		expectedPrefix := "https://github.com/login/oauth/authorize"
		if !containsString(authURL, expectedPrefix) {
			t.Errorf("GetAuthURL() should contain GitHub OAuth URL, got: %s", authURL)
		}

		// Should contain the state parameter
		if !containsString(authURL, "state="+state) {
			t.Errorf("GetAuthURL() should contain state parameter, got: %s", authURL)
		}
	})
}

func TestGitHubClient_ValidateOrgMembership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		allowedOrgs []string
		userOrgs    []GitHubOrg
		expected    bool
	}{
		{
			name:        "empty allowed orgs should allow any user",
			allowedOrgs: []string{},
			userOrgs:    []GitHubOrg{{Login: "some-org"}},
			expected:    true,
		},
		{
			name:        "user in allowed org",
			allowedOrgs: []string{"allowed-org"},
			userOrgs:    []GitHubOrg{{Login: "allowed-org"}},
			expected:    true,
		},
		{
			name:        "user not in allowed orgs",
			allowedOrgs: []string{"allowed-org"},
			userOrgs:    []GitHubOrg{{Login: "other-org"}},
			expected:    false,
		},
		{
			name:        "user in multiple orgs, one allowed",
			allowedOrgs: []string{"allowed-org"},
			userOrgs: []GitHubOrg{
				{Login: "other-org"},
				{Login: "allowed-org"},
			},
			expected: true,
		},
		{
			name:        "multiple allowed orgs, user in one",
			allowedOrgs: []string{"org1", "org2", "org3"},
			userOrgs:    []GitHubOrg{{Login: "org2"}},
			expected:    true,
		},
		{
			name:        "no user orgs",
			allowedOrgs: []string{"allowed-org"},
			userOrgs:    []GitHubOrg{},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := NewGitHubClient("client-id", "client-secret", "callback-url", tt.allowedOrgs)
			result := client.ValidateOrgMembership(tt.userOrgs)

			if result != tt.expected {
				t.Errorf("ValidateOrgMembership() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
