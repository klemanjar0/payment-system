#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
SERVICES_DIR="${ROOT_DIR}/services"

if ! command -v sqlc &> /dev/null; then
    echo -e "${RED}error: sqlc is not installed${NC}"
    echo "install it:"
    echo "  mac: brew install sqlc"
    echo "  go:  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
    exit 1
fi

generate_service() {
    local service="$1"
    local service_dir="${SERVICES_DIR}/${service}"
    local sqlc_config="${service_dir}/sqlc.yaml"

    if [ ! -f "${sqlc_config}" ]; then
        echo -e "${YELLOW}skipping ${service}: no sqlc.yaml found${NC}"
        return 0
    fi

    echo -e "  generating sqlc for ${YELLOW}${service}${NC}..."
    (cd "${service_dir}" && sqlc generate)
    echo -e "  ${GREEN}done: ${service}${NC}"
}

if [ -n "$1" ]; then
    service="$1"
    if [ ! -d "${SERVICES_DIR}/${service}" ]; then
        echo -e "${RED}error: service '${service}' not found in ${SERVICES_DIR}${NC}"
        exit 1
    fi
    echo -e "${YELLOW}generating sqlc for service: ${service}${NC}"
    generate_service "${service}"
else
    echo -e "${YELLOW}generating sqlc for all services...${NC}"
    for service_dir in "${SERVICES_DIR}"/*/; do
        service=$(basename "${service_dir}")
        generate_service "${service}"
    done
fi

echo -e "${GREEN}sqlc generation complete!${NC}"
