package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// localStorage is the filesystem implementation of ImageStore.
type localStorage struct {
	basePath string
}

// Save validates and stores an image on disk.
func (s *localStorage) Save(ctx context.Context, id int64, r io.Reader) (string, error) {
	// We make use of LimitReader to prevent reading files that are too big into memory. With 
	// LimitReader we read up until max allowable file size +1 byte. The idea with the +1 is to check if there is more data than allowed.
	lr := io.LimitReader(r, maxFileSize+1)

	buf, err := io.ReadAll(lr)
	if err != nil {
		return "", err
	}

	if int64(len(buf)) > maxFileSize {
		return "", ErrTooLarge
	}

	contentType := http.DetectContentType(buf)
	ext, ok := allowedTypes[contentType]
	if !ok {
		return "", ErrInvalidFormat
	}

	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return "", err
	}

	// Remove existing covers for this ID
	matches, _ := filepath.Glob(filepath.Join(s.basePath, fmt.Sprintf("%d.*", id)))
	for _, m := range matches {
		os.Remove(m)
	}

	filename := fmt.Sprintf("%d%s", id, ext)
	path := filepath.Join(s.basePath, filename)

	if err := os.WriteFile(path, buf, 0644); err != nil {
		return "", err
	}

	return path, nil
}

// Delete removes all cover images for the given book ID.
func (s *localStorage) Delete(_ context.Context, id int64) error {
	matches, _ := filepath.Glob(filepath.Join(s.basePath, fmt.Sprintf("%d.*", id)))
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			return err
		}
	}
	return nil
}

// NewLocalStorage creates a new ImageStore that saves files to the given directory.
func NewLocalStorage(basePath string) ImageStore {
	return &localStorage{basePath: basePath}
}
