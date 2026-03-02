package model

import "time"

// MediaReference tracks which external entity uses a media file.
type MediaReference struct {
	ID         int64     `db:"id"          json:"id"`
	MediaID    string    `db:"media_id"    json:"media_id"`
	RefService string    `db:"ref_service" json:"ref_service"`
	RefTable   string    `db:"ref_table"   json:"ref_table"`
	RefID      string    `db:"ref_id"      json:"ref_id"`
	RefField   string    `db:"ref_field"   json:"ref_field"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

// AddReferenceInput is the payload for adding a reference.
type AddReferenceInput struct {
	MediaID    string `json:"media_id"    binding:"required,uuid"`
	RefService string `json:"ref_service" binding:"required"`
	RefTable   string `json:"ref_table"   binding:"required"`
	RefID      string `json:"ref_id"      binding:"required"`
	RefField   string `json:"ref_field"   binding:"required"`
}

// RemoveReferenceInput is the payload for removing a reference.
type RemoveReferenceInput struct {
	MediaID    string `json:"media_id"    binding:"required,uuid"`
	RefService string `json:"ref_service" binding:"required"`
	RefTable   string `json:"ref_table"   binding:"required"`
	RefID      string `json:"ref_id"      binding:"required"`
	RefField   string `json:"ref_field"   binding:"required"`
}

// UsageResult wraps ref count and reference list.
type UsageResult struct {
	MediaID    string           `json:"media_id"`
	RefCount   int              `json:"ref_count"`
	References []MediaReference `json:"references"`
}
