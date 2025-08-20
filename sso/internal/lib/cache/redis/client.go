package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"loginform/sso/internal/lib/cache"
)

type Client struct {
	rdb *goredis.Client
}

func NewClient(opt *goredis.Options) *Client {
	return &Client{
		rdb: goredis.NewClient(opt),
	}
}

var _ cache.Cache = (*Client)(nil)

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	s, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return "", nil // если в кэше нету - ошибку не возвращаем, а идем в бд
	}

	return s, err
}

func (c *Client) SetEX(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
