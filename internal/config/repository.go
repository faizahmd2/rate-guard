package config

import "context"

type RuleRepository interface {
	GetRule(
		ctx context.Context,
		service string,
		resource string,
	) (*RateLimitRule, error)

	GetRules(
		ctx context.Context,
	) ([]*RateLimitRule, error)

	CreateRule(
		ctx context.Context,
		rule *RateLimitRule,
	) error

	UpdateRule(
		ctx context.Context,
		rule *RateLimitRule,
	) error

	DeleteRule(
		ctx context.Context,
		service string,
		resource string,
	) error
}
