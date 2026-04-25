#!/bin/bash

# GAIA Repository Setup & Initialization Script
# This script handles both first-time scaffolding and dependency installation.

set -e

echo "🚀 Setting up GAIA Development Environment..."

# 0. Dependency Checks (Ensure the environment is ready)
check_dep() {
    if ! command -v "$1" &> /dev/null; then
        echo "❌ Error: $1 is not installed. Please install it to proceed."
        exit 1
    fi
}

check_dep "go"
check_dep "npm"
check_dep "python3"

# 1. Kernel Setup (Go)
echo "📦 Setting up Go Kernel..."
mkdir -p src/kernel
cd src/kernel
if [ ! -f go.mod ]; then
    echo "  - Initializing go.mod..."
    go mod init gaia/kernel
else
    echo "  - Syncing dependencies..."
    go mod download
fi
cd ../..

# 2. TypeScript SDK Setup (Node.js)
echo "📦 Setting up TypeScript SDK..."
mkdir -p libs/sdk-ts
cd libs/sdk-ts
if [ ! -f package.json ]; then
    echo "  - Initializing package.json..."
    npm init -y > /dev/null
    # Update name to @gaia/sdk
    sed -i '' 's/"name": "sdk-ts"/"name": "@gaia\/sdk"/g' package.json
else
    echo "  - Installing npm dependencies..."
    npm install
fi
cd ../..

# 3. Python SDK Setup (Python)
echo "📦 Setting up Python SDK..."
mkdir -p libs/sdk-py
cd libs/sdk-py
if [ ! -f pyproject.toml ]; then
    echo "  - Creating pyproject.toml..."
    cat <<EOF > pyproject.toml
[project]
name = "gaia-sdk"
version = "0.1.0"
description = "Official GAIA Agent SDK for Python"
dependencies = []

[build-system]
requires = ["setuptools", "wheel"]
build-backend = "setuptools.build_meta"
EOF
fi
# Optional: Install in editable mode if a virtualenv is detected
if [[ "$VIRTUAL_ENV" != "" ]]; then
    echo "  - Installing in editable mode (VIRTUAL_ENV detected)..."
    pip install -e .
fi
cd ../..

echo ""
echo "✅ GAIA Environment Setup Complete!"
echo "----------------------------------------"
echo "Kernel: ./src/kernel/"
echo "TS SDK: ./libs/sdk-ts/"
echo "PY SDK: ./libs/sdk-py/"
echo "----------------------------------------"
echo "Next: Go to src/kernel/ and start building!"
