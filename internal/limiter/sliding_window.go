package limiter

import (
	"context"
	"encoding/json"
	"fmt"

	redisstore "github.com/faizahmd2/rate-guard/internal/redis"
)

type SlidingWindowConfig struct {
	Limit         int    `json:"limit"`
	WindowSeconds int    `json:"window_seconds"`
	KeyStrategy   string `json:"key_strategy"`
}

func checkSlidingWindow(
	ctx context.Context,
	redisClient *redisstore.Client,
	key string,
	configData []byte,
) (Decision, error) {

	var config SlidingWindowConfig

	fmt.Printf("SLIDING CONFIG RAW: %s\n", configData)

	if err := json.Unmarshal(configData, &config); err != nil {
		return Decision{}, fmt.Errorf(
			"%w: invalid sliding window config: %v",
			ErrInternal,
			err,
		)
	}

	if config.Limit <= 0 {
		return Decision{}, fmt.Errorf(
			"%w: sliding window limit must be greater than zero",
			ErrInternal,
		)
	}

	if config.WindowSeconds <= 0 {
		return Decision{}, fmt.Errorf(
			"%w: sliding window window_seconds must be greater than zero",
			ErrInternal,
		)
	}

	result, err := redisstore.CheckSlidingWindow(
		ctx,
		redisClient,
		key,
		config.Limit,
		config.WindowSeconds,
	)

	if err != nil {
		return Decision{}, fmt.Errorf(
			"%w: sliding window check failed: %v",
			ErrServiceUnavailable,
			err,
		)
	}

	decision := DecisionAllow

	if !result.Allowed {
		decision = DecisionDeny
	}

	return Decision{
		Decision:     decision,
		Limit:        result.Limit,
		Remaining:    result.Remaining,
		RetryAfterMs: result.RetryAfterMs,
	}, nil
}
