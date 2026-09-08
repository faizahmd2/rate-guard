package redis

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
)

//go:embed scripts/token_bucket.lua
var tokenBucketScript string

type TokenBucketResult struct {
	Allowed      bool
	Remaining    int
	Limit        int
	RetryAfterMs int64
}

func (c *Client) CheckTokenBucket(
	ctx context.Context,
	key string,
	capacity float64,
	refillRate float64,
	cost float64,
) (TokenBucketResult, error) {

	if capacity <= 0 {
		return TokenBucketResult{}, fmt.Errorf("capacity must be greater than zero")
	}

	if refillRate <= 0 {
		return TokenBucketResult{}, fmt.Errorf("refill rate must be greater than zero")
	}

	if cost <= 0 {
		return TokenBucketResult{}, fmt.Errorf("cost must be greater than zero")
	}

	result, err := c.client.Eval(
		ctx,
		tokenBucketScript,
		[]string{key},
		strconv.FormatFloat(capacity, 'f', -1, 64),
		strconv.FormatFloat(refillRate, 'f', -1, 64),
		strconv.FormatFloat(cost, 'f', -1, 64),
	).Result()

	if err != nil {
		return TokenBucketResult{}, fmt.Errorf(
			"execute token bucket script: %w",
			err,
		)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		return TokenBucketResult{}, fmt.Errorf(
			"unexpected token bucket result",
		)
	}

	allowed, err := redisInt(values[0])
	if err != nil {
		return TokenBucketResult{}, err
	}

	remaining, err := redisInt(values[1])
	if err != nil {
		return TokenBucketResult{}, err
	}

	limit, err := redisInt(values[2])
	if err != nil {
		return TokenBucketResult{}, err
	}

	retryAfterMs, err := redisInt64(values[3])
	if err != nil {
		return TokenBucketResult{}, err
	}

	return TokenBucketResult{
		Allowed:      allowed == 1,
		Remaining:    remaining,
		Limit:        limit,
		RetryAfterMs: retryAfterMs,
	}, nil
}

func redisInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("unexpected redis integer type: %T", value)
	}
}

func redisInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf(
			"unexpected redis integer type: %T",
			value,
		)
	}
}
