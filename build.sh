#!/bin/bash

# Build script for aiassist with version injection

# Get git version information
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Build output name
BINARY_NAME=${1:-aiassist}

# Construct ldflags
LDFLAGS="-X main.Version=${VERSION} -X main.Commit=${COMMIT}"

echo "Building ${BINARY_NAME}"
echo "  Version: ${VERSION}"
echo "  Commit:  ${COMMIT}"
echo "  Build Time: ${BUILD_TIME}"
echo ""

# Build the binary
go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/aiassist/

if [ $? -eq 0 ]; then
    echo "✓ Build successful: ${BINARY_NAME}"
    # Show version info
    ./"${BINARY_NAME}" version
else
    echo "✗ Build failed"
    exit 1
fi
