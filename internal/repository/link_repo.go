package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/winchester/shorturls/internal/models"
)

type LinkRepo struct {
	pool *pgxpool.Pool
}

func NewLinkRepo(pool *pgxpool.Pool) *LinkRepo {
	return &LinkRepo{pool: pool}
}

func (r *LinkRepo) Create(ctx context.Context, link *models.Link) error {
	query := `
		INSERT INTO links (alias, url, lifetime, is_deleted, is_deactive, is_private, is_single, access_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		link.Alias, link.URL, link.Lifetime,
		link.IsDeleted, link.IsDeactive, link.IsPrivate, link.IsSingle, link.AccessToken,
	).Scan(&link.ID, &link.CreatedAt, &link.UpdatedAt)
}

func (r *LinkRepo) GetByAlias(ctx context.Context, alias string) (*models.Link, error) {
	link := &models.Link{}
	query := `
		SELECT id, alias, url, lifetime, is_deleted, is_deactive, is_private, is_single,
		       access_token, created_at, updated_at
		FROM links WHERE alias = $1`
	err := r.pool.QueryRow(ctx, query, alias).Scan(
		&link.ID, &link.Alias, &link.URL, &link.Lifetime,
		&link.IsDeleted, &link.IsDeactive, &link.IsPrivate, &link.IsSingle,
		&link.AccessToken, &link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get link: %w", err)
	}
	return link, nil
}

func (r *LinkRepo) GetByID(ctx context.Context, id string) (*models.Link, error) {
	link := &models.Link{}
	query := `
		SELECT id, alias, url, lifetime, is_deleted, is_deactive, is_private, is_single,
		       access_token, created_at, updated_at
		FROM links WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&link.ID, &link.Alias, &link.URL, &link.Lifetime,
		&link.IsDeleted, &link.IsDeactive, &link.IsPrivate, &link.IsSingle,
		&link.AccessToken, &link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get link by id: %w", err)
	}
	return link, nil
}

func (r *LinkRepo) Update(ctx context.Context, link *models.Link) error {
	query := `
		UPDATE links SET
			url = COALESCE($2, url),
			alias = COALESCE($3, alias),
			lifetime = COALESCE($4, lifetime),
			is_private = COALESCE($5, is_private),
			is_single = COALESCE($6, is_single),
			is_deactive = COALESCE($7, is_deactive),
			updated_at = $8
		WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query,
		link.ID, nullableString(link.URL), nullableString(link.Alias),
		link.Lifetime, nullableBool(link.IsPrivate), nullableBool(link.IsSingle),
		nullableBool(link.IsDeactive), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

func (r *LinkRepo) SoftDelete(ctx context.Context, alias string) error {
	query := `UPDATE links SET is_deleted = TRUE, updated_at = $2 WHERE alias = $1`
	tag, err := r.pool.Exec(ctx, query, alias, time.Now())
	if err != nil {
		return fmt.Errorf("soft delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}

func (r *LinkRepo) IsAliasTaken(ctx context.Context, alias string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM links WHERE alias = $1)`, alias).Scan(&exists)
	return exists, err
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableBool(b bool) *bool {
	return &b
}
