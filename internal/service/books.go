package bookstore

import (
	"context"
	"io"
	"time"

	"goa.design/clue/log"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	"github.com/Velin-Todorov/zetta-task/internal/db"
	"github.com/Velin-Todorov/zetta-task/internal/repository"
)

// books service example implementation.
// The example methods log the requests and return zero values.
type bookssrvc struct{
	repo repository.BookRepository
}

// NewBooks returns the books service implementation.
func NewBooks(repo repository.BookRepository) books.Service {
	return &bookssrvc{repo: repo}
}

// GetBooks implements getBooks.
func (s *bookssrvc) GetBooks(ctx context.Context, p *books.GetBooksPayload) (res []*books.Book, err error) {
	log.Printf(ctx, "books.getBooks")
	return
}

// GetBook implements getBook.
func (s *bookssrvc) GetBook(ctx context.Context, p *books.GetBookPayload) (res *books.Book, err error) {
	dbBook, err := s.repo.GetBook(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	return toBook(dbBook), nil
}

// CreateBook implements createBook.
func (s *bookssrvc) CreateBook(ctx context.Context, p *books.CreateBookPayload) (res *books.Book, err error) {
	publishedAt, err := time.Parse(time.DateOnly, p.PublishedAt)
	if err != nil {
        return nil, err
    }

	book, err := s.repo.CreateBook(ctx, db.CreateBookParams{
		Title: p.Title,
		Author: p.Author,
		PublishedAt: publishedAt,
	})
	if err != nil {
		return nil, err
	}

	return toBook(book), nil
}

// CreateBookCover implements createBookCover.
func (s *bookssrvc) CreateBookCover(ctx context.Context, p *books.CreateBookCoverPayload, req io.ReadCloser) (res *books.Book, err error) {
	// req is the HTTP request body stream.
	defer req.Close()
	res = &books.Book{}
	log.Printf(ctx, "books.createBookCover")
	return
}

// UpdateBook implements updateBook.
func (s *bookssrvc) UpdateBook(ctx context.Context, p *books.UpdateBookPayload) (res *books.Book, err error) {
	res = &books.Book{}
	log.Printf(ctx, "books.updateBook")
	return
}

// UpdateBookCover implements updateBookCover.
func (s *bookssrvc) UpdateBookCover(ctx context.Context, p *books.UpdateBookCoverPayload, req io.ReadCloser) (res *books.Book, err error) {
	// req is the HTTP request body stream.
	defer req.Close()
	res = &books.Book{}
	log.Printf(ctx, "books.updateBookCover")
	return
}

// DeleteBook implements deleteBook.
func (s *bookssrvc) DeleteBook(ctx context.Context, p *books.DeleteBookPayload) (err error) {
	log.Printf(ctx, "books.deleteBook")
	return
}

func toBook(dbBook *db.Book) *books.Book {
	var coverURL *string
	if dbBook.CoverPath.Valid {
		coverURL = &dbBook.CoverPath.String
	}
	return &books.Book{
		ID: dbBook.ID,
		Title: dbBook.Title,
		Author: dbBook.Author,
		CoverURL: coverURL,
		PublishedAt: dbBook.PublishedAt.Format(time.DateOnly),
	}
}