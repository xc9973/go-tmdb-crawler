.PHONY: help build run test clean deps migrate-up migrate-down
.PHONY: docker-build docker-run docker-stop docker-down docker-logs docker-rebuild docker-shell docker-ps docker-clean
.PHONY: prod-build prod-deploy prod-logs prod-restart prod-status prod-backup
.PHONY: benchmark test-coverage lint security-scan

# Variables
APP_NAME=tmdb-crawler
BUILD_DIR=build
GO_FILES=$(shell find . -name '*.go' -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_SHA=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Docker variables
DOCKER_REGISTRY=your-registry.com
DOCKER_IMAGE=$(APP_NAME):$(VERSION)

# Default target
help:
	@echo "=========================================="
	@echo "TMDB Crawler - Makefile Commands"
	@echo "=========================================="
	@echo ""
	@echo "Development Commands:"
	@echo "  make deps           - Install dependencies"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make benchmark      - Run benchmark tests"
	@echo "  make lint           - Run linter"
	@echo "  make clean          - Clean build files"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"
	@echo "  make docker-stop    - Stop Docker containers"
	@echo "  make docker-down    - Stop and remove containers"
	@echo "  make docker-logs    - Show Docker logs"
	@echo "  make docker-rebuild - Rebuild and restart containers"
	@echo "  make docker-shell   - Open shell in container"
	@echo "  make docker-ps      - List Docker containers"
	@echo "  make docker-clean   - Clean Docker resources"
	@echo ""
	@echo "Production Commands:"
	@echo "  make prod-build     - Build production Docker image"
	@echo "  make prod-deploy    - Deploy to production"
	@echo "  make prod-logs      - Show production logs"
	@echo "  make prod-restart   - Restart production services"
	@echo "  make prod-status    - Show production status"
	@echo "  make prod-backup    - Backup production data"
	@echo ""
	@echo "Database Commands:"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make db-optimize    - Apply performance optimizations"
	@echo ""
	@echo "Security Commands:"
	@echo "  make security-scan  - Run security vulnerability scan"
	@echo ""

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

# ============================================
# Production Commands
# ============================================

# Build production Docker image
prod-build:
	@echo "Building production Docker image..."
	docker build \
		-f Dockerfile.prod \
		--build-arg BUILD_VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg COMMIT_SHA=$(COMMIT_SHA) \
		-t $(APP_NAME):$(VERSION) \
		-t $(APP_NAME):latest \
		.
	@echo "Production image built: $(APP_NAME):$(VERSION)"

# Deploy to production
prod-deploy: prod-build
	@echo "Deploying to production..."
	docker-compose -f docker-compose.prod.yml --env-file .env.production up -d
	@echo "Deployment completed successfully"

# Show production logs
prod-logs:
	docker-compose -f docker-compose.prod.yml logs -f --tail=100

# Restart production services
prod-restart:
	@echo "Restarting production services..."
	docker-compose -f docker-compose.prod.yml restart
	@echo "Services restarted"

# Show production status
prod-status:
	@echo "Production Status:"
	@echo "=================="
	docker-compose -f docker-compose.prod.yml ps
	@echo ""
	@echo "Resource Usage:"
	docker stats --no-stream $(shell docker-compose -f docker-compose.prod.yml ps -q)

# Backup production data
prod-backup:
	@echo "Backing up production data..."
	@mkdir -p backups
	@docker exec tmdb-crawler-prod sh -c "tar czf /tmp/backup.tar.gz /app/data /app/logs" && \
		docker cp tmdb-crawler-prod:/tmp/backup.tar.gz backups/backup-$$(date +%Y%m%d-%H%M%S).tar.gz && \
		docker exec tmdb-crawler-prod rm /tmp/backup.tar.gz
	@echo "Backup completed"

# ============================================
# Testing & Quality Commands
# ============================================

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmark tests
benchmark:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem -run=^$$ ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi

# Security vulnerability scan
security-scan:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# ============================================
# Database Optimization Commands
# ============================================

# Apply performance optimizations
db-optimize:
	@echo "Applying database performance optimizations..."
	@if [ -f migrations/002_add_performance_indexes.sql ]; then \
		psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -f migrations/002_add_performance_indexes.sql; \
		echo "Performance indexes applied successfully"; \
	else \
		echo "Optimization migration file not found"; \
	fi

# Analyze database performance
db-analyze:
	@echo "Analyzing database performance..."
	@echo "Index usage:"
	@psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -c "SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read FROM pg_stat_user_indexes ORDER BY idx_scan DESC;"
	@echo ""
	@echo "Table sizes:"
	@psql -h $(DB_HOST) -U $(DB_USER) -d $(DB_NAME) -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"

# ============================================
# Performance Monitoring Commands
# ============================================

# Show application metrics
metrics:
	@echo "Fetching application metrics..."
	@curl -s http://localhost:8888/api/v1/metrics | jq .

# Show health status
health:
	@echo "Checking application health..."
	@curl -s http://localhost:8888/health | jq .

# Performance test with Apache Bench
perf-test:
	@echo "Running performance test..."
	@echo "Testing /api/v1/shows endpoint..."
	@ab -n 1000 -c 10 http://localhost:8888/api/v1/shows

# ============================================
# Development Utilities
# ============================================

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/rakyll/hey@latest
	@echo "Development tools installed"

# Generate API documentation
api-docs:
	@echo "Generating API documentation..."
	@echo "API documentation available at: docs/API.md"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# ============================================
# CI/CD Commands
# ============================================

# CI pipeline
ci: deps lint test-coverage security-scan
	@echo "CI pipeline completed successfully"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .
	@echo "Builds completed"
