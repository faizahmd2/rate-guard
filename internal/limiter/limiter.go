package limiter

import (
	"context"
	"errors"
	"fmt"

	"github.com/faizahmd2/rate-guard/internal/config"
	redisstore "github.com/faizahmd2/rate-guard/internal/redis"
)

type Service struct {
	ruleCache   *config.RuleCache
	redisClient *redisstore.Client
}

func NewService(
	ruleCache *config.RuleCache,
	redisClient *redisstore.Client,
) *Service {
	return &Service{
		ruleCache:   ruleCache,
		redisClient: redisClient,
	}
}

func (s *Service) Check(
	ctx context.Context,
	input CheckInput,
) (Decision, error) {

	rule, err := s.ruleCache.Get(
		ctx,
		input.Service,
		input.Resource,
	)
	if err != nil {
		if errors.Is(err, config.ErrRuleNotFound) {
			return Decision{}, err
		}

		return Decision{}, fmt.Errorf(
			"%w: %v",
			ErrServiceUnavailable,
			err,
		)
	}

	switch rule.Algorithm {

	case "token_bucket":
		return checkTokenBucket(
			ctx,
			s.redisClient,
			bucketKey(
				rule.Service,
				rule.Resource,
				input.Key,
			),
			rule.Config,
		)

	case "fixed_window":
		return checkFixedWindow(
			ctx,
			s.redisClient,
			bucketKey(
				rule.Service,
				rule.Resource,
				input.Key,
			),
			rule.Config,
		)

	case "sliding_window":
		return checkSlidingWindow(
			ctx,
			s.redisClient,
			bucketKey(
				rule.Service,
				rule.Resource,
				input.Key,
			),
			rule.Config,
		)

	default:
		return Decision{}, fmt.Errorf(
			"%w: %v",
			ErrInternal,
			rule.Algorithm,
		)
	}
}

func bucketKey(
	service string,
	resource string,
	key string,
) string {
	return fmt.Sprintf(
		"rateguard:bucket:%s:%s:%s",
		service,
		resource,
		key,
	)
}
