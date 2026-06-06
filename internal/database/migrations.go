package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationCreateLinks = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alias VARCHAR(255) NOT NULL UNIQUE,
    url TEXT NOT NULL,
    lifetime INTEGER,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    is_deactive BOOLEAN NOT NULL DEFAULT FALSE,
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    is_single BOOLEAN NOT NULL DEFAULT FALSE,
    access_token TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationCreateAnalytics = `
CREATE TABLE IF NOT EXISTS analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL UNIQUE REFERENCES links(id) ON DELETE CASCADE,
    success_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    first_visit_at TIMESTAMPTZ,
    last_visit_at TIMESTAMPTZ,
    browser_stats JSONB NOT NULL DEFAULT '{}',
    qr_scan_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationAddIndexes = `
CREATE INDEX IF NOT EXISTS idx_links_alias ON links(alias);
CREATE INDEX IF NOT EXISTS idx_links_is_deleted ON links(is_deleted);
CREATE INDEX IF NOT EXISTS idx_links_is_deactive ON links(is_deactive);
CREATE INDEX IF NOT EXISTS idx_analytics_success_count ON analytics(success_count DESC);
`

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []struct {
		name string
		sql  string
	}{
		{"001_create_links", migrationCreateLinks},
		{"002_create_analytics", migrationCreateAnalytics},
		{"003_add_indexes", migrationAddIndexes},
	}

	for _, m := range migrations {
		if _, err := pool.Exec(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
		fmt.Printf("migration applied: %s\n", m.name)
	}
	return nil
}
