package repository

import (
	"context"
	"errors"

	"github.com/Velin-Todorov/zetta-task/internal/db"
)

// ErrConflict error that is thrown when you try to create a book with author and title
// that already exist.
var ErrConflict = errors.New("duplicate entry")

// BookFilter allows users to filter the results returned by GET /books
type BookFilter struct {
	Title *string
	Author *string
	PublishedAt *string
	PublishedAfter *string
	PublishedBefore *string
	Limit uint64
	Offset uint64
}

// BookRepository an interface for interacting with the db layer
type BookRepository interface {
	GetBooks(ctx context.Context, filter BookFilter) ([]db.Book, error)
	GetBook(ctx context.Context, id int64) (*db.Book, error)
	CreateBook(ctx context.Context, params db.CreateBookParams) (*db.Book, error)
	UpdateBook(ctx context.Context, params db.UpdateBookParams) (*db.Book, error)
	DeleteBook(ctx context.Context, id int64) error
	SetBookCover(ctx context.Context, params db.UpdateBookCoverParams) (*db.Book, error)
}

// IsConflict is a helper for checking if an error is ErrConflict
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}