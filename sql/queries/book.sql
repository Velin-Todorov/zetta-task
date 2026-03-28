-- name: GetBook :one
SELECT * FROM `books` WHERE id = ?;

-- name: CreateBook :execresult
INSERT INTO `books` (title, author, published_at)
VALUES (?, ?, ?);

-- name: UpdateBook :execresult
UPDATE `books`
SET title = ?, author = ?, published_at = ?
WHERE id = ?;

-- name: UpdateBookCover :execresult
UPDATE `books` SET cover_path = ? WHERE id = ?;

-- name: DeleteBook :exec
DELETE FROM `books` WHERE id = ?;
