package oauth

import "errors"

// OAuth error definitions for testing
var (
	ErrInvalidCode    = errors.New("invalid authorization code")
	ErrUserInfoFailed = errors.New("failed to fetch user information")
	ErrOrgsFailed     = errors.New("failed to fetch user organizations")
	ErrTeamsFailed    = errors.New("failed to fetch team memberships")
)
