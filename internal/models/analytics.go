package models

import (
	"encoding/json"
	"time"
)

type Analytics struct {
	ID           string           `json:"id" db:"id"`
	LinkID       string           `json:"link_id" db:"link_id"`
	SuccessCount int              `json:"success_count" db:"success_count"`
	ErrorCount   int              `json:"error_count" db:"error_count"`
	FirstVisitAt *time.Time       `json:"first_visit_at,omitempty" db:"first_visit_at"`
	LastVisitAt  *time.Time       `json:"last_visit_at,omitempty" db:"last_visit_at"`
	BrowserStats json.RawMessage  `json:"browser_stats" db:"browser_stats"`
	QRScanCount  int              `json:"qr_scan_count" db:"qr_scan_count"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

type AnalyticsSummary struct {
	LinkID          string `json:"link_id"`
	Alias           string `json:"alias"`
	URL             string `json:"url"`
	SuccessCount    int    `json:"success_count"`
	ErrorCount      int    `json:"error_count"`
	FirstVisitAt    *string `json:"first_visit_at,omitempty"`
	LastVisitAt     *string `json:"last_visit_at,omitempty"`
	BrowserStats    json.RawMessage `json:"browser_stats"`
	QRScanCount     int    `json:"qr_scan_count"`
}

type GlobalAnalytics struct {
	TopLinks      []AnalyticsSummary `json:"top_links"`
	TotalSuccess  int                `json:"total_success"`
	TotalError    int                `json:"total_error"`
	BrowserStats  json.RawMessage    `json:"browser_stats"`
}
