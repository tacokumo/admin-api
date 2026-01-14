package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tacokumo/admin-api/pkg/auth/session"
)

// MockSessionStore implements session.Store for testing
type MockSessionStore struct {
	sessions map[string]*session.Session
	getError error
}

func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{
		sessions: make(map[string]*session.Session),
	}
}

func (m *MockSessionStore) Create(ctx context.Context, session *session.Session) error {
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionStore) Get(ctx context.Context, sessionID string) (*session.Session, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	sess, exists := m.sessions[sessionID]
	if !exists {
		return nil, session.ErrSessionNotFound
	}
	return sess, nil
}

func (m *MockSessionStore) Delete(ctx context.Context, sessionID string) error {
	delete(m.sessions, sessionID)
	return nil
}

func (m *MockSessionStore) Refresh(ctx context.Context, sessionID string, newExpiry time.Time) error {
	if sess, exists := m.sessions[sessionID]; exists {
		sess.ExpiresAt = newExpiry
		return nil
	}
	return session.ErrSessionNotFound
}

func (m *MockSessionStore) SetGetError(err error) {
	m.getError = err
}

func TestIsPublicPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path     string
		expected bool
	}{
		{"/v1alpha1/health/liveness", true},
		{"/v1alpha1/health/readiness", true},
		{"/v1alpha1/auth/login", true},
		{"/v1alpha1/auth/callback", true},
		{"/v1alpha1/projects", false},
		{"/v1alpha1/users", false},
		{"/", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run("path: "+tt.path, func(t *testing.T) {
			t.Parallel()

			result := isPublicPath(tt.path)
			if result != tt.expected {
				t.Errorf("isPublicPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractSessionID(t *testing.T) {
	t.Parallel()

	t.Run("extracts session ID from Authorization header", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer test-session-id")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		sessionID := extractSessionID(c)
		if sessionID != "test-session-id" {
			t.Errorf("extractSessionID() = %q, want %q", sessionID, "test-session-id")
		}
	})

	t.Run("extracts session ID from cookie", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "cookie-session-id",
		})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		sessionID := extractSessionID(c)
		if sessionID != "cookie-session-id" {
			t.Errorf("extractSessionID() = %q, want %q", sessionID, "cookie-session-id")
		}
	})

	t.Run("prefers Authorization header over cookie", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer header-session-id")
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "cookie-session-id",
		})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		sessionID := extractSessionID(c)
		if sessionID != "header-session-id" {
			t.Errorf("extractSessionID() = %q, want %q", sessionID, "header-session-id")
		}
	})

	t.Run("returns empty string when no session ID found", func(t *testing.T) {
		t.Parallel()

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		sessionID := extractSessionID(c)
		if sessionID != "" {
			t.Errorf("extractSessionID() = %q, want empty string", sessionID)
		}
	})
}

func TestSessionMiddleware(t *testing.T) {
	t.Parallel()

	logger := slog.Default()

	t.Run("allows access to public paths without session", func(t *testing.T) {
		t.Parallel()

		store := NewMockSessionStore()
		middleware := SessionMiddleware(logger, store)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1alpha1/health/liveness", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		nextCalled := false
		next := func(c echo.Context) error {
			nextCalled = true
			return c.String(http.StatusOK, "OK")
		}

		err := middleware(next)(c)
		if err != nil {
			t.Errorf("SessionMiddleware() returned error for public path: %v", err)
		}
		if !nextCalled {
			t.Error("SessionMiddleware() should call next handler for public paths")
		}
	})

	t.Run("returns 401 for protected paths without session ID", func(t *testing.T) {
		t.Parallel()

		store := NewMockSessionStore()
		middleware := SessionMiddleware(logger, store)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1alpha1/projects", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c echo.Context) error {
			t.Error("Next handler should not be called when session ID is missing")
			return nil
		}

		err := middleware(next)(c)
		if err == nil {
			t.Error("SessionMiddleware() should return error when session ID is missing")
		}

		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Errorf("Expected *echo.HTTPError, got %T", err)
		} else if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, httpErr.Code)
		}
	})

	t.Run("returns 401 for non-existent session", func(t *testing.T) {
		t.Parallel()

		store := NewMockSessionStore()
		middleware := SessionMiddleware(logger, store)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1alpha1/projects", nil)
		req.Header.Set("Authorization", "Bearer non-existent-session")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c echo.Context) error {
			t.Error("Next handler should not be called for non-existent session")
			return nil
		}

		err := middleware(next)(c)
		if err == nil {
			t.Error("SessionMiddleware() should return error for non-existent session")
		}

		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Errorf("Expected *echo.HTTPError, got %T", err)
		} else if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, httpErr.Code)
		}
	})

	t.Run("returns 401 for expired session", func(t *testing.T) {
		t.Parallel()

		store := NewMockSessionStore()
		expiredSession := &session.Session{
			ID:        "expired-session",
			UserID:    "user-123",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		if err := store.Create(context.Background(), expiredSession); err != nil {
			t.Fatalf("Failed to create expired session: %v", err)
		}

		middleware := SessionMiddleware(logger, store)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1alpha1/projects", nil)
		req.Header.Set("Authorization", "Bearer expired-session")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c echo.Context) error {
			t.Error("Next handler should not be called for expired session")
			return nil
		}

		err := middleware(next)(c)
		if err == nil {
			t.Error("SessionMiddleware() should return error for expired session")
		}

		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Errorf("Expected *echo.HTTPError, got %T", err)
		} else if httpErr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, httpErr.Code)
		}
	})

	t.Run("allows access with valid session", func(t *testing.T) {
		t.Parallel()

		store := NewMockSessionStore()
		validSession := &session.Session{
			ID:        "valid-session",
			UserID:    "user-123",
			ExpiresAt: time.Now().Add(1 * time.Hour), // Valid for 1 hour
			CreatedAt: time.Now(),
		}
		if err := store.Create(context.Background(), validSession); err != nil {
			t.Fatalf("Failed to create valid session: %v", err)
		}

		middleware := SessionMiddleware(logger, store)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1alpha1/projects", nil)
		req.Header.Set("Authorization", "Bearer valid-session")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var sessionFromContext *session.Session
		next := func(c echo.Context) error {
			sessionFromContext = GetCurrentSession(c.Request().Context())
			return c.String(http.StatusOK, "OK")
		}

		err := middleware(next)(c)
		if err != nil {
			t.Errorf("SessionMiddleware() returned error for valid session: %v", err)
		}

		if sessionFromContext == nil {
			t.Error("Session should be available in context")
		} else if sessionFromContext.ID != "valid-session" {
			t.Errorf("Expected session ID %q, got %q", "valid-session", sessionFromContext.ID)
		}
	})

	t.Run("returns 500 for session store internal error", func(t *testing.T) {
		t.Parallel()

		store := NewMockSessionStore()
		store.SetGetError(errors.New("internal database error"))

		middleware := SessionMiddleware(logger, store)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/v1alpha1/projects", nil)
		req.Header.Set("Authorization", "Bearer some-session")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c echo.Context) error {
			t.Error("Next handler should not be called on store error")
			return nil
		}

		err := middleware(next)(c)
		if err == nil {
			t.Error("SessionMiddleware() should return error on store error")
		}

		httpErr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Errorf("Expected *echo.HTTPError, got %T", err)
		} else if httpErr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, httpErr.Code)
		}
	})
}

func TestGetCurrentSession(t *testing.T) {
	t.Parallel()

	t.Run("returns session from context", func(t *testing.T) {
		t.Parallel()

		testSession := &session.Session{
			ID:     "test-session",
			UserID: "user-123",
		}

		ctx := context.WithValue(context.Background(), CurrentSessionKey, testSession)
		result := GetCurrentSession(ctx)

		if result == nil {
			t.Error("GetCurrentSession() should return session from context")
		} else if result.ID != "test-session" {
			t.Errorf("Expected session ID %q, got %q", "test-session", result.ID)
		}
	})

	t.Run("returns nil when no session in context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		result := GetCurrentSession(ctx)

		if result != nil {
			t.Errorf("GetCurrentSession() should return nil when no session in context, got %v", result)
		}
	})

	t.Run("returns nil when wrong type in context", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(context.Background(), CurrentSessionKey, "not-a-session")
		result := GetCurrentSession(ctx)

		if result != nil {
			t.Errorf("GetCurrentSession() should return nil when wrong type in context, got %v", result)
		}
	})
}
