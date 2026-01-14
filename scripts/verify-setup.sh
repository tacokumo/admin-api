#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Verification functions
check_prerequisites() {
    print_header "Checking Prerequisites"

    local missing=0

    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        missing=1
    else
        print_success "Docker is available"
    fi

    if ! command -v docker compose &> /dev/null; then
        print_error "Docker Compose is not available"
        missing=1
    else
        print_success "Docker Compose is available"
    fi

    if ! command -v make &> /dev/null; then
        print_error "Make is not installed or not in PATH"
        missing=1
    else
        print_success "Make is available"
    fi

    if [[ $missing -ne 0 ]]; then
        print_error "Please install missing prerequisites"
        exit 1
    fi
}

check_environment_config() {
    print_header "Checking Environment Configuration"

    if [[ ! -f "${PROJECT_ROOT}/.env" ]]; then
        print_warning ".env file not found"
        print_info "Copy .env.example to .env and configure your settings:"
        print_info "  cp .env.example .env"
        print_info "  # Edit .env with your GitHub OAuth credentials"
    else
        print_success ".env file exists"

        # Check for required environment variables
        local required_vars=("GITHUB_CLIENT_ID" "GITHUB_CLIENT_SECRET" "GITHUB_ALLOWED_ORGS")
        local missing_vars=0

        for var in "${required_vars[@]}"; do
            if ! grep -q "^${var}=" "${PROJECT_ROOT}/.env" 2>/dev/null || \
               grep -q "^${var}=your_" "${PROJECT_ROOT}/.env" 2>/dev/null || \
               grep -q "^${var}=$" "${PROJECT_ROOT}/.env" 2>/dev/null; then
                print_warning "${var} is not configured in .env"
                missing_vars=1
            else
                print_success "${var} is configured"
            fi
        done

        if [[ $missing_vars -ne 0 ]]; then
            print_warning "Some environment variables need configuration"
            print_info "Edit .env file and set the required GitHub OAuth values"
        fi
    fi
}

check_docker_services() {
    print_header "Checking Docker Services"

    # Check if services are running
    if ! docker compose ps -q &> /dev/null; then
        print_error "Docker Compose services are not running"
        print_info "Start services with: make docker-compose-up"
        return 1
    fi

    local services=("postgresql" "valkey" "admin_api" "frontend")
    local all_healthy=0

    for service in "${services[@]}"; do
        local status
        status=$(docker compose ps "$service" --format "{{.State}}" 2>/dev/null || echo "missing")

        case $status in
            "running")
                print_success "$service is running"
                ;;
            "")
                print_warning "$service status unknown"
                all_healthy=1
                ;;
            *)
                print_error "$service is not running (status: $status)"
                all_healthy=1
                ;;
        esac
    done

    if [[ $all_healthy -ne 0 ]]; then
        print_warning "Some services are not healthy"
        print_info "Check service logs with: docker compose logs [service-name]"
    fi
}

check_service_endpoints() {
    print_header "Checking Service Endpoints"

    # Wait a bit for services to be ready
    sleep 2

    # Check Frontend (HTTP)
    if curl -sf "http://localhost:3000/" > /dev/null 2>&1; then
        print_success "Frontend is accessible at http://localhost:3000"
    else
        print_error "Frontend is not accessible at http://localhost:3000"
    fi

    # Check API Health (HTTPS, ignore certificate)
    if curl -sfk "https://localhost:8444/v1alpha1/health/liveness" > /dev/null 2>&1; then
        print_success "Admin API is accessible at https://localhost:8444"
    else
        print_error "Admin API is not accessible at https://localhost:8444"
        print_info "Check if certificates are generated: ls -la develop/certs/"
    fi

    # Check Database (via docker)
    if docker compose exec -T postgresql pg_isready -h localhost -U admin_api -d tacokumo_admin_db > /dev/null 2>&1; then
        print_success "PostgreSQL database is accessible"
    else
        print_error "PostgreSQL database is not accessible"
    fi

    # Check Redis (via docker)
    if docker compose exec -T valkey valkey-cli ping > /dev/null 2>&1; then
        print_success "Valkey/Redis is accessible"
    else
        print_error "Valkey/Redis is not accessible"
    fi
}

check_certificates() {
    print_header "Checking TLS Certificates"

    local cert_file="${PROJECT_ROOT}/develop/certs/api-server.crt"
    local key_file="${PROJECT_ROOT}/develop/certs/api-server.key"

    if [[ -f "$cert_file" ]] && [[ -f "$key_file" ]]; then
        print_success "TLS certificates exist"

        # Check certificate expiry
        local days_left
        days_left=$(openssl x509 -in "$cert_file" -noout -days 2>/dev/null || echo "0")
        if [[ $days_left -gt 30 ]]; then
            print_success "Certificate is valid for $days_left more days"
        else
            print_warning "Certificate expires in $days_left days"
            print_info "Regenerate with: make docker-compose-up (includes cert generation)"
        fi
    else
        print_error "TLS certificates are missing"
        print_info "Generate certificates with: bash scripts/generate-dev-certs.sh"
    fi
}

run_quick_api_tests() {
    print_header "Running Quick API Tests"

    # Test health endpoint
    local health_response
    health_response=$(curl -sfk "https://localhost:8444/v1alpha1/health/liveness" 2>/dev/null || echo "")

    if [[ -n "$health_response" ]]; then
        print_success "Health endpoint responded"
    else
        print_error "Health endpoint is not responding"
        return 1
    fi

    # Test CORS headers
    local cors_headers
    cors_headers=$(curl -sfkI "https://localhost:8444/v1alpha1/health/liveness" 2>/dev/null | grep -i "access-control" || echo "")

    if [[ -n "$cors_headers" ]]; then
        print_success "CORS headers are configured"
    else
        print_warning "CORS headers not found in response"
    fi
}

print_summary() {
    print_header "Verification Summary"

    print_info "Services should be accessible at:"
    echo "  • Frontend:  http://localhost:3000"
    echo "  • Admin API: https://localhost:8444"
    echo "  • PostgreSQL: localhost:5432 (via Docker)"
    echo "  • Valkey:     localhost:6379 (via Docker)"
    echo ""
    print_info "Useful commands:"
    echo "  • View logs:        docker compose logs [service]"
    echo "  • Restart services: make docker-compose-down && make docker-compose-up"
    echo "  • Reset environment: bash scripts/reset-dev-env.sh"
    echo "  • Run tests:        make test"
    echo ""

    if [[ ! -f "${PROJECT_ROOT}/.env" ]] || grep -q "your_" "${PROJECT_ROOT}/.env" 2>/dev/null; then
        print_warning "Don't forget to configure your .env file with GitHub OAuth credentials!"
    fi
}

main() {
    echo "🥙 Tacokumo Admin API - Setup Verification"
    echo "========================================"

    check_prerequisites
    check_environment_config
    check_docker_services
    check_service_endpoints
    check_certificates
    run_quick_api_tests

    print_summary

    print_success "Verification complete!"
}

main "$@"