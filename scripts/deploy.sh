#!/bin/bash

# ============================================
# TMDB Crawler - Production Deployment Script
# Version: 2.0
# Created: 2026-01-12
# ============================================

set -e  # Exit on error
set -u  # Exit on undefined variable

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
ENV_FILE="${PROJECT_DIR}/.env.production"
COMPOSE_FILE="${PROJECT_DIR}/docker-compose.prod.yml"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    log_success "All requirements met"
}

check_env_file() {
    log_info "Checking environment configuration..."
    
    if [ ! -f "$ENV_FILE" ]; then
        log_warning "Environment file not found: $ENV_FILE"
        log_info "Creating from example..."
        
        if [ -f "${PROJECT_DIR}/.env.production.example" ]; then
            cp "${PROJECT_DIR}/.env.production.example" "$ENV_FILE"
            log_warning "Please edit $ENV_FILE and fill in the required values"
            log_warning "Required variables: TMDB_API_KEY, TELEGRAPH_TOKEN, ADMIN_API_KEY"
            read -p "Press Enter to continue after editing..."
        else
            log_error "Example environment file not found"
            exit 1
        fi
    fi
    
    # Source environment file
    source "$ENV_FILE"
    
    # Check required variables
    local required_vars=("TMDB_API_KEY" "TELEGRAPH_TOKEN" "ADMIN_API_KEY")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ] || [[ "${!var}" == *"your_"* ]]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        log_error "Missing or invalid environment variables: ${missing_vars[*]}"
        log_error "Please set these variables in $ENV_FILE"
        exit 1
    fi
    
    log_success "Environment configuration is valid"
}

build_image() {
    log_info "Building production Docker image..."
    
    cd "$PROJECT_DIR"
    
    # Get version info
    VERSION=${BUILD_VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "latest")}
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    COMMIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    docker build \
        -f Dockerfile.prod \
        --build-arg BUILD_VERSION="$VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        --build-arg COMMIT_SHA="$COMMIT_SHA" \
        -t tmdb-crawler:"$VERSION" \
        -t tmdb-crawler:latest \
        .
    
    log_success "Docker image built successfully"
}

deploy_services() {
    log_info "Deploying services..."
    
    cd "$PROJECT_DIR"
    
    # Determine which profiles to enable
    local profiles=""
    if [ "${DB_TYPE:-sqlite}" = "postgres" ]; then
        profiles="${profiles}with-postgres,"
    fi
    
    if [ "${ENABLE_NGINX:-false}" = "true" ]; then
        profiles="${profiles}with-nginx,"
    fi
    
    if [ "${ENABLE_REDIS:-false}" = "true" ]; then
        profiles="${profiles}with-redis,"
    fi
    
    # Remove trailing comma
    profiles=${profiles%,}
    
    # Deploy
    if [ -n "$profiles" ]; then
        COMPOSE_PROFILES="$profiles" docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
    else
        docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
    fi
    
    log_success "Services deployed successfully"
}

wait_for_service() {
    log_info "Waiting for service to be ready..."
    
    local max_attempts=30
    local attempt=0
    local port=${APP_PORT:-8888}
    
    while [ $attempt -lt $max_attempts ]; do
        if curl -sf "http://localhost:${port}/health" > /dev/null 2>&1; then
            log_success "Service is ready"
            return 0
        fi
        
        attempt=$((attempt + 1))
        sleep 2
    done
    
    log_error "Service failed to start within expected time"
    return 1
}

show_status() {
    log_info "Service status:"
    echo ""
    
    cd "$PROJECT_DIR"
    docker-compose -f "$COMPOSE_FILE" ps
    
    echo ""
    log_info "Resource usage:"
    docker stats --no-stream $(docker-compose -f "$COMPOSE_FILE" ps -q) 2>/dev/null || true
}

show_logs() {
    log_info "Recent logs:"
    echo ""
    
    cd "$PROJECT_DIR"
    docker-compose -f "$COMPOSE_FILE" logs --tail=50 tmdb-crawler
}

run_migrations() {
    log_info "Running database migrations..."
    
    # Check if PostgreSQL is enabled
    if [ "${DB_TYPE:-sqlite}" = "postgres" ]; then
        log_info "Applying performance indexes to PostgreSQL..."
        
        # Wait for PostgreSQL to be ready
        local max_attempts=30
        local attempt=0
        
        while [ $attempt -lt $max_attempts ]; do
            if docker exec tmdb-postgres-prod pg_isready -U "${DB_USER:-tmdb}" -d "${DB_NAME:-tmdb}" > /dev/null 2>&1; then
                break
            fi
            attempt=$((attempt + 1))
            sleep 2
        done
        
        # Apply migrations
        if [ -f "${PROJECT_DIR}/migrations/001_init_schema.sql" ]; then
            docker exec -i tmdb-postgres-prod psql -U "${DB_USER:-tmdb}" -d "${DB_NAME:-tmdb}" < "${PROJECT_DIR}/migrations/001_init_schema.sql"
        fi
        
        if [ -f "${PROJECT_DIR}/migrations/002_add_performance_indexes.sql" ]; then
            docker exec -i tmdb-postgres-prod psql -U "${DB_USER:-tmdb}" -d "${DB_NAME:-tmdb}" < "${PROJECT_DIR}/migrations/002_add_performance_indexes.sql"
        fi
        
        log_success "Database migrations completed"
    else
        log_info "SQLite database will be initialized automatically"
    fi
}

backup_data() {
    log_info "Backing up existing data..."
    
    local backup_dir="${PROJECT_DIR}/backups"
    mkdir -p "$backup_dir"
    
    local backup_file="${backup_dir}/backup-$(date +%Y%m%d-%H%M%S).tar.gz"
    
    if docker ps | grep -q tmdb-crawler-prod; then
        docker exec tmdb-crawler-prod sh -c "tar czf /tmp/backup.tar.gz /app/data /app/logs 2>/dev/null || true"
        docker cp tmdb-crawler-prod:/tmp/backup.tar.gz "$backup_file" 2>/dev/null || true
        docker exec tmdb-crawler-prod rm -f /tmp/backup.tar.gz 2>/dev/null || true
        
        if [ -f "$backup_file" ]; then
            log_success "Backup created: $backup_file"
        else
            log_warning "Backup failed or no data to backup"
        fi
    else
        log_warning "Service not running, skipping backup"
    fi
}

# Main deployment flow
main() {
    log_info "Starting deployment..."
    echo ""
    
    # Parse arguments
    SKIP_BUILD=false
    SKIP_MIGRATIONS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-migrations)
                SKIP_MIGRATIONS=true
                shift
                ;;
            --backup-only)
                backup_data
                exit 0
                ;;
            --status-only)
                show_status
                exit 0
                ;;
            --logs-only)
                show_logs
                exit 0
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --skip-build       Skip building Docker image"
                echo "  --skip-migrations  Skip database migrations"
                echo "  --backup-only      Only backup data"
                echo "  --status-only      Only show service status"
                echo "  --logs-only        Only show logs"
                echo "  -h, --help         Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Execute deployment steps
    check_requirements
    check_env_file
    
    if [ "$SKIP_BUILD" = false ]; then
        build_image
    fi
    
    backup_data
    deploy_services
    
    if [ "$SKIP_MIGRATIONS" = false ]; then
        run_migrations
    fi
    
    wait_for_service
    show_status
    
    echo ""
    log_success "Deployment completed successfully!"
    echo ""
    log_info "Service URL: http://localhost:${APP_PORT:-8888}"
    log_info "Health Check: http://localhost:${APP_PORT:-8888}/health"
    log_info "API Metrics: http://localhost:${APP_PORT:-8888}/api/v1/metrics"
    echo ""
    log_info "To view logs: docker-compose -f $COMPOSE_FILE logs -f"
    log_info "To stop: docker-compose -f $COMPOSE_FILE down"
}

# Run main function
main "$@"
