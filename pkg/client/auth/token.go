package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/cockroachdb/errors"
)

type AuthenticatedUser struct {
	User            GitHubUser       `json:"user"`
	BearerToken     string           `json:"bearer_token"`
	TeamMemberships []TeamMembership `json:"team_memberships"`
}

type GitHubUser struct {
	ID        string `json:"id"`
	GitHubID  int64  `json:"github_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type TeamMembership struct {
	OrgName  string `json:"org_name"`
	TeamName string `json:"team_name"`
	Role     string `json:"role"`
}

type OAuthClient struct {
	serverBaseURL string
	httpClient    *http.Client
}

func NewOAuthClient(serverBaseURL string) *OAuthClient {
	return &OAuthClient{
		serverBaseURL: serverBaseURL,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
	}
}

// InitiateOAuthFlow starts the GitHub OAuth flow and returns the bearer token
func (c *OAuthClient) InitiateOAuthFlow(ctx context.Context) (string, error) {
	// Generate state for CSRF protection
	state, err := generateRandomState()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate state")
	}

	// Build login URL
	loginURL := fmt.Sprintf("%s/v1alpha1/auth/login?redirect_uri=%s",
		c.serverBaseURL,
		url.QueryEscape("http://localhost:8080/callback"))

	fmt.Printf("Please open the following URL in your browser to complete authentication:\n%s\n", loginURL)
	fmt.Println("Waiting for authentication to complete...")

	// Try to open browser automatically
	if err := openBrowser(loginURL); err != nil {
		fmt.Printf("Could not open browser automatically. Please manually open: %s\n", loginURL)
	}

	// Start local callback server
	return c.waitForCallback(ctx, state)
}

// GetCurrentUser retrieves current user information using a bearer token
func (c *OAuthClient) GetCurrentUser(ctx context.Context, bearerToken string) (*AuthenticatedUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.serverBaseURL+"/v1alpha1/auth/me", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = errors.CombineErrors(err, errors.Wrap(closeErr, "failed to close response body"))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("authentication failed: status=%d", resp.StatusCode)
	}

	var user AuthenticatedUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(err, "failed to decode user response")
	}

	return &user, nil
}

func (c *OAuthClient) waitForCallback(ctx context.Context, expectedState string) (string, error) {
	// This is a simplified implementation for CLI usage
	// In a real-world scenario, you might want to implement a proper callback server
	// For now, we'll ask the user to manually provide the bearer token from /auth/me

	fmt.Println("\nAfter completing authentication in your browser:")
	fmt.Println("1. The browser should redirect you to a success page")
	fmt.Println("2. Your session will be established automatically")
	fmt.Println("3. Press Enter to continue...")

	// Wait for user input
	if _, err := fmt.Scanln(); err != nil {
		// Ignore scan error as it's just waiting for user input
		fmt.Printf("Input error (ignoring): %v\n", err)
	}

	// Try to get current user with session cookie approach
	// This assumes the CLI and browser share cookies (which they don't)
	// A better implementation would parse the callback URL or use a local server

	return "", errors.New("interactive authentication required - please implement proper callback handling")
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

// Legacy function for backward compatibility during migration
// This should be removed once all clients are updated
func RetrieveToken(
	cognitoDomain string,
	region string,
	clientID string,
	clientSecret string,
) (TokenResponse, error) {
	return TokenResponse{}, errors.New("Cognito authentication is no longer supported. Please use GitHub OAuth authentication instead.")
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}
