package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"

	"github.com/Velin-Todorov/zetta-task/internal/db"
)

type mysqlBookRepository struct {
	queries *db.Queries
	db      *sql.DB
}

func (r *mysqlBookRepository) GetBooks(ctx context.Context, filter BookFilter) ([]db.Book, error) {
	query := sq.Select("*").
		From("books").
		Limit(filter.Limit).
		Offset(filter.Offset)
	
	if filter.Author != nil {
		query = query.Where(sq.Eq{"author": *filter.Author})
	}

	if filter.Title != nil {
		query = query.Where(sq.Eq{"title": *filter.Title})
	}

	if filter.PublishedAt != nil {
		query = query.Where(sq.Eq{"published_at": *filter.PublishedAt})
	}

	if filter.PublishedAfter != nil {
		query = query.Where(sq.GtOrEq{"published_at": *filter.PublishedAfter})
	}

	if filter.PublishedBefore != nil {
		query = query.Where(sq.LtOrEq{"published_at": *filter.PublishedBefore})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []db.Book

	for rows.Next() {
		var b db.Book
        if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.CoverPath, &b.PublishedAt); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	return books, rows.Err()
}

func (r *mysqlBookRepository) GetBook(ctx context.Context, id int64) (*db.Book, error) {
	book, err := r.queries.GetBook(ctx, id)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *mysqlBookRepository) CreateBook(ctx context.Context, params db.CreateBookParams) (*db.Book, error) {
	res, err := r.queries.CreateBook(ctx, params)
	if err != nil {
		return nil, err
	}

	insertedId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	book, err := r.queries.GetBook(ctx, insertedId)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *mysqlBookRepository) UpdateBook(ctx context.Context, params db.UpdateBookParams) (*db.Book, error) {
	_, err := r.queries.UpdateBook(ctx, params)
	if err != nil {
		return nil, err
	}

	book, err := r.queries.GetBook(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func (r *mysqlBookRepository) DeleteBook(ctx context.Context, id int64) error {
	err := r.queries.DeleteBook(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *mysqlBookRepository) SetBookCover(ctx context.Context, params db.UpdateBookCoverParams) (*db.Book, error) {
	_, err := r.queries.UpdateBookCover(ctx, params)
	if err != nil {
		return nil, err
	}

	book, err := r.queries.GetBook(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func NewMySQLBookRepository(database *sql.DB) BookRepository {
	return &mysqlBookRepository{
		queries: db.New(database),
		db:      database,
	}
}
