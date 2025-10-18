package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func RetrieveToken(
	auth0Domain string,
	clientID string,
	clientSecret string,
	audience string,
) (v TokenResponse, err error) {
	url := fmt.Sprintf("https://%s/oauth/token", auth0Domain)

	bodyJSON, err := json.Marshal(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      audience,
		"grant_type":    "client_credentials",
	})
	if err != nil {
		return TokenResponse{}, errors.WithStack(err)
	}

	payload := bytes.NewReader(bodyJSON)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer func() {
		err = res.Body.Close()
	}()
	tokenResponse := TokenResponse{}
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return TokenResponse{}, errors.WithStack(err)
	}

	return tokenResponse, nil
}
