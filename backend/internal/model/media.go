package model

import "time"

// MediaFile represents a stored media object.
type MediaFile struct {
	ID            string     `db:"id"             json:"id"`
	Bucket        string     `db:"bucket"         json:"bucket"`
	ObjectKey     string     `db:"object_key"     json:"object_key"`
	FileName      string     `db:"file_name"      json:"file_name"`
	MimeType      string     `db:"mime_type"      json:"mime_type"`
	Size          int64      `db:"size"           json:"size"`
	Checksum      string     `db:"checksum"       json:"checksum"`
	SourceService string     `db:"source_service" json:"source_service,omitempty"`
	SourceModule  string     `db:"source_module"  json:"source_module,omitempty"`
	UploadedBy    string     `db:"uploaded_by"    json:"uploaded_by"`
	IsPublic      bool       `db:"is_public"      json:"is_public"`
	RefCount      int        `db:"ref_count"      json:"ref_count"`
	URL           string     `db:"-"              json:"url,omitempty"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
	DeletedAt     *time.Time `db:"deleted_at"     json:"deleted_at,omitempty"`
}

// ListParams for paginated media queries.
type ListParams struct {
	Page          int
	Limit         int
	Pagination    string // "offset" (default) | "cursor"
	Cursor        string
	MimeGroup     string // "image", "video", "audio", "document", ""
	Search        string
	UploadedBy    string
	SourceService string
	SourceModule  string
	SortBy        string // "created_at", "size", "file_name"
	SortDir       string // "asc", "desc"
}

// ListResult wraps a page of media files.
type ListResult struct {
	Items      []*MediaFile `json:"items"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"total_pages"`
	HasMore    bool         `json:"has_more,omitempty"`
	NextCursor string       `json:"next_cursor,omitempty"`
}

// FilterOptions contains distinct values for media filters.
type FilterOptions struct {
	SourceServices []string `json:"source_services"`
	SourceModules  []string `json:"source_modules"`
}
