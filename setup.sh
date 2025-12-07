#!/bin/bash

# GophKeeper Setup Script
# This script helps you set up and run GophKeeper quickly

set -e

echo "🔐 GophKeeper Setup Script"
echo "=========================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.24 or higher."
    exit 1
fi
echo "✅ Go $(go version | awk '{print $3}') found"

# Check PostgreSQL (optional)
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL found"
else
    echo "⚠️  PostgreSQL not found in PATH (you can use Docker instead)"
fi

# Check Docker (optional)
if command -v docker &> /dev/null; then
    echo "✅ Docker found"
else
    echo "⚠️  Docker not found (optional for deployment)"
fi

echo ""
echo "Setting up GophKeeper..."
echo ""

# Install dependencies
echo "📦 Installing Go dependencies..."
go mod download
go mod tidy

# Create bin directory
mkdir -p bin

# Build server and client
echo "🔨 Building server..."
go build -ldflags "-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")'" -o bin/server cmd/server/main.go

echo "🔨 Building client..."
go build -ldflags "-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ")'" -o bin/client cmd/client/main.go

echo ""
echo "${GREEN}✅ Build successful!${NC}"
echo ""

# Offer to start services
echo "Select deployment option:"
echo "1) Docker Compose (recommended for testing)"
echo "2) Manual setup (requires PostgreSQL)"
echo "3) Skip for now"
read -p "Enter choice (1-3): " choice

case $choice in
    1)
        echo ""
        echo "Starting services with Docker Compose..."
        if command -v docker-compose &> /dev/null; then
            docker-compose up -d
            echo ""
            echo "${GREEN}✅ Services started!${NC}"
            echo "Server is running on localhost:50051"
        else
            echo "❌ docker-compose not found. Please install Docker Compose."
            exit 1
        fi
        ;;
    2)
        echo ""
        echo "Manual setup selected."
        echo ""
        echo "Please ensure PostgreSQL is running and create a database:"
        echo "  CREATE DATABASE gophkeeper;"
        echo ""
        echo "Set environment variables:"
        echo "  export DB_DSN='postgres://user:password@localhost:5432/gophkeeper?sslmode=disable'"
        echo "  export JWT_SECRET='your-secret-key'"
        echo ""
        echo "Then start the server:"
        echo "  ./bin/server"
        ;;
    3)
        echo "Skipping service startup."
        ;;
    *)
        echo "Invalid choice. Skipping service startup."
        ;;
esac

echo ""
echo "📖 Quick Start Guide:"
echo "====================="
echo ""
echo "1. Register a new user:"
echo "   ${YELLOW}./bin/client register -u alice@example.com${NC}"
echo ""
echo "2. Login:"
echo "   ${YELLOW}./bin/client login -u alice@example.com${NC}"
echo ""
echo "3. Add a credential:"
echo "   ${YELLOW}./bin/client add credential -n GitHub -l myuser -p mypass${NC}"
echo ""
echo "4. List all items:"
echo "   ${YELLOW}./bin/client list${NC}"
echo ""
echo "5. Get specific item:"
echo "   ${YELLOW}./bin/client get -n GitHub${NC}"
echo ""
echo "For more information, see README.md and PROJECT_SUMMARY.md"
echo ""
echo "${GREEN}Setup complete! 🎉${NC}"

