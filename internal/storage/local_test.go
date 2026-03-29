package storage_test

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Velin-Todorov/zetta-task/internal/storage"
)

func createJPEG(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, jpeg.Encode(buf, img, nil))
	return buf
}

func createPNG(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	require.NoError(t, png.Encode(buf, img))
	return buf
}

func TestSave_JPEG(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewLocalStorage(dir)

	path, err := store.Save(t.Context(), 1, createJPEG(t))

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "1.jpg"), path)

	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestSave_PNG(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewLocalStorage(dir)

	path, err := store.Save(t.Context(), 2, createPNG(t))

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "2.png"), path)

	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestSave_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewLocalStorage(dir)

	_, err := store.Save(t.Context(), 1, strings.NewReader("not an image"))

	assert.Error(t, err)
	assert.True(t, storage.IsInvalidFormat(err))
}

func TestSave_TooLarge(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewLocalStorage(dir)

	// 5MB + 1 byte
	largeData := make([]byte, 5<<20+1)
	_, err := store.Save(t.Context(), 1, bytes.NewReader(largeData))

	assert.Error(t, err)
	assert.True(t, storage.IsTooLarge(err))
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	store := storage.NewLocalStorage(dir)

	path, err := store.Save(t.Context(), 1, createJPEG(t))

	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "1.jpg"), path)
}

func TestSave_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	store := storage.NewLocalStorage(dir)

	path1, err := store.Save(t.Context(), 1, createJPEG(t))
	require.NoError(t, err)

	path2, err := store.Save(t.Context(), 1, createPNG(t))
	assert.NoError(t, err)

	// Different extension means different file
	assert.NotEqual(t, path1, path2)
}
