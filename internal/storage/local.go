package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type localStorage struct {
	basePath string
}

func (s *localStorage) Save(ctx context.Context, id int64, r io.Reader) (string, error) {
	lr := io.LimitReader(r, maxFileSize+1) // what is this?

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

func NewLocalStorage(basePath string) ImageStore {
	return &localStorage{basePath: basePath}
}
