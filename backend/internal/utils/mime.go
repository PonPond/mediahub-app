package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// defaultAllowedMIMEs is kept as a reusable strict allow-list when
// callers explicitly pass it in.
var defaultAllowedMIMEs = []string{
	// Images
	"image/jpeg", "image/png", "image/gif", "image/webp",
	"image/svg+xml", "image/bmp", "image/tiff",
	// Video
	"video/mp4", "video/webm", "video/ogg", "video/quicktime",
	"video/x-msvideo", "video/mpeg",
	// Audio
	"audio/mpeg", "audio/ogg", "audio/wav", "audio/webm",
	"audio/aac", "audio/flac",
	// Documents
	"application/pdf",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/vnd.ms-excel",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"application/vnd.ms-powerpoint",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"text/plain", "text/csv", "text/html", "text/markdown",
	// Archives
	"application/zip", "application/x-tar", "application/gzip",
	"application/x-7z-compressed", "application/x-rar-compressed",
}

// ValidateMIME checks whether the given MIME type is allowed.
// If allowed is nil/empty, all MIME types are accepted.
func ValidateMIME(mimeType string, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	for _, a := range allowed {
		if a == mime {
			return nil
		}
	}
	return fmt.Errorf("MIME type %q is not allowed", mimeType)
}

// MIMEGroup returns a human-readable category for a given MIME type.
func MIMEGroup(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case mimeType == "application/pdf" ||
		strings.HasPrefix(mimeType, "text/") ||
		strings.Contains(mimeType, "document") ||
		strings.Contains(mimeType, "spreadsheet") ||
		strings.Contains(mimeType, "presentation"):
		return "document"
	default:
		return "other"
	}
}

func UploadGroup(fileName, mimeType string) string {
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "image"
	case strings.HasPrefix(mime, "video/"):
		return "video"
	case strings.HasPrefix(mime, "audio/"):
		return "audio"
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if strings.Contains(mime, "zip") ||
		strings.Contains(mime, "tar") ||
		strings.Contains(mime, "gzip") ||
		strings.Contains(mime, "7z") ||
		strings.Contains(mime, "rar") ||
		ext == ".zip" || ext == ".tar" || ext == ".gz" || ext == ".tgz" || ext == ".7z" || ext == ".rar" {
		return "archive"
	}

	if mime == "application/pdf" ||
		strings.HasPrefix(mime, "text/") ||
		strings.Contains(mime, "document") ||
		strings.Contains(mime, "spreadsheet") ||
		strings.Contains(mime, "presentation") ||
		ext == ".sql" || ext == ".csv" || ext == ".doc" || ext == ".docx" || ext == ".xls" || ext == ".xlsx" || ext == ".ppt" || ext == ".pptx" {
		return "document"
	}

	return "other"
}
