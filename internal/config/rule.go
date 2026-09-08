package config

import (
	"time"
)

type RateLimitRule struct {
	ID        string
	Service   string
	Resource  string
	Algorithm string
	Config    []byte
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
