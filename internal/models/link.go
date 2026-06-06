package models

import (
	"time"
)

type Link struct {
	ID          string    `json:"id" db:"id"`
	Alias       string    `json:"alias" db:"alias"`
	URL         string    `json:"url" db:"url"`
	Lifetime    *int      `json:"lifetime,omitempty" db:"lifetime"`
	IsDeleted   bool      `json:"is_deleted" db:"is_deleted"`
	IsDeactive  bool      `json:"is_deactive" db:"is_deactive"`
	IsPrivate   bool      `json:"is_private" db:"is_private"`
	IsSingle    bool      `json:"is_single" db:"is_single"`
	AccessToken *string   `json:"access_token,omitempty" db:"access_token"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type CreateLinkRequest struct {
	URL       string `json:"url" binding:"required"`
	Alias     string `json:"alias,omitempty"`
	Lifetime  *int   `json:"lifetime,omitempty"`
	IsPrivate bool   `json:"is_private"`
	IsSingle  bool   `json:"is_single"`
}

type CreateLinkResponse struct {
	Link        *Link  `json:"link"`
	AccessToken string `json:"access_token,omitempty"`
}

type UpdateLinkRequest struct {
	URL        *string `json:"url,omitempty"`
	Alias      *string `json:"alias,omitempty"`
	Lifetime   *int    `json:"lifetime,omitempty"`
	IsPrivate  *bool   `json:"is_private,omitempty"`
	IsSingle   *bool   `json:"is_single,omitempty"`
	IsDeactive *bool   `json:"is_deactive,omitempty"`
}
