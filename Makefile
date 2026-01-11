.PHONY: help build run test clean deps migrate-up migrate-down

# Variables
APP_NAME=tmdb-crawler
BUILD_DIR=build
GO_FILES=$(shell find . -name '*.go' -type f)

# Default target
help:
	@echo "Available commands:"
	@echo "  make deps         - Install dependencies"
	@echo "  make build        - Build the application"
	@echo "  make run          - Run the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build files"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build the application
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(APP_NAME) .

# Run the application
run: deps
	@echo "Running $(APP_NAME)..."
	go run .

# Run tests
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Clean build files
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(APP_NAME)

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker-compose up -d

# Docker stop
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose stop

# Docker down
docker-down:
	@echo "Stopping and removing Docker containers..."
	docker-compose down

# Docker logs
docker-logs:
	@echo "Showing Docker logs..."
	docker-compose logs -f tmdb-crawler

# Docker rebuild
docker-rebuild:
	@echo "Rebuilding Docker image..."
	docker-compose build --no-cache
	docker-compose up -d

# Docker shell
docker-shell:
	@echo "Opening shell in container..."
	docker-compose exec tmdb-crawler sh

# Docker ps
docker-ps:
	@echo "Listing Docker containers..."
	docker-compose ps

# Docker clean
docker-clean:
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f migrations/001_init_schema.sql

migrate-down:
	@echo "Rolling back database migrations..."
	@echo "Please implement rollback manually"
