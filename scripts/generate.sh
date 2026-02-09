#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${YELLOW}generating proto files...${NC}"

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
PROTO_DIR="${ROOT_DIR}/proto"
OUT_DIR="${ROOT_DIR}/generated/proto"

if ! command -v protoc &> /dev/null; then
    echo -e "${RED}error: protoc is not installed${NC}"
    echo "install it:"
    echo "  mac: brew install protobuf"
    echo "  ubuntu: apt install -y protobuf-compiler"
    exit 1
fi

if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}installing protoc-gen-go-grpc...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

export PATH="${PATH}:$(go env GOPATH)/bin"

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

for proto_file in $(find "${PROTO_DIR}" -name "*.proto"); do
    echo -e "  generating: ${proto_file}"
    
    protoc \
        --proto_path="${PROTO_DIR}" \
        --go_out="${OUT_DIR}" \
        --go_opt=paths=source_relative \
        --go-grpc_out="${OUT_DIR}" \
        --go-grpc_opt=paths=source_relative \
        "${proto_file}"
done

echo -e "${GREEN}done! generated files in ${OUT_DIR}${NC}"