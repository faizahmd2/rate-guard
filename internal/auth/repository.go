package auth

import "context"

type Repository interface {
	GetAdmin(ctx context.Context) (*AdminUser, error)
	CreateAdmin(ctx context.Context, admin *AdminUser) error
	UpdateAdminPassword(
		ctx context.Context,
		id string,
		passwordHash string,
	) error

	CreateToken(ctx context.Context, token *APIToken) error
	GetActiveTokens(ctx context.Context) ([]APIToken, error)
	RevokeToken(ctx context.Context, id string) error
}
