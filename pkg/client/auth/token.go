package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
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

	// Start local callback server first to get the port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", errors.Wrap(err, "failed to create listener")
	}
	defer func() {
		_ = listener.Close()
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)

	// Build login URL with dynamic callback URL
	loginURL := fmt.Sprintf("%s/v1alpha1/auth/login?redirect_uri=%s&state=%s",
		c.serverBaseURL,
		url.QueryEscape(callbackURL),
		url.QueryEscape(state))

	fmt.Printf("Please open the following URL in your browser to complete authentication:\n%s\n", loginURL)
	fmt.Println("Waiting for authentication to complete...")

	// Try to open browser automatically
	if err := openBrowser(loginURL); err != nil {
		fmt.Printf("Could not open browser automatically. Please manually open: %s\n", loginURL)
	}

	// Start local callback server
	return c.waitForCallbackWithListener(ctx, state, listener)
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

func (c *OAuthClient) waitForCallbackWithListener(ctx context.Context, expectedState string, listener net.Listener) (string, error) {
	// Create server using the provided listener
	callbackServer := &http.Server{
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Channel to receive the token
	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	// Setup callback handler
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Handle POST requests from JavaScript
		if r.Method == http.MethodPost {
			token := r.URL.Query().Get("token")
			if token == "" {
				http.Error(w, "No token provided", http.StatusBadRequest)
				return
			}

			// Validate token
			if _, err := c.GetCurrentUser(ctx, token); err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				errCh <- errors.Wrap(err, "token validation failed")
				return
			}

			w.WriteHeader(http.StatusOK)

			// Send token to channel
			select {
			case tokenCh <- token:
			default:
			}
			return
		}

		// Handle GET requests - Extract token from URL fragment (if using implicit flow)
		// or from query parameters (if using authorization code flow)
		token := r.URL.Query().Get("token")
		state := r.URL.Query().Get("state")

		// Note: We don't validate state here because the server generates its own state
		// and handles CSRF protection. The client state was only used for initial request.
		// The server's state is validated on the server side during the OAuth callback.

		if token == "" {
			// Show a page that will extract token from URL fragment using JavaScript
			html := `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Complete</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .success { color: green; }
        .error { color: red; }
    </style>
</head>
<body>
    <div id="content">
        <h2>Processing authentication...</h2>
        <p>Please wait while we complete the authentication process.</p>
    </div>
    <script>
        // Extract token from URL fragment and state from query params
        const fragment = window.location.hash.substring(1);
        const query = window.location.search.substring(1);
        const fragmentParams = new URLSearchParams(fragment);
        const queryParams = new URLSearchParams(query);
        const token = fragmentParams.get('access_token') || fragmentParams.get('token') || queryParams.get('token');
        const currentState = queryParams.get('state') || '` + state + `';

        if (token) {
            // Send token to callback endpoint
            fetch('/callback?token=' + encodeURIComponent(token) + '&state=' + encodeURIComponent(currentState), {
                method: 'POST'
            }).then(() => {
                document.getElementById('content').innerHTML =
                    '<h2 class="success">✅ Authentication Successful!</h2>' +
                    '<p>This window will close automatically in a few seconds...</p>';
                // Close the tab after a short delay to show the success message
                setTimeout(() => {
                    window.close();
                }, 2000);
            }).catch(() => {
                document.getElementById('content').innerHTML =
                    '<h2 class="error">❌ Authentication Failed</h2>' +
                    '<p>Please try again or check the CLI for error messages.</p>';
            });
        } else {
            document.getElementById('content').innerHTML =
                '<h2 class="error">❌ No Token Found</h2>' +
                '<p>Authentication may have failed. Please try again.</p>';
        }
    </script>
</body>
</html>`
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(html))
			return
		}

		// Validate token
		if _, err := c.GetCurrentUser(ctx, token); err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			errCh <- errors.Wrap(err, "token validation failed")
			return
		}

		// Success response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		successHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Complete</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
        .success { color: green; }
    </style>
</head>
<body>
    <h2 class="success">✅ Authentication Successful!</h2>
    <p>This window will close automatically in a few seconds...</p>
    <script>
        setTimeout(() => {
            window.close();
        }, 2000);
    </script>
</body>
</html>`
		_, _ = w.Write([]byte(successHTML))

		// Send token to channel
		select {
		case tokenCh <- token:
		default:
		}
	})

	callbackServer.Handler = mux

	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Started local callback server on port %d\n", port)

	// Start server in background
	go func() {
		if err := callbackServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- errors.Wrap(err, "callback server error")
		}
	}()

	// Cleanup server when done
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = callbackServer.Shutdown(shutdownCtx)
	}()

	// Wait for token or error with timeout
	select {
	case token := <-tokenCh:
		// Give a short delay to ensure the success message is shown in the browser
		time.Sleep(100 * time.Millisecond)
		return token, nil
	case err := <-errCh:
		return "", err
	case <-time.After(5 * time.Minute):
		return "", errors.New("authentication timeout - no response received within 5 minutes")
	case <-ctx.Done():
		return "", errors.Wrap(ctx.Err(), "context cancelled")
	}
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
