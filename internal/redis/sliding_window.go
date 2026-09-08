package redis

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
)

//go:embed scripts/sliding_window.lua
var slidingWindowScript string

type SlidingWindowResult struct {
	Allowed      bool
	Remaining    int
	Limit        int
	RetryAfterMs int64
}

func CheckSlidingWindow(
	ctx context.Context,
	client *Client,
	key string,
	limit int,
	windowSeconds int,
) (SlidingWindowResult, error) {

	if limit <= 0 {
		return SlidingWindowResult{}, fmt.Errorf(
			"sliding window limit must be greater than zero",
		)
	}

	if windowSeconds <= 0 {
		return SlidingWindowResult{}, fmt.Errorf(
			"sliding window window_seconds must be greater than zero",
		)
	}

	result, err := client.client.Eval(
		ctx,
		slidingWindowScript,
		[]string{key},
		strconv.Itoa(limit),
		strconv.Itoa(windowSeconds),
	).Result()

	if err != nil {
		return SlidingWindowResult{}, fmt.Errorf(
			"execute sliding window script: %w",
			err,
		)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		return SlidingWindowResult{}, fmt.Errorf(
			"unexpected sliding window result",
		)
	}

	allowed, ok := values[0].(int64)
	if !ok {
		return SlidingWindowResult{}, fmt.Errorf(
			"invalid sliding window allowed result",
		)
	}

	remaining, ok := values[1].(int64)
	if !ok {
		return SlidingWindowResult{}, fmt.Errorf(
			"invalid sliding window remaining result",
		)
	}

	resultLimit, ok := values[2].(int64)
	if !ok {
		return SlidingWindowResult{}, fmt.Errorf(
			"invalid sliding window limit result",
		)
	}

	retryAfterMs, ok := values[3].(int64)
	if !ok {
		return SlidingWindowResult{}, fmt.Errorf(
			"invalid sliding window retry_after result",
		)
	}

	return SlidingWindowResult{
		Allowed:      allowed == 1,
		Remaining:    int(remaining),
		Limit:        int(resultLimit),
		RetryAfterMs: retryAfterMs,
	}, nil
}
