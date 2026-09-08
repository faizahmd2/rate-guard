package config

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRuleRepository struct {
	db *pgxpool.Pool
}

var ErrRuleNotFound = errors.New("rate limit rule not found")

func NewPostgresRuleRepository(
	db *pgxpool.Pool,
) *PostgresRuleRepository {
	return &PostgresRuleRepository{
		db: db,
	}
}

func (r *PostgresRuleRepository) GetRule(
	ctx context.Context,
	service string,
	resource string,
) (*RateLimitRule, error) {

	var rule RateLimitRule

	err := r.db.QueryRow(
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
		WHERE service = $1
		  AND resource = $2
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

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}

	if err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *PostgresRuleRepository) CreateRule(
	ctx context.Context,
	rule *RateLimitRule,
) error {
	_, err := r.db.Exec(
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
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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

func (r *PostgresRuleRepository) GetRules(
	ctx context.Context,
) ([]*RateLimitRule, error) {

	rows, err := r.db.Query(
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

func (r *PostgresRuleRepository) UpdateRule(
	ctx context.Context,
	rule *RateLimitRule,
) error {
	tag, err := r.db.Exec(
		ctx,
		`
		UPDATE rate_limit_rules
		SET algorithm = $1,
		    config = $2,
		    status = $3,
		    updated_at = $4
		WHERE service = $5
		  AND resource = $6
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

	if tag.RowsAffected() == 0 {
		return ErrRuleNotFound
	}

	return nil
}

func (r *PostgresRuleRepository) DeleteRule(
	ctx context.Context,
	service string,
	resource string,
) error {
	tag, err := r.db.Exec(
		ctx,
		`
		DELETE FROM rate_limit_rules
		WHERE service = $1
		  AND resource = $2
		`,
		service,
		resource,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrRuleNotFound
	}

	return nil
}
