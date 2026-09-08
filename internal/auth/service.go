package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSetupCompleted     = errors.New("setup already completed")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Setup(
	ctx context.Context,
	username string,
	password string,
) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.New("username is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	_, err := s.repository.GetAdmin(ctx)

	if err == nil {
		return ErrSetupCompleted
	}

	if !errors.Is(err, ErrAdminNotConfigured) {
		return fmt.Errorf("check admin: %w", err)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()

	admin := &AdminUser{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repository.CreateAdmin(ctx, admin); err != nil {
		if errors.Is(err, ErrAdminAlreadyExists) {
			return ErrSetupCompleted
		}

		return err
	}

	return nil
}

func (s *Service) Login(
	ctx context.Context,
	username string,
	password string,
) error {
	admin, err := s.repository.GetAdmin(ctx)
	if err != nil {
		return ErrInvalidCredentials
	}

	if admin.Username != username {
		return ErrInvalidCredentials
	}

	if !CheckPassword(admin.PasswordHash, password) {
		return ErrInvalidCredentials
	}

	return nil
}

func (s *Service) GetAdmin(
	ctx context.Context,
) (*AdminUser, error) {
	return s.repository.GetAdmin(ctx)
}

func (s *Service) CreateToken(
	ctx context.Context,
	name string,
	client string,
) (APIToken, string, error) {
	name = strings.TrimSpace(name)
	client = strings.TrimSpace(client)

	if name == "" {
		return APIToken{}, "", errors.New("token name is required")
	}

	if client == "" {
		return APIToken{}, "", errors.New("client is required")
	}

	plaintext, tokenHash, err := GenerateAPIToken()
	if err != nil {
		return APIToken{}, "", err
	}

	now := time.Now()

	token := APIToken{
		ID:         uuid.NewString(),
		Name:       name,
		Client:     client,
		TokenHash:  tokenHash,
		Status:     "active",
		CreatedAt:  now,
		UpdatedAt:  now,
		LastUsedAt: nil,
	}

	if err := s.repository.CreateToken(ctx, &token); err != nil {
		return APIToken{}, "", fmt.Errorf("create token: %w", err)
	}

	return token, plaintext, nil
}

func (s *Service) ListTokens(
	ctx context.Context,
) ([]APIToken, error) {
	tokens, err := s.repository.GetActiveTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active tokens: %w", err)
	}

	return tokens, nil
}

func (s *Service) RevokeToken(
	ctx context.Context,
	id string,
) error {
	if err := s.repository.RevokeToken(ctx, id); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
}
