#!/bin/bash

# PushIQ Quick Start Script
# This script starts PostgreSQL, backend, and frontend with one command

set -e

PROJECT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BACKEND_DIR="$PROJECT_DIR/backend"
FRONTEND_DIR="$PROJECT_DIR/frontend"

echo "🚀 PushIQ Quick Start"
echo "===================="
echo ""

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
    exit 1
fi

echo "✅ Docker found"

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Please install Go 1.22+ from https://golang.org/dl/"
    exit 1
fi

echo "✅ Go found"

# Check Node
if ! command -v node &> /dev/null; then
    echo "❌ Node.js not found. Please install Node.js 16+ from https://nodejs.org/"
    exit 1
fi

echo "✅ Node.js found"
echo ""

# Start PostgreSQL
echo "📦 Starting PostgreSQL..."
cd "$PROJECT_DIR"
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker exec pushiq-postgres pg_isready -U postgres &> /dev/null; then
        echo "✅ PostgreSQL is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "❌ PostgreSQL failed to start"
        exit 1
    fi
    sleep 1
done

echo ""

# Build backend
echo "🔨 Building backend..."
cd "$BACKEND_DIR"
go build -o pushiq-api ./cmd/pushiq-api
echo "✅ Backend built"

echo ""
echo "===================="
echo "✅ Setup Complete!"
echo "===================="
echo ""
echo "📍 Next steps:"
echo ""
echo "1️⃣  Start Backend (Terminal 1):"
echo "   cd $BACKEND_DIR"
echo "   export \$(cat .env | xargs)"
echo "   ./pushiq-api"
echo ""
echo "2️⃣  Start Frontend (Terminal 2):"
echo "   cd $FRONTEND_DIR"
echo "   npm install"
echo "   npm run dev"
echo ""
echo "3️⃣  Open Dashboard:"
echo "   http://localhost:3000"
echo "   API Key: test-key-12345"
echo ""
echo "📚 Documentation:"
echo "   Read DOCKER_SETUP.md for detailed instructions"
echo ""
