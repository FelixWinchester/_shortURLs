package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/winchester/shorturls/internal/models"
)

type AnalyticsRepo struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepo(pool *pgxpool.Pool) *AnalyticsRepo {
	return &AnalyticsRepo{pool: pool}
}

func (r *AnalyticsRepo) EnsureExists(ctx context.Context, linkID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO analytics (link_id)
		VALUES ($1)
		ON CONFLICT (link_id) DO NOTHING`, linkID)
	return err
}

func (r *AnalyticsRepo) RecordSuccess(ctx context.Context, linkID, browser string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE analytics SET
			success_count = success_count + 1,
			last_visit_at = $2,
			first_visit_at = COALESCE(first_visit_at, $2),
			browser_stats = browser_stats || jsonb_build_object($3::text, COALESCE((browser_stats->>$3::text)::int, 0) + 1),
			updated_at = $2
		WHERE link_id = $1`, linkID, now, browser)
	return err
}

func (r *AnalyticsRepo) RecordError(ctx context.Context, linkID, browser string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE analytics SET
			error_count = error_count + 1,
			last_visit_at = $2,
			first_visit_at = COALESCE(first_visit_at, $2),
			browser_stats = browser_stats || jsonb_build_object($3::text, COALESCE((browser_stats->>$3::text)::int, 0) + 1),
			updated_at = $2
		WHERE link_id = $1`, linkID, now, browser)
	return err
}

func (r *AnalyticsRepo) DeactivateSingleUse(ctx context.Context, linkID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE links SET is_deactive = TRUE, updated_at = $2 WHERE id = $1 AND is_single = TRUE`, linkID, time.Now())
	return err
}

func (r *AnalyticsRepo) GetByLinkID(ctx context.Context, linkID string) (*models.Analytics, error) {
	a := &models.Analytics{}
	query := `
		SELECT id, link_id, success_count, error_count, first_visit_at, last_visit_at,
		       browser_stats, qr_scan_count, created_at, updated_at
		FROM analytics WHERE link_id = $1`
	err := r.pool.QueryRow(ctx, query, linkID).Scan(
		&a.ID, &a.LinkID, &a.SuccessCount, &a.ErrorCount,
		&a.FirstVisitAt, &a.LastVisitAt, &a.BrowserStats,
		&a.QRScanCount, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get analytics: %w", err)
	}
	return a, nil
}

func (r *AnalyticsRepo) GetTopLinks(ctx context.Context, limit int) ([]models.AnalyticsSummary, error) {
	query := `
		SELECT a.link_id, l.alias, l.url, a.success_count, a.error_count,
		       a.first_visit_at, a.last_visit_at, a.browser_stats, a.qr_scan_count
		FROM analytics a
		JOIN links l ON l.id = a.link_id
		WHERE l.is_deleted = FALSE
		ORDER BY a.success_count DESC
		LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get top links: %w", err)
	}
	defer rows.Close()

	var result []models.AnalyticsSummary
	for rows.Next() {
		var s models.AnalyticsSummary
		var fv, lv *time.Time
		if err := rows.Scan(&s.LinkID, &s.Alias, &s.URL, &s.SuccessCount, &s.ErrorCount,
			&fv, &lv, &s.BrowserStats, &s.QRScanCount); err != nil {
			return nil, fmt.Errorf("scan top link: %w", err)
		}
		if fv != nil {
			v := fv.Format(time.RFC3339)
			s.FirstVisitAt = &v
		}
		if lv != nil {
			v := lv.Format(time.RFC3339)
			s.LastVisitAt = &v
		}
		result = append(result, s)
	}
	return result, nil
}

func (r *AnalyticsRepo) GetTotals(ctx context.Context) (totalSuccess, totalError int, browserStats json.RawMessage, err error) {
	row := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(success_count), 0), COALESCE(SUM(error_count), 0)
		FROM analytics a
		JOIN links l ON l.id = a.link_id
		WHERE l.is_deleted = FALSE`)
	if err := row.Scan(&totalSuccess, &totalError); err != nil {
		return 0, 0, nil, fmt.Errorf("get totals: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT a.browser_stats
		FROM analytics a
		JOIN links l ON l.id = a.link_id
		WHERE l.is_deleted = FALSE`)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("get browser stats: %w", err)
	}
	defer rows.Close()

	merged := make(map[string]int)
	for rows.Next() {
		var bs json.RawMessage
		if err := rows.Scan(&bs); err != nil {
			continue
		}
		var m map[string]int
		if json.Unmarshal(bs, &m) == nil {
			for k, v := range m {
				merged[k] += v
			}
		}
	}

	mergedJSON, _ := json.Marshal(merged)
	return totalSuccess, totalError, mergedJSON, nil
}

func (r *AnalyticsRepo) RecordQRScan(ctx context.Context, linkID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE analytics SET qr_scan_count = qr_scan_count + 1, updated_at = $2
		WHERE link_id = $1`, linkID, time.Now())
	return err
}
