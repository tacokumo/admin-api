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

# Reset functions
confirm_reset() {
    echo "🥙 Tacokumo Admin API - Development Environment Reset"
    echo "===================================================="
    echo ""
    print_warning "This will:"
    echo "  • Stop all Docker Compose services"
    echo "  • Remove all containers and volumes (data will be lost)"
    echo "  • Clean up Docker images"
    echo "  • Remove development certificates"
    echo "  • Clear any cached build artifacts"
    echo ""
    print_error "ALL DATABASE DATA WILL BE LOST!"
    echo ""

    read -p "Are you sure you want to reset the development environment? [y/N]: " -r
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Reset cancelled"
        exit 0
    fi
}

stop_services() {
    print_header "Stopping Services"

    if docker compose ps -q &> /dev/null; then
        docker compose down --volumes --remove-orphans
        print_success "Services stopped and containers removed"
    else
        print_info "No services were running"
    fi
}

cleanup_docker() {
    print_header "Cleaning Up Docker Resources"

    # Remove project-specific images
    local project_images
    project_images=$(docker images --filter=reference="admin-api*" -q 2>/dev/null || true)

    if [[ -n "$project_images" ]]; then
        docker rmi $project_images 2>/dev/null || true
        print_success "Removed project Docker images"
    else
        print_info "No project Docker images to remove"
    fi

    # Prune unused volumes (be careful with system-wide cleanup)
    docker volume prune -f &> /dev/null || true
    print_success "Pruned unused Docker volumes"
}

remove_certificates() {
    print_header "Removing Development Certificates"

    local certs_dir="${PROJECT_ROOT}/develop/certs"

    if [[ -d "$certs_dir" ]]; then
        rm -rf "$certs_dir"
        print_success "Removed development certificates"
    else
        print_info "No certificates directory found"
    fi
}

cleanup_build_artifacts() {
    print_header "Cleaning Build Artifacts"

    # Remove Go build artifacts
    if [[ -d "${PROJECT_ROOT}/bin" ]]; then
        rm -rf "${PROJECT_ROOT}/bin"
        print_success "Removed Go build artifacts"
    else
        print_info "No Go build artifacts found"
    fi

    # Clean Go module cache (be careful, this affects the entire system)
    # Uncomment if you want to clean the entire Go module cache
    # go clean -modcache 2>/dev/null || true
    # print_info "Cleaned Go module cache"

    # Remove any temporary files
    find "${PROJECT_ROOT}" -name "*.tmp" -type f -delete 2>/dev/null || true
    find "${PROJECT_ROOT}" -name ".DS_Store" -type f -delete 2>/dev/null || true
    print_success "Removed temporary files"
}

cleanup_logs() {
    print_header "Cleaning Up Logs"

    # Remove any log files (if they exist)
    find "${PROJECT_ROOT}" -name "*.log" -type f -delete 2>/dev/null || true
    print_success "Removed log files"
}

setup_fresh_environment() {
    print_header "Setting Up Fresh Environment"

    # Recreate certificates directory
    mkdir -p "${PROJECT_ROOT}/develop/certs"
    print_success "Created certificates directory"

    # Recreate bin directory
    mkdir -p "${PROJECT_ROOT}/bin"
    print_success "Created bin directory"

    print_info "Environment reset complete!"
    echo ""
    print_info "Next steps:"
    echo "  1. Configure your .env file (copy from .env.example if needed)"
    echo "  2. Start services: make docker-compose-up"
    echo "  3. Verify setup: bash scripts/verify-setup.sh"
}

print_completion_message() {
    print_header "Reset Complete"

    print_success "Development environment has been reset successfully!"
    echo ""
    print_info "To start fresh:"
    echo "  • Configure environment: cp .env.example .env && vim .env"
    echo "  • Start services:        make docker-compose-up"
    echo "  • Verify setup:          bash scripts/verify-setup.sh"
    echo "  • Access frontend:       http://localhost:3000"
    echo "  • Access API:            https://localhost:8444"
    echo ""
    print_warning "Remember to configure your GitHub OAuth credentials in .env!"
}

main() {
    cd "$PROJECT_ROOT"

    # Check if we're in the right directory
    if [[ ! -f "compose.yaml" ]]; then
        print_error "Could not find compose.yaml. Are you in the right directory?"
        exit 1
    fi

    confirm_reset
    stop_services
    cleanup_docker
    remove_certificates
    cleanup_build_artifacts
    cleanup_logs
    setup_fresh_environment
    print_completion_message
}

# Allow running specific functions for testing
if [[ $# -gt 0 ]]; then
    case "$1" in
        "stop")
            stop_services
            ;;
        "cleanup-docker")
            cleanup_docker
            ;;
        "cleanup-certs")
            remove_certificates
            ;;
        "cleanup-build")
            cleanup_build_artifacts
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Available commands: stop, cleanup-docker, cleanup-certs, cleanup-build"
            exit 1
            ;;
    esac
else
    main "$@"
fi