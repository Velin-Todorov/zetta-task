CREATE TABLE `books` (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    cover_path VARCHAR(255),
    published_at DATE NOT NULL
);