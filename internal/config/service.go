package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RuleService struct {
	repository RuleRepository
	cache      *RuleCache
}

type UpdateRuleInput struct {
	Algorithm string
	Config    json.RawMessage
	Status    string
}

func NewRuleService(
	repository RuleRepository,
	cache *RuleCache,
) *RuleService {
	return &RuleService{
		repository: repository,
		cache:      cache,
	}
}

type CreateRuleInput struct {
	Service   string
	Resource  string
	Algorithm string
	Config    json.RawMessage
}

func (s *RuleService) CreateRule(
	ctx context.Context,
	input CreateRuleInput,
) (*RateLimitRule, error) {

	service := strings.TrimSpace(input.Service)
	resource := strings.TrimSpace(input.Resource)
	algorithm := strings.TrimSpace(input.Algorithm)

	if service == "" {
		return nil, errors.New("service is required")
	}

	if resource == "" {
		return nil, errors.New("resource is required")
	}

	if algorithm == "" {
		return nil, errors.New("algorithm is required")
	}

	if len(input.Config) == 0 {
		return nil, errors.New("config is required")
	}

	if !json.Valid(input.Config) {
		return nil, errors.New(
			"config must be valid JSON",
		)
	}

	if err := validateRuleConfig(
		algorithm,
		input.Config,
	); err != nil {
		return nil, err
	}

	rule := &RateLimitRule{
		ID:        uuid.New().String(),
		Service:   service,
		Resource:  resource,
		Algorithm: algorithm,
		Config:    input.Config,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repository.CreateRule(ctx, rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *RuleService) GetRules(
	ctx context.Context,
) ([]*RateLimitRule, error) {
	return s.repository.GetRules(ctx)
}

func (s *RuleService) GetRule(
	ctx context.Context,
	service string,
	resource string,
) (*RateLimitRule, error) {
	return s.repository.GetRule(
		ctx,
		service,
		resource,
	)
}

func (s *RuleService) UpdateRule(
	ctx context.Context,
	service string,
	resource string,
	input UpdateRuleInput,
) (*RateLimitRule, error) {

	service = strings.TrimSpace(service)
	resource = strings.TrimSpace(resource)
	algorithm := strings.TrimSpace(input.Algorithm)

	if service == "" {
		return nil, errors.New("service is required")
	}

	if resource == "" {
		return nil, errors.New("resource is required")
	}

	if algorithm == "" {
		return nil, errors.New("algorithm is required")
	}

	if len(input.Config) == 0 {
		return nil, errors.New("config is required")
	}

	if !json.Valid(input.Config) {
		return nil, errors.New("config must be valid JSON")
	}

	if err := validateRuleConfig(algorithm, input.Config); err != nil {
		return nil, err
	}

	rule, err := s.repository.GetRule(
		ctx,
		service,
		resource,
	)
	if err != nil {
		return nil, err
	}

	rule.Algorithm = algorithm
	rule.Config = input.Config

	if input.Status != "" {
		rule.Status = input.Status
	}

	rule.UpdatedAt = time.Now().UTC()

	if err := s.repository.UpdateRule(ctx, rule); err != nil {
		return nil, err
	}

	if err := s.cache.Invalidate(
		ctx,
		service,
		resource,
	); err != nil {
		return nil, fmt.Errorf("invalidate rule cache: %w", err)
	}

	return rule, nil
}

func (s *RuleService) DeleteRule(
	ctx context.Context,
	service string,
	resource string,
) error {
	service = strings.TrimSpace(service)
	resource = strings.TrimSpace(resource)

	if service == "" {
		return errors.New("service is required")
	}

	if resource == "" {
		return errors.New("resource is required")
	}

	err := s.repository.DeleteRule(
		ctx,
		service,
		resource,
	)

	if err != nil {
		return err
	}

	if err := s.cache.Invalidate(
		ctx,
		service,
		resource,
	); err != nil {
		return fmt.Errorf("invalidate rule cache: %w", err)
	}

	return nil
}
