.PHONY: help build test test-fast test-html clean install-poetry setup-poetry run-web run-mcp swagger web

# Variables
BINARY_NAME=lazy-ai-coder
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
PORT=8888

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)
	cd web && npm run build
	@echo "✅ Build complete!"

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if command -v swag &> /dev/null; then \
		swag init -g main.go --output docs; \
		echo "✅ Swagger docs generated at docs/"; \
		echo "   View at: http://localhost:$(PORT)/swagger/index.html"; \
	else \
		echo "⚠️  swag not installed. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g main.go --output docs; \
		echo "✅ Swagger docs generated!"; \
	fi

build-with-swagger: swagger build ## Generate swagger docs and build

test: ## Run all tests (requires Poetry)
	@echo "Running tests..."
	@if command -v poetry &> /dev/null; then \
		poetry run pytest tests/test_mcp_http.py -v; \
	else \
		pytest tests/test_mcp_http.py -v; \
	fi

test-fast: ## Run fast tests (skip external dependencies)
	@./run-tests.sh --fast

test-html: ## Run tests and generate HTML report
	@./run-tests.sh --html

test-coverage: ## Run tests with coverage
	@./run-tests.sh --coverage

install-poetry: ## Install Poetry package manager
	@echo "Installing Poetry..."
	@curl -sSL https://install.python-poetry.org | python3 -
	@echo "✅ Poetry installed!"

setup-poetry: ## Install Python dependencies with Poetry
	@echo "Installing Python dependencies..."
	@if command -v poetry &> /dev/null; then \
		poetry install --only test; \
	else \
		echo "❌ Poetry not found. Run: make install-poetry"; \
		exit 1; \
	fi
	@echo "✅ Dependencies installed!"

setup-pip: ## Install Python dependencies with pip
	@echo "Installing Python dependencies with pip..."
	pip3 install -r tests/requirements.txt
	@echo "✅ Dependencies installed!"

run-web: build ## Build and run web server with MCP
	@echo "Starting web server on port $(PORT)..."
	./$(BINARY_NAME) web -p $(PORT)

run-mcp: build ## Build and run stdio MCP server
	@echo "Starting stdio MCP server..."
	./$(BINARY_NAME) mcp

import-projects: build ## Import projects from config.yaml to database
	@echo "Importing projects from config.yaml..."
	./$(BINARY_NAME) import projects

clean: ## Clean build artifacts and test outputs
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf htmlcov/
	rm -f .coverage
	rm -f report.html
	rm -rf .pytest_cache/
	rm -rf assets/
	@echo "✅ Clean complete!"

fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "✅ Format complete!"

lint: ## Lint Go code
	@echo "Linting Go code..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run; \
	else \
		go vet ./...; \
	fi

web: ## Build web application
	@echo "Building web..."
	cd web && npm run build
	@echo "✅ Web build complete!"

shutdown:
	pkill lazy-ai-coder

# Development workflows
dev: build-with-swagger web run-web ## Generate swagger, build and run for development

test-all: build test-fast ## Build and run all tests

quick-test: ## Quick test (build, start server, run tests, stop server)
	@echo "Running quick test..."
	@make build
	@./$(BINARY_NAME) web -p $(PORT) & echo $$! > /tmp/mcp-server.pid
	@sleep 2
	@make test-fast || (kill `cat /tmp/mcp-server.pid` && rm /tmp/mcp-server.pid && exit 1)
	@kill `cat /tmp/mcp-server.pid` && rm /tmp/mcp-server.pid
	@echo "✅ Quick test complete!"

# CI/CD targets
ci-test: ## CI/CD test target
	@echo "Running CI tests..."
	@make build
	@./$(BINARY_NAME) web -p $(PORT) & echo $$! > /tmp/mcp-server.pid
	@sleep 2
	@./run-tests.sh --fast || (kill `cat /tmp/mcp-server.pid` && rm /tmp/mcp-server.pid && exit 1)
	@kill `cat /tmp/mcp-server.pid` && rm /tmp/mcp-server.pid

# Show info
info: ## Show project information
	@echo "Binary: $(BINARY_NAME)"
	@echo "Port: $(PORT)"
	@echo "Go version: $$(go version)"
	@if command -v poetry &> /dev/null; then \
		echo "Poetry version: $$(poetry --version)"; \
	else \
		echo "Poetry: not installed"; \
	fi
	@if command -v pytest &> /dev/null; then \
		echo "Pytest version: $$(pytest --version)"; \
	else \
		echo "Pytest: not installed"; \
	fi
	@if command -v swag &> /dev/null; then \
		echo "Swag version: $$(swag --version)"; \
	else \
		echo "Swag: not installed (run: make swagger to install)"; \
	fi

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t lazy-ai-coder:latest .
	@echo "✅ Docker image built!"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p $(PORT):$(PORT) lazy-ai-coder:latest

docker-compose-up: ## Start all services with docker-compose
	@echo "Starting services with docker-compose..."
	cd deploy && docker-compose up -d
	@echo "✅ Services started!"

docker-compose-down: ## Stop all services
	@echo "Stopping services..."
	cd deploy && docker-compose down
	@echo "✅ Services stopped!"

docker-compose-logs: ## View docker-compose logs
	cd deploy && docker-compose logs -f
