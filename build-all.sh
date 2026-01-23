#!/bin/bash

# Multi-platform build script for aiassist
# This script builds binaries for multiple OS and architectures

set -e

# Color output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get version info
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build variables
BINARY_NAME="aiassist"
OUTPUT_DIR="dist"
LDFLAGS="-X main.Version=${VERSION} -X main.Commit=${COMMIT} -s -w"

# Platforms to build for
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/arm"
    "linux/386"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
    "windows/arm64"
    "windows/386"
    "freebsd/amd64"
    "freebsd/arm64"
)

echo -e "${BLUE}Building ${BINARY_NAME} version ${VERSION}${NC}"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    # Split platform into OS and ARCH
    IFS='/' read -r os arch <<< "$platform"
    
    # Set output name
    output_name="${OUTPUT_DIR}/${BINARY_NAME}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo -e "${BLUE}Building for ${os}/${arch}...${NC}"
    
    # Build
    GOOS=$os GOARCH=$arch go build -ldflags "${LDFLAGS}" -o "$output_name" ./cmd/aiassist/
    
    if [ $? -eq 0 ]; then
        # Get file size
        size=$(du -h "$output_name" | cut -f1)
        echo -e "${GREEN}✓${NC} Built: $output_name (${size})"
    else
        echo "✗ Failed to build for ${os}/${arch}"
    fi
    
    echo ""
done

echo -e "${GREEN}All builds completed!${NC}"
echo ""
echo "Output directory: $OUTPUT_DIR"
echo "Total files: $(ls -1 "$OUTPUT_DIR" | wc -l | tr -d ' ')"
echo ""
echo "To create release archives, run:"
echo "  cd $OUTPUT_DIR && for file in *; do tar czf \${file}.tar.gz \$file; done"
