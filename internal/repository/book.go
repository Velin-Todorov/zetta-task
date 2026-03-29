package repository

import (
	"context"
	"errors"

	"github.com/Velin-Todorov/zetta-task/internal/db"
)

var ErrConflict = errors.New("duplicate entry")

type BookFilter struct {
	Title *string
	Author *string
	PublishedAt *string
	PublishedAfter *string
	PublishedBefore *string
	Limit uint64
	Offset uint64
}

type BookRepository interface {
	GetBooks(ctx context.Context, filter BookFilter) ([]db.Book, error)
	GetBook(ctx context.Context, id int64) (*db.Book, error)
	CreateBook(ctx context.Context, params db.CreateBookParams) (*db.Book, error)
	UpdateBook(ctx context.Context, params db.UpdateBookParams) (*db.Book, error)
	DeleteBook(ctx context.Context, id int64) error
	SetBookCover(ctx context.Context, params db.UpdateBookCoverParams) (*db.Book, error)
}

func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}