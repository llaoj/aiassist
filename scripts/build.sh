#!/bin/bash

# Build script for aiassist with version injection

# Get git version information
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%d")

# Build output name
BINARY_NAME="aiassist"

# Construct ldflags
LDFLAGS="-X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE} -s -w"

echo "Building ${BINARY_NAME}"
echo "  Version:    ${VERSION}"
echo "  Commit:     ${COMMIT}"
echo "  Build Date: ${BUILD_DATE}"
echo ""

# Build the binary with static linking
CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/aiassist/

if [ $? -eq 0 ]; then
    echo "✓ Build successful: ${BINARY_NAME}"
    # Show version info
    ./"${BINARY_NAME}" version
else
    echo "✗ Build failed"
    exit 1
fi
