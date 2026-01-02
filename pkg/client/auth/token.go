package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/cockroachdb/errors"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func RetrieveToken(
	cognitoDomain string,
	region string,
	clientID string,
	clientSecret string,
) (v TokenResponse, err error) {
	// Cognito token endpoint format: https://<domain>.auth.<region>.amazoncognito.com/oauth2/token
	tokenURL := fmt.Sprintf("https://%s.auth.%s.amazoncognito.com/oauth2/token", cognitoDomain, region)

	// Cognito requires application/x-www-form-urlencoded format
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return TokenResponse{}, errors.Wrapf(err, "failed to create token request")
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return TokenResponse{}, errors.Wrapf(err, "failed to send token request")
	}

	defer func() {
		err = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		var errorBody map[string]any
		if decodeErr := json.NewDecoder(res.Body).Decode(&errorBody); decodeErr == nil {
			return TokenResponse{}, errors.Errorf("failed to retrieve token: status=%d, error=%v", res.StatusCode, errorBody)
		}
		return TokenResponse{}, errors.Errorf("failed to retrieve token: status=%d", res.StatusCode)
	}

	tokenResponse := TokenResponse{}
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return TokenResponse{}, errors.Wrapf(err, "failed to decode token response")
	}

	return tokenResponse, nil
}
