# Bookstore API

A RESTful API for managing books, built with Go using the Goa framework. Supports CRUD operations, book cover image uploads, filtering, and pagination.

## Prerequisites

- Go 1.25+
- Docker & Docker Compose

## Setup & Run

1. Start everything with Docker Compose:

```bash
docker compose up --build
```

This starts both the MySQL database and the application. Database migrations run automatically on startup.

The server starts on `http://localhost:8080`.

2. Alternatively, run locally (requires a running MySQL instance):

```bash
docker compose up mysql -d
go run ./cmd/bookstore --http-port 8080
```

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

The `DB_HOST` environment variable overrides the database host, which is used by Docker Compose to connect to the MySQL container.

## API Documentation (Swagger)

Interactive Swagger UI is available at:

```
http://localhost:8080/swagger
```

The raw OpenAPI 3.0 spec is available at:

```
http://localhost:8080/openapi.json
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

Book covers are stored on the local filesystem under `uploads/covers/`. The file is validated by reading the file header (`http.DetectContentType`), not by trusting the file extension. Uploading a new cover replaces any existing one for that book.

The cover URL is returned in the book response as `cover_url` and can be accessed at:
```
GET /uploads/covers/{filename}
```

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

## Architecture

### Project Structure

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
  schema/            - Database migrations (golang-migrate)
  queries/           - sqlc query definitions
public/              - Static files (Swagger UI)
```

### Repository Pattern

The data access layer uses the repository pattern. `BookRepository` is an interface defined in `internal/repository/book.go`, with the MySQL implementation in `internal/repository/mysql.go`. This provides:

- **Testability** - the service layer depends on an interface, not a concrete implementation. Tests inject a mock repository without needing a database.
- **Separation of concerns** - SQL and database-specific logic stays in the repository. The service layer only knows about the interface.
- **Swappability** - the MySQL implementation can be replaced (e.g. PostgreSQL) without touching the service layer.

### Database

- **sqlc** generates type-safe Go code from SQL queries for static operations (get, create, update, delete).
- **squirrel** is used for the dynamic `GetBooks` query where filters are optional and built at runtime.
- **golang-migrate** runs schema migrations automatically on application startup.

## Running Tests

The project includes three types of tests:

### Whitebox Tests
Tests in the same package (`package bookstore`) that verify internal/unexported functions like `toBook` mapping logic.

### Blackbox Tests
Tests in an external package (`package bookstore_test`) that verify the public service API. Dependencies (repository, storage) are mocked using [mockery](https://github.com/vektra/mockery)-generated mocks.

### Integration Tests (Testcontainers)
Repository tests use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to spin up a real MySQL container for each test run. This ensures the SQL queries, migrations, and constraint handling (e.g. unique index conflicts) are tested against an actual database, not mocks. The container is created and destroyed automatically - no manual setup required, just Docker running.

```bash
# All tests
go test ./...

# Service layer tests (unit)
go test ./internal/service/ -v

# Storage tests
go test ./internal/storage/ -v

# Repository integration tests (requires Docker)
go test ./internal/repository/ -v -timeout 120s
```

The timeout flag for repository tests is needed to allow time for the MySQL container to start.

## API Specification

The full OpenAPI 3.0 specification is available at `gen/http/openapi3.yaml` or served at `/openapi.json` when the application is running.
