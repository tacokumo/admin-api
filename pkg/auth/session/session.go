package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/cockroachdb/errors"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	GitHubUserID    int64            `json:"github_user_id"`
	GitHubUsername  string           `json:"github_username"`
	Email           string           `json:"email"`
	Name            string           `json:"name"`
	AvatarURL       string           `json:"avatar_url"`
	AccessToken     string           `json:"access_token"`
	RefreshToken    string           `json:"refresh_token"`
	TeamMemberships []TeamMembership `json:"team_memberships"`
	ExpiresAt       time.Time        `json:"expires_at"`
	CreatedAt       time.Time        `json:"created_at"`
}

type TeamMembership struct {
	OrgName  string `json:"org_name"`
	TeamName string `json:"team_name"`
	Role     string `json:"role"`
}

type Store interface {
	Create(ctx context.Context, session *Session) error
	Get(ctx context.Context, sessionID string) (*Session, error)
	Delete(ctx context.Context, sessionID string) error
	Refresh(ctx context.Context, sessionID string, newExpiry time.Time) error
}

func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "failed to generate random bytes")
	}
	return hex.EncodeToString(bytes), nil
}
