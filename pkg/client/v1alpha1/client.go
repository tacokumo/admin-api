package v1alpha1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/client/auth"
)

type Client interface {
	CreateProject(ctx context.Context, req *generated.CreateProjectRequest) error
}

type DefaultClient struct {
	c             http.Client
	serverBaseURL string
	auth0Domain   string
	auth0ClientID string
	auth0Secret   string
	auth0Audience string
}

func NewDefaultClient(httpClient http.Client) *DefaultClient {
	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "localhost"
	}
	return &DefaultClient{
		c:             httpClient,
		serverBaseURL: fmt.Sprintf("https://%s:8444", serverHost),
		auth0Domain:   os.Getenv("AUTH0_DOMAIN"),
		auth0ClientID: os.Getenv("AUTH0_CLIENT_ID"),
		auth0Secret:   os.Getenv("AUTH0_CLIENT_SECRET"),
		auth0Audience: os.Getenv("AUTH0_AUDIENCE"),
	}
}

func (c *DefaultClient) post(
	ctx context.Context,
	endpoint string,
	reqBody any,
) (*http.Response, error) {
	token, err := auth.RetrieveToken(c.auth0Domain, c.auth0ClientID, c.auth0Secret, c.auth0Audience)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	uri := fmt.Sprintf("%s%s/", c.serverBaseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.c.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}
