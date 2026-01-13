package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "session:"

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStore(client *redis.Client, ttl time.Duration) *RedisStore {
	return &RedisStore{
		client: client,
		ttl:    ttl,
	}
}

func (s *RedisStore) Create(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = s.ttl
	}

	if err := s.client.Set(ctx, keyPrefix+session.ID, data, ttl).Err(); err != nil {
		return errors.Wrap(err, "failed to store session in redis")
	}

	return nil
}

func (s *RedisStore) Get(ctx context.Context, sessionID string) (*Session, error) {
	data, err := s.client.Get(ctx, keyPrefix+sessionID).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrSessionNotFound
		}
		return nil, errors.Wrap(err, "failed to get session from redis")
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal session")
	}

	return &session, nil
}

func (s *RedisStore) Delete(ctx context.Context, sessionID string) error {
	if err := s.client.Del(ctx, keyPrefix+sessionID).Err(); err != nil {
		return errors.Wrap(err, "failed to delete session from redis")
	}
	return nil
}

func (s *RedisStore) Refresh(ctx context.Context, sessionID string, newExpiry time.Time) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = newExpiry

	return s.Create(ctx, session)
}
