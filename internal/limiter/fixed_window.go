package limiter

import (
	"context"
	"encoding/json"
	"fmt"

	redisstore "github.com/faizahmd2/rate-limitter-service/internal/redis"
)

type FixedWindowConfig struct {
	Limit         int    `json:"limit"`
	WindowSeconds int    `json:"window_seconds"`
	KeyStrategy   string `json:"key_strategy"`
}

func checkFixedWindow(
	ctx context.Context,
	redisClient *redisstore.Client,
	key string,
	configData []byte,
) (Decision, error) {

	var config FixedWindowConfig

	if err := json.Unmarshal(configData, &config); err != nil {
		return Decision{}, fmt.Errorf(
			"%w: invalid fixed window config: %v",
			ErrInternal,
			err,
		)
	}

	if config.Limit <= 0 {
		return Decision{}, fmt.Errorf(
			"%w: fixed window limit must be greater than zero",
			ErrInternal,
		)
	}

	if config.WindowSeconds <= 0 {
		return Decision{}, fmt.Errorf(
			"%w: fixed window window_seconds must be greater than zero",
			ErrInternal,
		)
	}

	result, err := redisstore.CheckFixedWindow(
		ctx,
		redisClient,
		key,
		config.Limit,
		config.WindowSeconds,
	)
	if err != nil {
		return Decision{}, fmt.Errorf(
			"%w: fixed window check failed: %v",
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
