FROM golang:alpine AS builder

# Install required packages for cgo and sqlite
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
# CGO_ENABLED=1 is required for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o todo_app .

# Runner stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache sqlite tzdata

# Copy binary and required files
COPY --from=builder /build/todo_app .
COPY --from=builder /build/config.ini .
COPY --from=builder /build/app/views ./app/views

# Update config to point to the mounted volume for database persistence
RUN sed -i 's|name = webapp.sql|name = /app/data/webapp.sql|' config.ini

# Ensure the data directory exists
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./todo_app"]
