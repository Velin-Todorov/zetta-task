package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Velin-Todorov/zetta-task/internal/db"
	"github.com/Velin-Todorov/zetta-task/internal/repository"
)

var testDB *sql.DB

// TestMain is used here to set up a test container
func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mysql:8",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "root",
				"MYSQL_DATABASE":      "bookstore_test",
			},
			WaitingFor: wait.ForLog("ready for connections").WithOccurrence(2),
		},
		Started: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to start container: %v", err))
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "3306")

	dsn := fmt.Sprintf("root:root@tcp(%s:%s)/bookstore_test?parseTime=true", host, port.Port())

	testDB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to connect: %v", err))
	}

	// Wait for MySQL to be ready
	for range 30 {
		if err := testDB.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err := testDB.PingContext(ctx); err != nil {
		panic(fmt.Sprintf("mysql not ready: %v", err))
	}

	// Run migrations
	createTable := `
		CREATE TABLE IF NOT EXISTS books (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			title VARCHAR(255) NOT NULL,
			author VARCHAR(255) NOT NULL,
			cover_path VARCHAR(255),
			published_at DATE NOT NULL
		)`
	if _, err := testDB.ExecContext(ctx, createTable); err != nil {
		panic(fmt.Sprintf("failed to create table: %v", err))
	}

	createIndex := `CREATE UNIQUE INDEX title_author ON books (title, author)`
	if _, err := testDB.ExecContext(ctx, createIndex); err != nil {
		panic(fmt.Sprintf("failed to create index: %v", err))
	}

	code := m.Run()

	testDB.Close()
	container.Terminate(ctx)
	os.Exit(code)
}

func cleanup(t *testing.T) {
	t.Helper()
	_, err := testDB.ExecContext(t.Context(), "DELETE FROM books")
	require.NoError(t, err)
}

func newRepo() repository.BookRepository {
	return repository.NewMySQLBookRepository(testDB)
}

func seedBook(t *testing.T, title, author string, publishedAt time.Time) *db.Book {
	t.Helper()
	repo := newRepo()
	book, err := repo.CreateBook(t.Context(), db.CreateBookParams{
		Title:       title,
		Author:      author,
		PublishedAt: publishedAt,
	})
	require.NoError(t, err)
	return book
}

func TestCreateBook_Success(t *testing.T) {
	cleanup(t)
	repo := newRepo()

	book, err := repo.CreateBook(t.Context(), db.CreateBookParams{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dune", book.Title)
	assert.Equal(t, "Frank Herbert", book.Author)
	assert.False(t, book.CoverPath.Valid)
}

func TestCreateBook_Conflict(t *testing.T) {
	cleanup(t)
	repo := newRepo()
	date := time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC)

	_, err := repo.CreateBook(t.Context(), db.CreateBookParams{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: date,
	})
	require.NoError(t, err)

	_, err = repo.CreateBook(t.Context(), db.CreateBookParams{
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: date,
	})

	assert.Error(t, err)
	assert.True(t, repository.IsConflict(err))
}

// --- GetBook ---

func TestGetBook_Success(t *testing.T) {
	cleanup(t)
	seeded := seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	book, err := repo.GetBook(t.Context(), seeded.ID)

	assert.NoError(t, err)
	assert.Equal(t, seeded.ID, book.ID)
	assert.Equal(t, "Dune", book.Title)
}

func TestGetBook_NotFound(t *testing.T) {
	cleanup(t)
	repo := newRepo()

	_, err := repo.GetBook(t.Context(), 99999)

	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestGetBooks_NoFilters(t *testing.T) {
	cleanup(t)
	seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	seedBook(t, "1984", "George Orwell", time.Date(1949, 6, 8, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	result, err := repo.GetBooks(t.Context(), repository.BookFilter{
		Limit:  100,
		Offset: 0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGetBooks_FilterByAuthor(t *testing.T) {
	cleanup(t)
	seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	seedBook(t, "1984", "George Orwell", time.Date(1949, 6, 8, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	author := "Frank Herbert"
	result, err := repo.GetBooks(t.Context(), repository.BookFilter{
		Author: &author,
		Limit:  100,
		Offset: 0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Dune", result[0].Title)
}

func TestGetBooks_FilterByTitle(t *testing.T) {
	cleanup(t)
	seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	seedBook(t, "1984", "George Orwell", time.Date(1949, 6, 8, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	title := "1984"
	result, err := repo.GetBooks(t.Context(), repository.BookFilter{
		Title:  &title,
		Limit:  100,
		Offset: 0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "George Orwell", result[0].Author)
}

func TestGetBooks_FilterByDateRange(t *testing.T) {
	cleanup(t)
	seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	seedBook(t, "1984", "George Orwell", time.Date(1949, 6, 8, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	after := "1960-01-01"
	before := "1970-01-01"
	result, err := repo.GetBooks(t.Context(), repository.BookFilter{
		PublishedAfter:  &after,
		PublishedBefore: &before,
		Limit:           100,
		Offset:          0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Dune", result[0].Title)
}

func TestGetBooks_Pagination(t *testing.T) {
	cleanup(t)
	seedBook(t, "Book A", "Author A", time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	seedBook(t, "Book B", "Author B", time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC))
	seedBook(t, "Book C", "Author C", time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	result, err := repo.GetBooks(t.Context(), repository.BookFilter{
		Limit:  2,
		Offset: 0,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	result, err = repo.GetBooks(t.Context(), repository.BookFilter{
		Limit:  2,
		Offset: 2,
	})

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestUpdateBook_Success(t *testing.T) {
	cleanup(t)
	seeded := seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	book, err := repo.UpdateBook(t.Context(), db.UpdateBookParams{
		ID:          seeded.ID,
		Title:       "Dune Messiah",
		Author:      "Frank Herbert",
		PublishedAt: time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dune Messiah", book.Title)
	assert.Equal(t, time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC), book.PublishedAt)
}

func TestUpdateBook_Conflict(t *testing.T) {
	cleanup(t)
	seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	other := seedBook(t, "Dune Messiah", "Frank Herbert", time.Date(1969, 10, 1, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	_, err := repo.UpdateBook(t.Context(), db.UpdateBookParams{
		ID:          other.ID,
		Title:       "Dune",
		Author:      "Frank Herbert",
		PublishedAt: other.PublishedAt,
	})

	assert.Error(t, err)
	assert.True(t, repository.IsConflict(err))
}

func TestSetBookCover_Success(t *testing.T) {
	cleanup(t)
	seeded := seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	book, err := repo.SetBookCover(t.Context(), db.UpdateBookCoverParams{
		ID:        seeded.ID,
		CoverPath: sql.NullString{String: "uploads/covers/1.jpg", Valid: true},
	})

	assert.NoError(t, err)
	assert.True(t, book.CoverPath.Valid)
	assert.Equal(t, "uploads/covers/1.jpg", book.CoverPath.String)
}

func TestDeleteBook_Success(t *testing.T) {
	cleanup(t)
	seeded := seedBook(t, "Dune", "Frank Herbert", time.Date(1965, 8, 1, 0, 0, 0, 0, time.UTC))
	repo := newRepo()

	err := repo.DeleteBook(t.Context(), seeded.ID)
	assert.NoError(t, err)

	_, err = repo.GetBook(t.Context(), seeded.ID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}
