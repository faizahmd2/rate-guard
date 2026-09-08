package redis

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
)

//go:embed scripts/fixed_window.lua
var fixedWindowScript string

type FixedWindowResult struct {
	Allowed      bool
	Remaining    int
	Limit        int
	RetryAfterMs int64
}

func CheckFixedWindow(
	ctx context.Context,
	client *Client,
	key string,
	limit int,
	windowSeconds int,
) (FixedWindowResult, error) {

	if limit <= 0 {
		return FixedWindowResult{}, fmt.Errorf(
			"fixed window limit must be greater than zero",
		)
	}

	if windowSeconds <= 0 {
		return FixedWindowResult{}, fmt.Errorf(
			"fixed window window_seconds must be greater than zero",
		)
	}

	result, err := client.client.Eval(
		ctx,
		fixedWindowScript,
		[]string{key},
		strconv.Itoa(limit),
		strconv.Itoa(windowSeconds),
	).Result()

	if err != nil {
		return FixedWindowResult{}, fmt.Errorf(
			"execute fixed window script: %w",
			err,
		)
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 4 {
		return FixedWindowResult{}, fmt.Errorf(
			"unexpected fixed window result",
		)
	}

	allowed, ok := values[0].(int64)
	if !ok {
		return FixedWindowResult{}, fmt.Errorf(
			"invalid fixed window allowed result",
		)
	}

	remaining, ok := values[1].(int64)
	if !ok {
		return FixedWindowResult{}, fmt.Errorf(
			"invalid fixed window remaining result",
		)
	}

	resultLimit, ok := values[2].(int64)
	if !ok {
		return FixedWindowResult{}, fmt.Errorf(
			"invalid fixed window limit result",
		)
	}

	retryAfterMs, ok := values[3].(int64)
	if !ok {
		return FixedWindowResult{}, fmt.Errorf(
			"invalid fixed window retry_after result",
		)
	}

	return FixedWindowResult{
		Allowed:      allowed == 1,
		Remaining:    int(remaining),
		Limit:        int(resultLimit),
		RetryAfterMs: retryAfterMs,
	}, nil
}
