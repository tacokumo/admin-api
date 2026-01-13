package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubClient struct {
	config      *oauth2.Config
	httpClient  *http.Client
	allowedOrgs []string
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubOrg struct {
	Login string `json:"login"`
}

type GitHubTeam struct {
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Organization GitHubOrg `json:"organization"`
}

type TeamMembership struct {
	OrgName  string
	TeamName string
	Role     string
}

func NewGitHubClient(clientID, clientSecret, callbackURL string, allowedOrgs []string) *GitHubClient {
	return &GitHubClient{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  callbackURL,
			Scopes:       []string{"user:email", "read:org"},
			Endpoint:     github.Endpoint,
		},
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		allowedOrgs: allowedOrgs,
	}
}

func (c *GitHubClient) GetAuthURL(state string) string {
	return c.config.AuthCodeURL(state)
}

func (c *GitHubClient) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exchange code for token")
	}
	return token, nil
}

func (c *GitHubClient) GetUser(ctx context.Context, token *oauth2.Token) (*GitHubUser, error) {
	client := c.config.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user info")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("github api returned status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(err, "failed to decode user info")
	}

	if user.Email == "" {
		email, err := c.getPrimaryEmail(ctx, client)
		if err != nil {
			return nil, err
		}
		user.Email = email
	}

	return &user, nil
}

func (c *GitHubClient) getPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", errors.Wrap(err, "failed to get user emails")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.Newf("github api returned status %d for emails", resp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", errors.Wrap(err, "failed to decode emails")
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}

	return "", errors.New("no primary verified email found")
}

func (c *GitHubClient) GetUserOrgs(ctx context.Context, token *oauth2.Token) ([]GitHubOrg, error) {
	client := c.config.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user/orgs")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user orgs")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("github api returned status %d for orgs", resp.StatusCode)
	}

	var orgs []GitHubOrg
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, errors.Wrap(err, "failed to decode orgs")
	}

	return orgs, nil
}

func (c *GitHubClient) ValidateOrgMembership(orgs []GitHubOrg) bool {
	if len(c.allowedOrgs) == 0 {
		return true
	}

	for _, org := range orgs {
		if lo.Contains(c.allowedOrgs, org.Login) {
			return true
		}
	}

	return false
}

func (c *GitHubClient) GetTeamMemberships(ctx context.Context, token *oauth2.Token) ([]TeamMembership, error) {
	client := c.config.Client(ctx, token)

	resp, err := client.Get("https://api.github.com/user/teams")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user teams")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("github api returned status %d for teams", resp.StatusCode)
	}

	var teams []GitHubTeam
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, errors.Wrap(err, "failed to decode teams")
	}

	memberships := make([]TeamMembership, 0, len(teams))
	for _, team := range teams {
		memberships = append(memberships, TeamMembership{
			OrgName:  team.Organization.Login,
			TeamName: team.Slug,
			Role:     "member",
		})
	}

	return memberships, nil
}
