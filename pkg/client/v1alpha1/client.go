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
	c                   http.Client
	logger              *slog.Logger
	serverBaseURL       string
	cognitoDomain       string
	cognitoRegion       string
	cognitoClientID     string
	cognitoClientSecret string
}

func NewDefaultClient(logger *slog.Logger, httpClient http.Client) *DefaultClient {
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "localhost"
	}
	return &DefaultClient{
		c:                   httpClient,
		logger:              logger.With(slog.String("component", "v1alpha1client")),
		serverBaseURL:       fmt.Sprintf("https://%s:8444", serverHost),
		cognitoDomain:       os.Getenv("COGNITO_DOMAIN"),
		cognitoRegion:       os.Getenv("COGNITO_REGION"),
		cognitoClientID:     os.Getenv("COGNITO_CLIENT_ID"),
		cognitoClientSecret: os.Getenv("COGNITO_CLIENT_SECRET"),
	}
}

func (c *DefaultClient) post(
	ctx context.Context,
	endpoint string,
	reqBody any,
) (*http.Response, error) {
	token, err := auth.RetrieveToken(c.cognitoDomain, c.cognitoRegion, c.cognitoClientID, c.cognitoClientSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve token")
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
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
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
	token, err := auth.RetrieveToken(c.cognitoDomain, c.cognitoRegion, c.cognitoClientID, c.cognitoClientSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve token")
	}

	uri := fmt.Sprintf("%s%s", c.serverBaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GET request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
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
