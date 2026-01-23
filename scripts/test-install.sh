#!/bin/bash

# Test installation script locally
# This script simulates the installation process for testing purposes

set -e

echo "========================================"
echo "Testing aiassist Installation Script"
echo "========================================"
echo ""

# Check if install.sh exists
if [ ! -f "scripts/install.sh" ]; then
    echo "Error: scripts/install.sh not found"
    exit 1
fi

# Detect OS
echo "1. Detecting Operating System..."
OS_TYPE=""
ARCH_TYPE=""

case "$(uname -s)" in
    Linux*)
        OS_TYPE="linux"
        ;;
    Darwin*)
        OS_TYPE="darwin"
        ;;
    *)
        OS_TYPE="unknown"
        ;;
esac

case "$(uname -m)" in
    x86_64|amd64)
        ARCH_TYPE="amd64"
        ;;
    aarch64|arm64)
        ARCH_TYPE="arm64"
        ;;
    *)
        ARCH_TYPE="unknown"
        ;;
esac

echo "   ✓ OS: ${OS_TYPE}"
echo "   ✓ Arch: ${ARCH_TYPE}"
echo ""

# Check for build binary
echo "2. Checking for built binary..."
BINARY_PATH="build/aiassist"

if [ ! -f "$BINARY_PATH" ]; then
    echo "   Building binary first..."
    make build
    if [ ! -f "$BINARY_PATH" ]; then
        echo "   Error: Failed to build binary"
        exit 1
    fi
fi

echo "   ✓ Binary found at: $BINARY_PATH"
echo ""

# Test binary
echo "3. Testing binary..."
chmod +x "$BINARY_PATH"
VERSION_OUTPUT=$($BINARY_PATH version 2>&1 || echo "")

if [ -z "$VERSION_OUTPUT" ]; then
    echo "   Error: Binary doesn't work"
    exit 1
fi

echo "   ✓ Binary version output:"
echo "   $VERSION_OUTPUT"
echo ""

# Check install script syntax
echo "4. Checking scripts/install.sh syntax..."
bash -n scripts/install.sh
if [ $? -eq 0 ]; then
    echo "   ✓ Syntax check passed"
else
    echo "   Error: Syntax errors found in scripts/install.sh"
    exit 1
fi
echo ""

# Check required tools
echo "5. Checking required tools..."
REQUIRED_TOOLS=("curl" "tar" "chmod" "mv")

for tool in "${REQUIRED_TOOLS[@]}"; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo "   ✓ $tool: found"
    else
        echo "   ✗ $tool: not found"
    fi
done
echo ""

# Simulate installation directory structure
echo "6. Testing installation directory creation..."
TEST_DIR=$(mktemp -d)
echo "   Test directory: $TEST_DIR"

# Copy binary to test dir
cp "$BINARY_PATH" "$TEST_DIR/aiassist"
chmod +x "$TEST_DIR/aiassist"

# Test execution from test dir
cd "$TEST_DIR"
EXEC_TEST=$(./aiassist version 2>&1 || echo "")
if [ -n "$EXEC_TEST" ]; then
    echo "   ✓ Binary executes successfully from test directory"
else
    echo "   ✗ Binary execution failed"
fi

# Cleanup
cd - > /dev/null
rm -rf "$TEST_DIR"
echo ""

# Summary
echo "========================================"
echo "Installation Script Test Complete!"
echo "========================================"
echo ""
echo "Next steps to test full installation:"
echo "  1. Create a release on GitHub"
echo "  2. Upload binaries for different platforms"
echo "  3. Test installation script:"
echo "     bash scripts/install.sh"
echo ""
