# Fantasy Draft Board - Common Commands

# Default: show available commands
default:
    @just --list

# Build the server binary
build:
    @echo "Building draft-board..."
    go build -o draft-board ./cmd/server/main.go

# Run the server
run:
    @echo "Starting server on http://localhost:8080..."
    go run ./cmd/server/main.go

# Run the server on a specific port (usage: just run-port 3000)
run-port port:
    @echo "Starting server on http://localhost:{{port}}..."
    PORT={{port}} go run ./cmd/server/main.go

# Build and run the server
start: build
    @echo "Starting server..."
    ./draft-board

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    rm -f draft-board
    rm -f main
    rm -f draft-board.db

# Run tests
test:
    @echo "Running tests..."
    go test ./...

# Run tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    go test -cover ./...

# Run tests with verbose output
test-verbose:
    @echo "Running tests (verbose)..."
    go test -v ./...

# Format code
fmt:
    @echo "Formatting code..."
    go fmt ./...

# Run linter
lint:
    @echo "Running linter..."
    golangci-lint run ./...

# Install dependencies
deps:
    @echo "Downloading dependencies..."
    go mod download
    go mod tidy

# Update dependencies
deps-update:
    @echo "Updating dependencies..."
    go get -u ./...
    go mod tidy

# Show database info
db-info:
    @echo "Database: draft-board.db"
    @if [ -f draft-board.db ]; then \
        echo "Database exists"; \
        sqlite3 draft-board.db ".tables"; \
    else \
        echo "Database does not exist yet"; \
    fi

# Open database shell
db-shell:
    @echo "Opening SQLite shell..."
    sqlite3 draft-board.db

# Remove database (reset)
db-reset:
    @echo "Removing database..."
    rm -f draft-board.db
    @echo "Database reset. Will be recreated on next run."

# Run with auto-reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
    @echo "Starting development server with auto-reload..."
    @if command -v air > /dev/null; then \
        air; \
    else \
        echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
        echo "Falling back to regular run..."; \
        go run ./cmd/server/main.go; \
    fi

# Check for common issues
check:
    @echo "Checking for common issues..."
    @echo "1. Checking Go version..."
    @go version
    @echo "\n2. Checking dependencies..."
    @go mod verify
    @echo "\n3. Checking for syntax errors..."
    @go build ./... > /dev/null && echo "✓ No syntax errors"
    @echo "\n4. Checking database..."
    @if [ -f draft-board.db ]; then echo "✓ Database exists"; else echo "✗ Database does not exist (will be created on first run)"; fi

# Show project structure
tree:
    @echo "Project structure:"
    @tree -I 'node_modules|.git|draft-board.db' -L 3 || find . -type f -name "*.go" | head -20

# Generate templ files (if using templ)
templ-generate:
    @echo "Generating templ files..."
    @if command -v templ > /dev/null; then \
        templ generate; \
    else \
        echo "Templ not installed. Install with: go install github.com/a-h/templ/cmd/templ@latest"; \
    fi

# Run all checks before commit
pre-commit: fmt test check
    @echo "\n✓ All checks passed!"

# Seed player data from CSV (usage: just seed players.csv)
seed file:
    @echo "Seeding player data from {{file}}..."
    go run ./cmd/seed/main.go -file {{file}}

# Seed with sample data
seed-sample:
    @echo "Seeding with sample player data..."
    go run ./cmd/seed/main.go -file data/sample-players.csv

# Build Docker image with Ko
ko-build:
    @echo "Building Docker image with Ko..."
    @if command -v ko > /dev/null; then \
        ko build --local github.com/vibes/draft-board/cmd/server; \
    else \
        echo "Ko not installed. Install with: go install github.com/google/ko@latest"; \
    fi

# Build and publish Docker image with Ko
ko-publish registry:
    @echo "Building and publishing Docker image with Ko..."
    @if command -v ko > /dev/null; then \
        ko publish --push=true {{registry}}/github.com/vibes/draft-board/cmd/server; \
    else \
        echo "Ko not installed. Install with: go install github.com/google/ko@latest"; \
    fi

# Build Docker image with Ko (local only, no push)
# Note: This uses Docker for cross-compilation (required for CGO)
ko-build-local:
    @echo "Building Docker image with Ko (local only)..."
    @if command -v ko > /dev/null && command -v docker > /dev/null; then \
        KO_DOCKER_REPO=ko.local ko build --push=false github.com/vibes/draft-board/cmd/server; \
    else \
        if ! command -v ko > /dev/null; then \
            echo "Ko not installed. Install with: go install github.com/google/ko@latest"; \
        fi; \
        if ! command -v docker > /dev/null; then \
            echo "Docker not installed or not running. Docker is required for CGO cross-compilation."; \
        fi; \
    fi

# Run container built with Ko (requires docker)
ko-run:
    @echo "Running container built with Ko..."
    @if command -v ko > /dev/null; then \
        IMAGE=$$(ko build --local github.com/vibes/draft-board/cmd/server 2>/dev/null) && \
        docker run --rm -p 8080:8080 \
            -e PORT=8080 \
            -e DB_PATH=/app/data/draft-board.db \
            -v draft-board-data:/app/data \
            $$IMAGE; \
    else \
        echo "Ko not installed. Install with: go install github.com/google/ko@latest"; \
    fi

