package config

import (
	"encoding/json"
	"errors"
	"fmt"
)

func validateRuleConfig(
	algorithm string,
	rawConfig json.RawMessage,
) error {
	switch algorithm {
	case "token_bucket":
		return validateTokenBucketConfig(rawConfig)

	case "fixed_window":
		return validateFixedWindowConfig(rawConfig)

	case "sliding_window":
		return validateSlidingWindowConfig(rawConfig)

	default:
		return fmt.Errorf(
			"unsupported algorithm: %s",
			algorithm,
		)
	}
}

func validateTokenBucketConfig(
	rawConfig json.RawMessage,
) error {
	var config struct {
		Capacity    int     `json:"capacity"`
		RefillRate  float64 `json:"refill_rate"`
		KeyStrategy string  `json:"key_strategy"`
	}

	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return errors.New("invalid token_bucket config")
	}

	if config.Capacity <= 0 {
		return errors.New(
			"capacity must be greater than 0",
		)
	}

	if config.RefillRate <= 0 {
		return errors.New(
			"refill_rate must be greater than 0",
		)
	}

	if config.KeyStrategy == "" {
		return errors.New(
			"key_strategy is required",
		)
	}

	return validateKeyStrategy(config.KeyStrategy)
}

func validateFixedWindowConfig(
	rawConfig json.RawMessage,
) error {
	var config struct {
		Limit         int    `json:"limit"`
		WindowSeconds int    `json:"window_seconds"`
		KeyStrategy   string `json:"key_strategy"`
	}

	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return errors.New("invalid fixed_window config")
	}

	if config.Limit <= 0 {
		return errors.New(
			"limit must be greater than 0",
		)
	}

	if config.WindowSeconds <= 0 {
		return errors.New(
			"window_seconds must be greater than 0",
		)
	}

	if config.KeyStrategy == "" {
		return errors.New(
			"key_strategy is required",
		)
	}

	return validateKeyStrategy(config.KeyStrategy)
}

func validateSlidingWindowConfig(
	rawConfig json.RawMessage,
) error {
	var config struct {
		Limit         int    `json:"limit"`
		WindowSeconds int    `json:"window_seconds"`
		KeyStrategy   string `json:"key_strategy"`
	}

	if err := json.Unmarshal(rawConfig, &config); err != nil {
		return errors.New("invalid sliding_window config")
	}

	if config.Limit <= 0 {
		return errors.New(
			"limit must be greater than 0",
		)
	}

	if config.WindowSeconds <= 0 {
		return errors.New(
			"window_seconds must be greater than 0",
		)
	}

	if config.KeyStrategy == "" {
		return errors.New(
			"key_strategy is required",
		)
	}

	return validateKeyStrategy(config.KeyStrategy)
}

func validateKeyStrategy(
	keyStrategy string,
) error {
	switch keyStrategy {
	case "account":
		return nil

	default:
		return fmt.Errorf(
			"unsupported key_strategy: %s",
			keyStrategy,
		)
	}
}
