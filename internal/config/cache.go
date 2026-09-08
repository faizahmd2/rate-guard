package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/faizahmd2/rate-limitter-service/internal/redis"

	"golang.org/x/sync/singleflight"
)

type cachedRule struct {
	rule      *RateLimitRule
	expiresAt time.Time
}

type RuleCache struct {
	mu    sync.RWMutex
	rules map[string]cachedRule

	redis      *redis.Client
	repository RuleRepository

	l1TTL time.Duration
	l2TTL time.Duration

	sf singleflight.Group
}

func NewRuleCache(
	redisClient *redis.Client,
	repository RuleRepository,
	l1TTL time.Duration,
	l2TTL time.Duration,
) *RuleCache {
	return &RuleCache{
		rules:      make(map[string]cachedRule),
		redis:      redisClient,
		repository: repository,
		l1TTL:      l1TTL,
		l2TTL:      l2TTL,
	}
}

func (c *RuleCache) Get(
	ctx context.Context,
	service string,
	resource string,
) (*RateLimitRule, error) {

	key := ruleKey(service, resource)

	// L1: in-memory cache
	if rule, ok := c.getL1(key); ok {
		return rule, nil
	}

	// L1 miss.
	// Only one goroutine should continue to Redis/DB
	// for the same rule.
	value, err, _ := c.sf.Do(key, func() (interface{}, error) {
		// Another request may have populated L1 while
		// this goroutine was waiting for the singleflight.
		if rule, ok := c.getL1(key); ok {
			return rule, nil
		}

		// L2: Redis
		rule, err := c.getL2(ctx, key)
		if err != nil {
			return nil, err
		}

		if rule != nil {
			c.setL1(key, rule)
			return rule, nil
		}

		// L3: PostgreSQL
		rule, err = c.repository.GetRule(ctx, service, resource)
		if err != nil {
			return nil, err
		}

		if rule == nil {
			return nil, fmt.Errorf(
				"rate limit rule not found: %s/%s",
				service,
				resource,
			)
		}

		// Populate L2.
		if err := c.setL2(ctx, key, rule); err != nil {
			return nil, err
		}

		// Populate L1.
		c.setL1(key, rule)

		return rule, nil
	})

	if err != nil {
		return nil, err
	}

	return value.(*RateLimitRule), nil
}

func (c *RuleCache) getL1(key string) (*RateLimitRule, bool) {
	c.mu.RLock()
	entry, ok := c.rules[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.rules, key)
		c.mu.Unlock()

		return nil, false
	}

	return entry.rule, true
}

func (c *RuleCache) setL1(key string, rule *RateLimitRule) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.rules[key] = cachedRule{
		rule:      rule,
		expiresAt: time.Now().Add(c.l1TTL),
	}
}

func (c *RuleCache) getL2(
	ctx context.Context,
	key string,
) (*RateLimitRule, error) {

	value, err := c.redis.Get(ctx, key)

	if err == redis.ErrKeyNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("redis get rule: %w", err)
	}

	var rule RateLimitRule

	if err := json.Unmarshal([]byte(value), &rule); err != nil {
		return nil, fmt.Errorf("decode cached rule: %w", err)
	}

	return &rule, nil
}

func (c *RuleCache) setL2(
	ctx context.Context,
	key string,
	rule *RateLimitRule,
) error {

	value, err := json.Marshal(rule)
	if err != nil {
		return fmt.Errorf("encode rule for redis: %w", err)
	}

	if err := c.redis.Set(ctx, key, value, c.l2TTL); err != nil {
		return fmt.Errorf("redis set rule: %w", err)
	}

	return nil
}

func (c *RuleCache) Invalidate(
	ctx context.Context,
	service string,
	resource string,
) error {
	key := ruleKey(service, resource)

	// Remove from Redis L2 cache.
	if err := c.redis.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete cached rule from redis: %w", err)
	}

	// Remove from local L1 cache.
	c.mu.Lock()
	delete(c.rules, key)
	c.mu.Unlock()

	return nil
}

func ruleKey(service, resource string) string {
	return fmt.Sprintf(
		"rateguard:rule:%s:%s",
		service,
		resource,
	)
}
