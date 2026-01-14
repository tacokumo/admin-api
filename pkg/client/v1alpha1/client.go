package v1alpha1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/client/auth"
)

type Client interface {
	CreateProject(ctx context.Context, req *generated.CreateProjectRequest) error
	ListProjects(ctx context.Context) ([]generated.Project, error)
	LivenessCheck(ctx context.Context) error
	ReadinessCheck(ctx context.Context) error
}

type DefaultClient struct {
	c             http.Client
	logger        *slog.Logger
	serverBaseURL string
	oauthClient   *auth.OAuthClient
	bearerToken   string
}

func NewDefaultClient(logger *slog.Logger, httpClient http.Client) *DefaultClient {
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "localhost"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8444"
	}

	serverScheme := os.Getenv("SERVER_SCHEME")
	if serverScheme == "" {
		serverScheme = "https"
	}

	serverBaseURL := fmt.Sprintf("%s://%s:%s", serverScheme, serverHost, serverPort)
	oauthClient := auth.NewOAuthClient(serverBaseURL)

	return &DefaultClient{
		c:             httpClient,
		logger:        logger.With(slog.String("component", "v1alpha1client")),
		serverBaseURL: serverBaseURL,
		oauthClient:   oauthClient,
		bearerToken:   "", // Will be set during authentication
	}
}

func (c *DefaultClient) post(
	ctx context.Context,
	endpoint string,
	reqBody any,
) (*http.Response, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, errors.Wrap(err, "authentication failed")
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal request body")
	}
	uri := fmt.Sprintf("%s%s", c.serverBaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create POST request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send POST request")
	}

	return resp, nil
}

func (c *DefaultClient) get(
	ctx context.Context,
	endpoint string,
	queryParams map[string]string,
) (*http.Response, error) {
	if err := c.ensureAuthenticated(ctx); err != nil {
		return nil, errors.Wrap(err, "authentication failed")
	}

	uri := fmt.Sprintf("%s%s", c.serverBaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GET request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	q := req.URL.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to send GET request")
	}

	return resp, nil
}

func (c *DefaultClient) LivenessCheck(ctx context.Context) error {
	_, err := c.get(ctx, "/v1alpha1/health/liveness", nil)
	if err != nil {
		return errors.Wrapf(err, "failed to check liveness")
	}
	return nil
}

func (c *DefaultClient) ReadinessCheck(ctx context.Context) error {
	_, err := c.get(ctx, "/v1alpha1/health/readiness", nil)
	if err != nil {
		return errors.Wrapf(err, "failed to check readiness")
	}
	return nil
}

// ensureAuthenticated checks if the client has a valid bearer token and initiates OAuth if needed
func (c *DefaultClient) ensureAuthenticated(ctx context.Context) error {
	if c.bearerToken != "" {
		// Verify token is still valid
		_, err := c.oauthClient.GetCurrentUser(ctx, c.bearerToken)
		if err == nil {
			return nil // Token is valid
		}
		c.logger.WarnContext(ctx, "Bearer token appears to be invalid, re-authenticating", slog.String("error", err.Error()))
	}

	// Check if token is provided via environment variable
	if envToken := os.Getenv("BEARER_TOKEN"); envToken != "" {
		c.bearerToken = envToken
		_, err := c.oauthClient.GetCurrentUser(ctx, c.bearerToken)
		if err == nil {
			c.logger.InfoContext(ctx, "Using bearer token from environment variable")
			return nil
		}
		c.logger.WarnContext(ctx, "Bearer token from environment variable is invalid", slog.String("error", err.Error()))
	}

	// No valid token, need to authenticate
	c.logger.InfoContext(ctx, "No valid authentication found, initiating OAuth flow")

	// For CLI usage, we'll use a simplified approach
	// In production, you might want to implement a proper OAuth callback server
	fmt.Println("GitHub OAuth authentication required.")
	fmt.Printf("Please visit: %s/v1alpha1/auth/login\n", c.serverBaseURL)
	fmt.Println("After completing authentication in your browser, please enter your bearer token:")
	fmt.Println("(You can get this from the browser's developer tools or by calling /v1alpha1/auth/me)")

	var token string
	if _, err := fmt.Scanln(&token); err != nil {
		return errors.Wrap(err, "failed to read bearer token")
	}

	// Verify the token works
	if _, err := c.oauthClient.GetCurrentUser(ctx, token); err != nil {
		return errors.Wrap(err, "provided bearer token is invalid")
	}

	c.bearerToken = token
	c.logger.InfoContext(ctx, "Authentication successful")
	return nil
}

// Authenticate allows explicit authentication with OAuth
func (c *DefaultClient) Authenticate(ctx context.Context) error {
	c.bearerToken = "" // Force re-authentication
	return c.ensureAuthenticated(ctx)
}

// SetBearerToken allows setting the bearer token directly (useful for testing or when token is known)
func (c *DefaultClient) SetBearerToken(token string) {
	c.bearerToken = token
}
