package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	client *goredis.Client
}

var ErrKeyNotFound = goredis.Nil

func NewClient(
	ctx context.Context,
	host string,
	port string,
) (*Client, error) {

	client := goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

func (c *Client) Get(
	ctx context.Context,
	key string,
) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Set(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) error {
	return c.client.Set(
		ctx,
		key,
		value,
		expiration,
	).Err()
}

func (c *Client) Delete(
	ctx context.Context,
	key string,
) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
