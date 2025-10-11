#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CERTS_DIR="${PROJECT_ROOT}/develop/certs"

generate_api_server_cert() {
  echo "Generating API server certificate (for HTTPS)..."
  openssl genrsa -out "${CERTS_DIR}/api-server.key" 2048
  openssl req -new -x509 -key "${CERTS_DIR}/api-server.key" \
    -out "${CERTS_DIR}/api-server.crt" \
    -days 365 \
    -subj "/C=JP/ST=Tokyo/L=Tokyo/O=TacokumoAPI/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
}

mkdir -p "${CERTS_DIR}"

# Generate API server certificate
generate_api_server_cert

echo "Certificates generated successfully in ${CERTS_DIR}"
echo ""
echo "API server certificates (HTTPS):"
echo "  - ${CERTS_DIR}/api-server.key"
echo "  - ${CERTS_DIR}/api-server.crt"
