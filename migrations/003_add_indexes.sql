CREATE INDEX IF NOT EXISTS idx_links_alias ON links(alias);
CREATE INDEX IF NOT EXISTS idx_links_is_deleted ON links(is_deleted);
CREATE INDEX IF NOT EXISTS idx_links_is_deactive ON links(is_deactive);
CREATE INDEX IF NOT EXISTS idx_analytics_success_count ON analytics(success_count DESC);
