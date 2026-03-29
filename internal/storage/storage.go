package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrInvalidFormat = errors.New("unsupported image format")
	ErrTooLarge      = errors.New("file exceeds max size")

	allowedTypes = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	// 5MB
	maxFileSize = int64(5 << 20)
)

func IsInvalidFormat(err error) bool {
    return errors.Is(err, ErrInvalidFormat)
}

func IsTooLarge(err error) bool {
    return errors.Is(err, ErrTooLarge)
}

type ImageStore interface {
	Save(ctx context.Context, id int64, r io.Reader) (string, error)
}
