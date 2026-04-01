package storage

import (
	"context"
	"errors"
	"io"
)

var (
	// ErrInvalidFormat is returned when the uploaded file is not a supported image type.
	ErrInvalidFormat = errors.New("unsupported image format")
	// ErrTooLarge is returned when the uploaded file exceeds the max allowed size.
	ErrTooLarge = errors.New("file exceeds max size")

	allowedTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	// 5MB
	maxFileSize = int64(5 << 20)
)

// IsInvalidFormat checks if the error is an unsupported image format error.
func IsInvalidFormat(err error) bool {
    return errors.Is(err, ErrInvalidFormat)
}

// IsTooLarge checks if the error is a file size exceeded error.
func IsTooLarge(err error) bool {
    return errors.Is(err, ErrTooLarge)
}

// ImageStore defines the interface for storing and retrieving book cover images.
type ImageStore interface {
	Save(ctx context.Context, id int64, r io.Reader) (string, error)
	// Delete removes the cover image for the given book ID.
	Delete(ctx context.Context, id int64) error
}
