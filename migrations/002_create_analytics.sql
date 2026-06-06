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
