package redis

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/redis/go-redis/v9"
	"github.com/tacokumo/admin-api/pkg/config"
)

type Client struct {
	rdb *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "failed to ping redis")
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Underlying() *redis.Client {
	return c.rdb
}
