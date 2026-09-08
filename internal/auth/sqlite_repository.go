package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrAdminNotConfigured = errors.New("admin not configured")
var ErrAdminAlreadyExists = errors.New("admin already exists")

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) GetAdmin(
	ctx context.Context,
) (*AdminUser, error) {
	var admin AdminUser

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			username,
			password_hash,
			created_at,
			updated_at
		FROM admin_users
		LIMIT 1
		`,
	).Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAdminNotConfigured
	}

	if err != nil {
		return nil, fmt.Errorf("get admin: %w", err)
	}

	return &admin, nil
}

func (r *SQLiteRepository) CreateAdmin(
	ctx context.Context,
	admin *AdminUser,
) error {
	var count int

	err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM admin_users`,
	).Scan(&count)

	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}

	if count > 0 {
		return ErrAdminAlreadyExists
	}

	_, err = r.db.ExecContext(
		ctx,
		`
		INSERT INTO admin_users (
			id,
			username,
			password_hash,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?)
		`,
		admin.ID,
		admin.Username,
		admin.PasswordHash,
		admin.CreatedAt,
		admin.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) UpdateAdminPassword(
	ctx context.Context,
	id string,
	passwordHash string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE admin_users
		SET
			password_hash = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		`,
		passwordHash,
		id,
	)

	if err != nil {
		return fmt.Errorf("update admin password: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) CreateToken(
	ctx context.Context,
	token *APIToken,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO api_tokens (
			id,
			name,
			client,
			token_hash,
			status,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		token.ID,
		token.Name,
		token.Client,
		token.TokenHash,
		token.Status,
		token.CreatedAt,
		token.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("create api token: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetActiveTokens(
	ctx context.Context,
) ([]APIToken, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			name,
			client,
			token_hash,
			status,
			created_at,
			updated_at,
			last_used_at
		FROM api_tokens
		WHERE status = 'active'
		`,
	)

	if err != nil {
		return nil, fmt.Errorf("get active tokens: %w", err)
	}
	defer rows.Close()

	var tokens []APIToken

	for rows.Next() {
		var token APIToken

		if err := rows.Scan(
			&token.ID,
			&token.Name,
			&token.Client,
			&token.TokenHash,
			&token.Status,
			&token.CreatedAt,
			&token.UpdatedAt,
			&token.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("scan api token: %w", err)
		}

		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api tokens: %w", err)
	}

	return tokens, nil
}

func (r *SQLiteRepository) RevokeToken(
	ctx context.Context,
	id string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		UPDATE api_tokens
		SET
			status = 'revoked',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("revoke api token: %w", err)
	}

	return nil
}
