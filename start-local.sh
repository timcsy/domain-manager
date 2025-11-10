#!/bin/bash

# Kubernetes Domain Manager - Local Development Startup Script

set -e

echo "🚀 Starting Kubernetes Domain Manager (Local Development Mode)"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed"
    echo "   Please install Go 1.22+ from https://go.dev/dl/"
    exit 1
fi

# Navigate to backend directory
cd backend

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env from .env.local..."
    cp .env.local .env
else
    echo "✅ .env already exists"
fi

# Create data directory
mkdir -p data

# Remove old database to ensure clean migrations
if [ -f data/database.db ]; then
    echo "🗑️  Removing old database for clean start..."
    rm -f data/database.db
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Build frontend CSS (if npm is available)
if command -v npm &> /dev/null; then
    echo "🎨 Building frontend CSS..."
    cd ../frontend
    if [ ! -d node_modules ]; then
        npm install
    fi
    npm run build:css
    cd ../backend
else
    echo "⚠️  npm not found - skipping frontend build"
    echo "   Install Node.js to build TailwindCSS"
fi

echo ""
echo "✅ Setup complete!"
echo ""
echo "🔧 Running in MOCK MODE (no Kubernetes required)"
echo "   - Database: ./backend/data/database.db"
echo "   - All K8s operations will be simulated"
echo ""
echo "📍 Starting server on http://localhost:8080"
echo "   Username: admin"
echo "   Password: admin"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Set environment variables for mock mode
export K8S_MOCK=true
export K8S_IN_CLUSTER=false

# Run the application
go run src/main.go
