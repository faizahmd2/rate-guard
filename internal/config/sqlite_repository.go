package config

import (
	"context"
	"database/sql"
	"errors"
)

type SQLiteRuleRepository struct {
	db *sql.DB
}

func NewSQLiteRuleRepository(
	db *sql.DB,
) *SQLiteRuleRepository {
	return &SQLiteRuleRepository{
		db: db,
	}
}

func (r *SQLiteRuleRepository) GetRule(
	ctx context.Context,
	service string,
	resource string,
) (*RateLimitRule, error) {

	var rule RateLimitRule

	err := r.db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			service,
			resource,
			algorithm,
			config,
			status,
			created_at,
			updated_at
		FROM rate_limit_rules
		WHERE service = ?
		  AND resource = ?
		  AND status = 'active'
		`,
		service,
		resource,
	).Scan(
		&rule.ID,
		&rule.Service,
		&rule.Resource,
		&rule.Algorithm,
		&rule.Config,
		&rule.Status,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRuleNotFound
	}

	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *SQLiteRuleRepository) CreateRule(
	ctx context.Context,
	rule *RateLimitRule,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`
		INSERT INTO rate_limit_rules (
			id,
			service,
			resource,
			algorithm,
			config,
			status,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
		rule.ID,
		rule.Service,
		rule.Resource,
		rule.Algorithm,
		rule.Config,
		rule.Status,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	return err
}

func (r *SQLiteRuleRepository) GetRules(
	ctx context.Context,
) ([]*RateLimitRule, error) {

	rows, err := r.db.QueryContext(
		ctx,
		`
		SELECT
			id,
			service,
			resource,
			algorithm,
			config,
			status,
			created_at,
			updated_at
		FROM rate_limit_rules
		ORDER BY created_at DESC
		`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rules []*RateLimitRule

	for rows.Next() {
		var rule RateLimitRule

		if err := rows.Scan(
			&rule.ID,
			&rule.Service,
			&rule.Resource,
			&rule.Algorithm,
			&rule.Config,
			&rule.Status,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, err
		}

		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *SQLiteRuleRepository) UpdateRule(
	ctx context.Context,
	rule *RateLimitRule,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`
		UPDATE rate_limit_rules
		SET algorithm = ?,
		    config = ?,
		    status = ?,
		    updated_at = ?
		WHERE service = ?
		  AND resource = ?
		`,
		rule.Algorithm,
		rule.Config,
		rule.Status,
		rule.UpdatedAt,
		rule.Service,
		rule.Resource,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrRuleNotFound
	}

	return nil
}

func (r *SQLiteRuleRepository) DeleteRule(
	ctx context.Context,
	service string,
	resource string,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`
		DELETE FROM rate_limit_rules
		WHERE service = ?
		  AND resource = ?
		`,
		service,
		resource,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrRuleNotFound
	}

	return nil
}
