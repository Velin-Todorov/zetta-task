FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o bookstore ./cmd/bookstore

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/bookstore .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/sql ./sql
COPY --from=builder /app/gen/http/openapi3.json ./gen/http/openapi3.json
COPY --from=builder /app/public ./public

EXPOSE 8080

CMD ["./bookstore", "--http-port", "8080"]
