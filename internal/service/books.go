package bookstore

import (
	"context"
	"database/sql"
	"io"
	"time"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	"github.com/Velin-Todorov/zetta-task/internal/db"
	"github.com/Velin-Todorov/zetta-task/internal/repository"
	"github.com/Velin-Todorov/zetta-task/internal/storage"
)

type bookssrvc struct {
	repo    repository.BookRepository
	storage storage.ImageStore
}

// NewBooks returns the books service implementation.
func NewBooks(repo repository.BookRepository, storage storage.ImageStore) books.Service {
	return &bookssrvc{repo: repo, storage: storage}
}

// GetBooks implements getBooks.
func (s *bookssrvc) GetBooks(ctx context.Context, p *books.GetBooksPayload) ([]*books.Book, error) {
	var result []*books.Book
	limit := uint64(20)
	offset := uint64(0)
	
	if p.Offset != nil {
		offset = *p.Offset
	}

	if p.Limit != nil {
		limit = *p.Limit
	}
	
	filter := repository.BookFilter{
		Title:           p.Title,
		Author:          p.Author,
		PublishedAt:     p.PublishedAt,
		PublishedAfter:  p.PublishedAfter,
		PublishedBefore: p.PublishedBefore,
		Limit:           limit,
		Offset:          offset,
	}


	res, err := s.repo.GetBooks(ctx, filter)
	if err != nil {
		return nil, books.InternalError(err.Error())
	}

	for _, r := range res {
		result = append(result, toBook(&r))
	}

	return result, nil
}

// GetBook implements getBook.
func (s *bookssrvc) GetBook(ctx context.Context, p *books.GetBookPayload) (*books.Book, error) {
	dbBook, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return nil, books.NotFound("book not found")
	}
	if err != nil {
		return nil, books.InternalError(err.Error())
	}

	return toBook(dbBook), nil
}

// CreateBook implements createBook.
func (s *bookssrvc) CreateBook(ctx context.Context, p *books.CreateBookPayload) (*books.Book, error) {
	publishedAt, err := time.Parse(time.DateOnly, p.PublishedAt)
	if err != nil {
		return nil, books.InvalidInput("invalid date format")
	}

	book, err := s.repo.CreateBook(ctx, db.CreateBookParams{
		Title:       p.Title,
		Author:      p.Author,
		PublishedAt: publishedAt,
	})
	if err != nil {
		if repository.IsConflict(err) {
			return nil, books.Conflict("book with this title and author already exists")
		}
		return nil, books.InternalError(err.Error())
	}

	return toBook(book), nil
}

// SetBookCover implements setBookCover.
func (s *bookssrvc) SetBookCover(ctx context.Context, p *books.SetBookCoverPayload, req io.ReadCloser) (*books.Book, error) {
	defer req.Close()

	path, err := s.storage.Save(ctx, p.ID, req)
	if err != nil {
		if storage.IsInvalidFormat(err) {
			return nil, books.InvalidImageFormat(err.Error())
		}
		if storage.IsTooLarge(err) {
			return nil, books.PayloadTooLarge(err.Error())
		}
		return nil, books.InternalError(err.Error())
	}

	book, err := s.repo.SetBookCover(ctx, db.UpdateBookCoverParams{
		ID:        p.ID,
		CoverPath: sql.NullString{String: path, Valid: true},
	})
	if err == sql.ErrNoRows {
		return nil, books.NotFound("book not found")
	}
	if err != nil {
		return nil, books.InternalError(err.Error())
	}

	return toBook(book), nil
}

// UpdateBook implements updateBook.
func (s *bookssrvc) UpdateBook(ctx context.Context, p *books.UpdateBookPayload) (*books.Book, error) {
	existing, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return nil, books.NotFound("book not found")
	}
	if err != nil {
		return nil, books.InternalError(err.Error())
	}

	params := db.UpdateBookParams{
		ID:          existing.ID,
		Title:       existing.Title,
		Author:      existing.Author,
		PublishedAt: existing.PublishedAt,
	}

	if p.Title != nil {
		params.Title = *p.Title
	}
	if p.Author != nil {
		params.Author = *p.Author
	}
	if p.PublishedAt != nil {
		params.PublishedAt, err = time.Parse(time.DateOnly, *p.PublishedAt)
		if err != nil {
			return nil, books.InvalidInput("invalid date format")
		}
	}
	book, err := s.repo.UpdateBook(ctx, params)
	if err != nil {
		if repository.IsConflict(err) {
			return nil, books.Conflict("book with this title and author already exists")
		}
		return nil, books.InternalError(err.Error())
	}

	return toBook(book), nil
}

// DeleteBook implements deleteBook.
func (s *bookssrvc) DeleteBook(ctx context.Context, p *books.DeleteBookPayload) error {
	_, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return books.NotFound("book not found")
	}
	if err != nil {
		return books.InternalError(err.Error())
	}

	if err := s.repo.DeleteBook(ctx, p.ID); err != nil {
		return books.InternalError(err.Error())
	}
	return nil
}

func toBook(dbBook *db.Book) *books.Book {
	var coverURL *string
	if dbBook.CoverPath.Valid {
		coverURL = &dbBook.CoverPath.String
	}
	return &books.Book{
		ID:          dbBook.ID,
		Title:       dbBook.Title,
		Author:      dbBook.Author,
		CoverURL:    coverURL,
		PublishedAt: dbBook.PublishedAt.Format(time.DateOnly),
	}
}
