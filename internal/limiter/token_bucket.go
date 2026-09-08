package limiter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/faizahmd2/rate-guard/internal/redis"
)

type TokenBucketConfig struct {
	Capacity    float64 `json:"capacity"`
	RefillRate  float64 `json:"refill_rate"`
	KeyStrategy string  `json:"key_strategy"`
}

func checkTokenBucket(
	ctx context.Context,
	redisClient *redis.Client,
	ruleKey string,
	configData []byte,
) (Decision, error) {

	var config TokenBucketConfig

	if err := json.Unmarshal(configData, &config); err != nil {
		return Decision{}, fmt.Errorf(
			"invalid token bucket config: %w",
			err,
		)
	}

	if config.Capacity <= 0 {
		return Decision{}, fmt.Errorf(
			"token bucket capacity must be greater than zero",
		)
	}

	if config.RefillRate <= 0 {
		return Decision{}, fmt.Errorf(
			"token bucket refill rate must be greater than zero",
		)
	}

	result, err := redisClient.CheckTokenBucket(
		ctx,
		ruleKey,
		config.Capacity,
		config.RefillRate,
		1,
	)
	if err != nil {
		return Decision{}, fmt.Errorf(
			"%w: %v",
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
