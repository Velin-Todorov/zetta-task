package bookstore

import (
	"bytes"
	"context"
	"database/sql"
	"time"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	"github.com/Velin-Todorov/zetta-task/internal/db"
	"github.com/Velin-Todorov/zetta-task/internal/repository"
	"github.com/Velin-Todorov/zetta-task/internal/storage"
)

// bookssrvc implements the books service interface.
type bookssrvc struct {
	repo    repository.BookRepository
	storage storage.ImageStore
}

// NewBooks returns the books service implementation.
func NewBooks(repo repository.BookRepository, storage storage.ImageStore) books.Service {
	return &bookssrvc{repo: repo, storage: storage}
}

// GetBooks returns a list of books matching the given filters with pagination.
// Defaults to limit=20, offset=0 if not provided.
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
		return nil, books.MakeInternalError(err)
	}

	for _, r := range res {
		result = append(result, toBook(&r))
	}

	return result, nil
}

// GetBook returns a single book by ID.
func (s *bookssrvc) GetBook(ctx context.Context, p *books.GetBookPayload) (*books.Book, error) {
	dbBook, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return nil, books.MakeNotFound(err)
	}
	if err != nil {
		return nil, books.MakeInternalError(err)
	}

	return toBook(dbBook), nil
}

// CreateBook creates a new book. Returns conflict if the title+author combination already exists.
func (s *bookssrvc) CreateBook(ctx context.Context, p *books.CreateBookPayload) (*books.Book, error) {
	publishedAt, err := time.Parse(time.DateOnly, p.PublishedAt)
	if err != nil {
		return nil, books.MakeInvalidInput(err)
	}

	book, err := s.repo.CreateBook(ctx, db.CreateBookParams{
		Title:       p.Title,
		Author:      p.Author,
		PublishedAt: publishedAt,
	})
	if err != nil {
		if repository.IsConflict(err) {
			return nil, books.MakeConflict(err)
		}
		return nil, books.MakeInternalError(err)
	}

	return toBook(book), nil
}

// SetBookCover uploads a cover image for a book. Validates the book exists before saving the file.
func (s *bookssrvc) SetBookCover(ctx context.Context, p *books.SetBookCoverPayload) (*books.Book, error) {
	_, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return nil, books.MakeNotFound(err)
	}
	if err != nil {
		return nil, books.MakeInternalError(err)
	}

	path, err := s.storage.Save(ctx, p.ID, bytes.NewReader(p.Cover))
	if err != nil {
		if storage.IsInvalidFormat(err) {
			return nil, books.MakeInvalidImageFormat(err)
		}
		if storage.IsTooLarge(err) {
			return nil, books.MakePayloadTooLarge(err)
		}
		return nil, books.MakeInternalError(err)
	}

	book, err := s.repo.SetBookCover(ctx, db.UpdateBookCoverParams{
		ID:        p.ID,
		CoverPath: sql.NullString{String: path, Valid: true},
	})
	if err == sql.ErrNoRows {
		return nil, books.MakeNotFound(err)
	}
	if err != nil {
		return nil, books.MakeInternalError(err)
	}

	return toBook(book), nil
}

// UpdateBook partially updates a book. Fetches the existing book first and merges only the provided fields.
func (s *bookssrvc) UpdateBook(ctx context.Context, p *books.UpdateBookPayload) (*books.Book, error) {
	existing, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return nil, books.MakeNotFound(err)
	}
	if err != nil {
		return nil, books.MakeInternalError(err)
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
			return nil, books.MakeInvalidInput(err)
		}
	}
	book, err := s.repo.UpdateBook(ctx, params)
	if err != nil {
		if repository.IsConflict(err) {
			return nil, books.MakeConflict(err)
		}
		return nil, books.MakeInternalError(err)
	}

	return toBook(book), nil
}

// DeleteBook removes a book by ID. Returns not found if the book doesn't exist.
func (s *bookssrvc) DeleteBook(ctx context.Context, p *books.DeleteBookPayload) error {
	_, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return books.MakeNotFound(err)
	}
	if err != nil {
		return books.MakeInternalError(err)
	}

	if err := s.repo.DeleteBook(ctx, p.ID); err != nil {
		return books.MakeInternalError(err)
	}
	return nil
}

// toBook converts a db.Book to the API response type.
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
