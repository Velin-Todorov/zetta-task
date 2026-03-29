// Whitebox test
package bookstore

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	"github.com/Velin-Todorov/zetta-task/internal/db"
)

func TestToBook_WithCover(t *testing.T) {
	dbBook := &db.Book{
		ID:          1,
		Title:       "Dune",
		Author:      "Frank Herbert",
		CoverPath:   sql.NullString{String: "uploads/covers/1.jpg", Valid: true},
		PublishedAt: time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC),
	}

	result := toBook(dbBook)

	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Dune", result.Title)
	assert.Equal(t, "Frank Herbert", result.Author)
	assert.NotNil(t, result.CoverURL)
	assert.Equal(t, "uploads/covers/1.jpg", *result.CoverURL)
	assert.Equal(t, "1965-08-01", result.PublishedAt)
}

func TestToBook_WithoutCover(t *testing.T) {
	dbBook := &db.Book{
		ID:          2,
		Title:       "1984",
		Author:      "George Orwell",
		CoverPath:   sql.NullString{Valid: false},
		PublishedAt: time.Date(1949, 6, 8, 0, 0, 0, 0, time.UTC),
	}

	result := toBook(dbBook)

	assert.Nil(t, result.CoverURL)
	assert.Equal(t, "1949-06-08", result.PublishedAt)
}

func TestToBook_DateFormatting(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "standard date",
			date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: "2024-01-15",
		},
		{
			name:     "end of year",
			date:     time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			expected: "2023-12-31",
		},
		{
			name:     "leap day",
			date:     time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
			expected: "2024-02-29",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbBook := &db.Book{
				ID:          1,
				Title:       "Test",
				Author:      "Author",
				PublishedAt: tt.date,
			}
			result := toBook(dbBook)
			assert.Equal(t, tt.expected, result.PublishedAt)
		})
	}
}

func TestToBook_ReturnsCorrectType(t *testing.T) {
	dbBook := &db.Book{
		ID:          1,
		Title:       "Test",
		Author:      "Author",
		PublishedAt: time.Now(),
	}

	result := toBook(dbBook)

	assert.IsType(t, &books.Book{}, result)
}
