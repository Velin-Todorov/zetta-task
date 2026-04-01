// Blackbox tests
package bookstore_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	goa "goa.design/goa/v3/pkg"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	"github.com/Velin-Todorov/zetta-task/internal/db"
	"github.com/Velin-Todorov/zetta-task/internal/repository"
	repoMocks "github.com/Velin-Todorov/zetta-task/internal/repository/mocks"
	bookstore "github.com/Velin-Todorov/zetta-task/internal/service"
	"github.com/Velin-Todorov/zetta-task/internal/storage"
	storageMocks "github.com/Velin-Todorov/zetta-task/internal/storage/mocks"
)

func assertServiceError(t *testing.T, err error, name string) {
	t.Helper()
	var svcErr *goa.ServiceError
	if assert.ErrorAs(t, err, &svcErr) {
		assert.Equal(t, name, svcErr.Name)
	}
}

func DBMock() *db.Book {
	return &db.Book{
		ID:          1,
		Title:       "Dune",
		Author:      "Frank Herbert",
		CoverPath:   sql.NullString{String: "uploads/covers/1.jpg", Valid: true},
		PublishedAt: time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC),
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestGetBooks_Success(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.Anything).Return([]db.Book{*DBMock()}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Dune", result[0].Title)
	assert.Equal(t, "Frank Herbert", result[0].Author)
}

func TestGetBooks_EmptyResult(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.Anything).Return([]db.Book{}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{})

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetBooks_RepoError(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{})

	assert.Error(t, err)
	assertServiceError(t, err, "internal_error")
}

func TestGetBooks_DefaultLimitAndOffset(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return f.Limit == 20 && f.Offset == 0
	})).Return([]db.Book{}, nil)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{})

	assert.NoError(t, err)
}

func TestGetBooks_CustomLimitAndOffset(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return f.Limit == 10 && f.Offset == 5
	})).Return([]db.Book{}, nil)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		Limit:  ptr(uint64(10)),
		Offset: ptr(uint64(5)),
	})

	assert.NoError(t, err)
}

func TestGetBooks_PassesFilters(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return *f.Title == "Dune" && *f.Author == "Frank Herbert"
	})).Return([]db.Book{}, nil)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		Title:  ptr("Dune"),
		Author: ptr("Frank Herbert"),
	})

	assert.NoError(t, err)
}

func TestGetBooks_FilterByPublishedAt(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return f.PublishedAt != nil && *f.PublishedAt == "1965-08-01"
	})).Return([]db.Book{*DBMock()}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		PublishedAt: ptr("1965-08-01"),
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetBooks_FilterByPublishedAfter(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return f.PublishedAfter != nil && *f.PublishedAfter == "1960-01-01"
	})).Return([]db.Book{*DBMock()}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		PublishedAfter: ptr("1960-01-01"),
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetBooks_FilterByPublishedBefore(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return f.PublishedBefore != nil && *f.PublishedBefore == "1970-01-01"
	})).Return([]db.Book{*DBMock()}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		PublishedBefore: ptr("1970-01-01"),
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetBooks_FilterByDateRange(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return f.PublishedAfter != nil && *f.PublishedAfter == "1960-01-01" &&
			f.PublishedBefore != nil && *f.PublishedBefore == "1970-01-01"
	})).Return([]db.Book{*DBMock()}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		PublishedAfter:  ptr("1960-01-01"),
		PublishedBefore: ptr("1970-01-01"),
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetBooks_AllFilters(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBooks(mock.Anything, mock.MatchedBy(func(f repository.BookFilter) bool {
		return *f.Title == "Dune" &&
			*f.Author == "Frank Herbert" &&
			*f.PublishedAfter == "1960-01-01" &&
			*f.PublishedBefore == "1970-01-01" &&
			f.Limit == 10 &&
			f.Offset == 5
	})).Return([]db.Book{*DBMock()}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBooks(t.Context(), &books.GetBooksPayload{
		Title:          ptr("Dune"),
		Author:         ptr("Frank Herbert"),
		PublishedAfter: ptr("1960-01-01"),
		PublishedBefore: ptr("1970-01-01"),
		Limit:          ptr(uint64(10)),
		Offset:         ptr(uint64(5)),
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetBook_Success(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.GetBook(t.Context(), &books.GetBookPayload{ID: 1})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "Dune", result.Title)
	assert.Equal(t, "1965-08-01", result.PublishedAt)
}

func TestGetBook_NotFound(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.GetBook(t.Context(), &books.GetBookPayload{ID: 999})

	assert.Error(t, err)
	assertServiceError(t, err, "not_found")
}

func TestGetBook_InternalError(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(nil, errors.New("connection refused"))

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.GetBook(t.Context(), &books.GetBookPayload{ID: 1})

	assert.Error(t, err)
	assertServiceError(t, err, "internal_error")
}

// --- CreateBook ---

func TestCreateBook_Success(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().CreateBook(mock.Anything, mock.MatchedBy(func(p db.CreateBookParams) bool {
		return p.Title == "Dune" && p.Author == "Frank Herbert"
	})).Return(DBMock(), nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.CreateBook(t.Context(), &books.CreateBookPayload{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: "1965-08-01",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dune", result.Title)
	assert.Equal(t, "1965-08-01", result.PublishedAt)
}

func TestCreateBook_InvalidDate(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.CreateBook(t.Context(), &books.CreateBookPayload{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: "not-a-date",
	})

	assert.Error(t, err)
	assertServiceError(t, err, "invalid_input")
}

func TestCreateBook_Conflict(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().CreateBook(mock.Anything, mock.Anything).Return(nil, repository.ErrConflict)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.CreateBook(t.Context(), &books.CreateBookPayload{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: "1965-08-01",
	})

	assert.Error(t, err)
	assertServiceError(t, err, "conflict")
}

func TestCreateBook_InternalError(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().CreateBook(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.CreateBook(t.Context(), &books.CreateBookPayload{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: "1965-08-01",
	})

	assert.Error(t, err)
	assertServiceError(t, err, "internal_error")
}

func TestUpdateBook_Success(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)
	existing := DBMock()

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(existing, nil)
	repo.EXPECT().UpdateBook(mock.Anything, mock.MatchedBy(func(p db.UpdateBookParams) bool {
		return p.Title == "Dune Messiah" && p.Author == "Frank Herbert"
	})).Return(&db.Book{
		ID:          1,
		Title:       "Dune Messiah",
		Author:      "Frank Herbert",
		CoverPath:   existing.CoverPath,
		PublishedAt: existing.PublishedAt,
	}, nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.UpdateBook(t.Context(), &books.UpdateBookPayload{
		ID:    1,
		Title: ptr("Dune Messiah"),
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dune Messiah", result.Title)
	assert.Equal(t, "Frank Herbert", result.Author)
}

func TestUpdateBook_NotFound(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.UpdateBook(t.Context(), &books.UpdateBookPayload{
		ID:    999,
		Title: ptr("Dune Messiah"),
	})

	assert.Error(t, err)
	assertServiceError(t, err, "not_found")
}

func TestUpdateBook_InvalidDate(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.UpdateBook(t.Context(), &books.UpdateBookPayload{
		ID:          1,
		PublishedAt: ptr("not-a-date"),
	})

	assert.Error(t, err)
	assertServiceError(t, err, "invalid_input")
}

func TestUpdateBook_Conflict(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)
	repo.EXPECT().UpdateBook(mock.Anything, mock.Anything).Return(nil, repository.ErrConflict)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.UpdateBook(t.Context(), &books.UpdateBookPayload{
		ID:    1,
		Title: ptr("Existing Title"),
	})

	assert.Error(t, err)
	assertServiceError(t, err, "conflict")
}

func TestSetBookCover_Success(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)
	store.EXPECT().Save(mock.Anything, int64(1), mock.Anything).Return("uploads/covers/1.jpg", nil)
	repo.EXPECT().SetBookCover(mock.Anything, mock.Anything).Return(DBMock(), nil)

	svc := bookstore.NewBooks(repo, store)
	result, err := svc.SetBookCover(t.Context(), &books.SetBookCoverPayload{ID: 1, Cover: []byte("fake image data")})

	assert.NoError(t, err)
	assert.NotNil(t, result.CoverURL)
	assert.Equal(t, "uploads/covers/1.jpg", *result.CoverURL)
}

func TestSetBookCover_InvalidFormat(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)
	store.EXPECT().Save(mock.Anything, int64(1), mock.Anything).Return("", storage.ErrInvalidFormat)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.SetBookCover(t.Context(), &books.SetBookCoverPayload{ID: 1, Cover: []byte("not an image")})

	assert.Error(t, err)
	assertServiceError(t, err, "invalid_image_format")
}

func TestSetBookCover_TooLarge(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)
	store.EXPECT().Save(mock.Anything, int64(1), mock.Anything).Return("", storage.ErrTooLarge)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.SetBookCover(t.Context(), &books.SetBookCoverPayload{ID: 1, Cover: []byte("huge file")})

	assert.Error(t, err)
	assertServiceError(t, err, "payload_too_large")
}

func TestSetBookCover_BookNotFound(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	svc := bookstore.NewBooks(repo, store)
	_, err := svc.SetBookCover(t.Context(), &books.SetBookCoverPayload{ID: 999, Cover: []byte("fake image")})

	assert.Error(t, err)
	assertServiceError(t, err, "not_found")
}

func TestDeleteBook_Success(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)
	repo.EXPECT().DeleteBook(mock.Anything, int64(1)).Return(nil)
	store.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)

	svc := bookstore.NewBooks(repo, store)
	err := svc.DeleteBook(t.Context(), &books.DeleteBookPayload{ID: 1})

	assert.NoError(t, err)
}

func TestDeleteBook_NotFound(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	svc := bookstore.NewBooks(repo, store)
	err := svc.DeleteBook(t.Context(), &books.DeleteBookPayload{ID: 999})

	assert.Error(t, err)
	assertServiceError(t, err, "not_found")
}

func TestDeleteBook_InternalError(t *testing.T) {
	repo := repoMocks.NewBookRepository(t)
	store := storageMocks.NewImageStore(t)

	repo.EXPECT().GetBook(mock.Anything, int64(1)).Return(DBMock(), nil)
	repo.EXPECT().DeleteBook(mock.Anything, int64(1)).Return(errors.New("db error"))

	svc := bookstore.NewBooks(repo, store)
	err := svc.DeleteBook(t.Context(), &books.DeleteBookPayload{ID: 1})

	assert.Error(t, err)
	assertServiceError(t, err, "internal_error")
}
