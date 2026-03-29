# Bookstore API

A RESTful API for managing books, built with Go using the Goa framework. Supports CRUD operations, book cover image uploads, filtering, and pagination.

## Prerequisites

- Go 1.25+
- Docker & Docker Compose

## Setup & Run

1. Start the MySQL database:

```bash
docker compose up -d
```

2. Run the application:

```bash
go run ./cmd/bookstore --http-port 8080
```

The server starts on `http://localhost:8080`. Database migrations run automatically on startup.

## Configuration

Application settings are in `config.yaml`:

```yaml
database:
  host: localhost
  port: 3306
  user: root
  password: secret
  name: bookstore

server:
  port: 8080

storage:
  upload_path: uploads/covers
```

## API Endpoints

### List Books

```
GET /books
```

Query parameters (all optional):
- `title` - filter by title
- `author` - filter by author
- `publishedAt` - filter by exact date (YYYY-MM-DD)
- `published_after` - filter by date range start (YYYY-MM-DD)
- `published_before` - filter by date range end (YYYY-MM-DD)
- `limit` - pagination limit (default: 20)
- `offset` - pagination offset (default: 0)

Example:
```bash
curl "http://localhost:8080/books?author=Frank+Herbert&limit=10"
```

### Get Book

```
GET /books/{id}
```

Example:
```bash
curl http://localhost:8080/books/1
```

### Create Book

```
POST /books
Content-Type: application/json
```

Body:
```json
{
  "title": "Dune",
  "author": "Frank Herbert",
  "publishedAt": "1965-08-01"
}
```

Example:
```bash
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"Dune","author":"Frank Herbert","publishedAt":"1965-08-01"}'
```

### Update Book (Partial)

```
PATCH /books/{id}
Content-Type: application/json
```

Only include the fields you want to update:
```json
{
  "title": "Dune Messiah"
}
```

Example:
```bash
curl -X PATCH http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Dune Messiah"}'
```

### Upload Book Cover

```
PUT /books/{id}/cover
Content-Type: multipart/form-data
```

Supported formats: JPEG, PNG, WebP. Max size: 5MB.

Example:
```bash
curl -X PUT http://localhost:8080/books/1/cover \
  -F "cover=@/path/to/image.jpg"
```

### Delete Book

```
DELETE /books/{id}
```

Example:
```bash
curl -X DELETE http://localhost:8080/books/1
```

## Response Codes

| Code | Description |
|------|-------------|
| 200  | Success |
| 201  | Created |
| 204  | No Content (delete) |
| 400  | Bad Request (invalid input / unsupported image format) |
| 404  | Not Found |
| 409  | Conflict (duplicate title + author) |
| 413  | Payload Too Large (image exceeds 5MB) |
| 500  | Internal Server Error |

## Project Structure

```
cmd/bookstore/       - Application entry point and HTTP server
design/              - Goa API design DSL
gen/                 - Goa generated code (do not edit)
internal/
  config/            - YAML configuration loader
  db/                - sqlc generated database code
  repository/        - Database access layer (repository pattern)
  service/           - Business logic (implements Goa service interface)
  storage/           - File storage for book covers
sql/
  schema/            - Database migrations
  queries/           - sqlc query definitions
```

## Running Tests

```bash
# All tests
go test ./...

# Service layer tests (unit)
go test ./internal/service/ -v

# Storage tests
go test ./internal/storage/ -v

# Repository integration tests (requires Docker)
go test ./internal/repository/ -v -timeout 120s # You need the timeout to give enough for the testcontainer to start
```

## API Specification

The full OpenAPI 3.0 specification is available at `gen/http/openapi3.yaml`.
