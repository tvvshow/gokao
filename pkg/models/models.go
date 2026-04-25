package models

import "time"

// IDModel provides a minimal shared base model for cross-service payloads.
type IDModel struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PageRequest is a shared pagination request shape.
type PageRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PageResponse is a shared pagination response shape.
type PageResponse[T any] struct {
	Items      []T `json:"items"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}
