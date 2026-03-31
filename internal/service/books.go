package bookstore

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
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
		return nil, books.MakeInternalError(err)
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
		return nil, books.MakeNotFound(err)
	}
	if err != nil {
		return nil, books.MakeInternalError(err)
	}

	return toBook(dbBook), nil
}

// CreateBook implements createBook.
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

// SetBookCover implements setBookCover.
func (s *bookssrvc) SetBookCover(ctx context.Context, p *books.SetBookCoverPayload, req io.ReadCloser) (*books.Book, error) {
	defer req.Close()

	_, err := s.repo.GetBook(ctx, p.ID)
	if err == sql.ErrNoRows {
		return nil, books.MakeNotFound(err)
	}
	if err != nil {
		return nil, books.MakeInternalError(err)
	}

	file, err := extractFile(req, p.ContentType)
	if err != nil {
		return nil, books.MakeInvalidInput(err)
	}
	defer file.Close()

	path, err := s.storage.Save(ctx, p.ID, file)
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

// UpdateBook implements updateBook.
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

// DeleteBook implements deleteBook.
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

func extractFile(body io.Reader, contentType string) (io.ReadCloser, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, fmt.Errorf("no boundary in content type")
	}
	reader := multipart.NewReader(body, boundary)
	part, err := reader.NextPart()
	if err != nil {
		return nil, err
	}
	return part, nil
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
