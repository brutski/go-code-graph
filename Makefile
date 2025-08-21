# Go Code Graph Makefile
.PHONY: build clean test run-server run-neo4j stop-neo4j build-docker run-docker push-docker help

# Variables
BINARY_DIR = bin
DOCKER_IMAGE = go-code-graph-mcp
DOCKER_TAG = latest

# Go build flags
GO_BUILD_FLAGS = -ldflags="-s -w"
GO_BUILD_ENV = CGO_ENABLED=0

# Default target
all: build

# Build all binaries
build:
	@echo "📦 Building all binaries..."
	mkdir -p $(BINARY_DIR)
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/analyze ./cmd/analyze
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/import-neo4j ./cmd/import-neo4j
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/server ./cmd/server
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/mcp-server ./cmd/mcp-server
	@echo "✅ All binaries built successfully"

# Build only the MCP server
build-mcp:
	@echo "📦 Building MCP server..."
	mkdir -p $(BINARY_DIR)
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/mcp-server ./cmd/mcp-server
	@echo "✅ MCP server built successfully"

# Build only the analyze binary
build-analyze:
	@echo "📦 Building analyze binary..."
	mkdir -p $(BINARY_DIR)
	$(GO_BUILD_ENV) go build $(GO_BUILD_FLAGS) -o $(BINARY_DIR)/analyze ./cmd/analyze
	@echo "✅ Analyze binary built successfully"
	@echo "💡 Usage: ./$(BINARY_DIR)/analyze -repo=/path/to/project"


# Build Docker image for MCP server
build-docker:
	@echo "🐳 Building Docker image for MCP server..."
	docker build -f Dockerfile.mcp -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "✅ Docker image $(DOCKER_IMAGE):$(DOCKER_TAG) built successfully"

# Build and tag for registry (requires REGISTRY env var)
build-docker-registry: build-docker
	@if [ -z "$(REGISTRY)" ]; then echo "❌ REGISTRY environment variable required"; exit 1; fi
	@echo "🏷️  Tagging for registry..."
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Tagged as $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run MCP server via Docker (local Neo4j)
run-docker:
	@echo "🚀 Running MCP server via Docker..."
	docker run -i --rm \
		--network host \
		-e NEO4J_URI=bolt://localhost:7687 \
		-e NEO4J_USER=neo4j \
		-e NEO4J_PASSWORD=codeGraph123 \
		-e VERBOSE=true \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Run MCP server via Docker with workspace volume mounting
run-docker-with-workspace:
	@echo "🚀 Running MCP server via Docker with workspace access..."
	@echo "💡 This allows analyze_workspace to access your local filesystem"
	docker run -i --rm \
		--network host \
		-v "$(PWD):/workspace:ro" \
		-e NEO4J_URI=bolt://localhost:7687 \
		-e NEO4J_USER=neo4j \
		-e NEO4J_PASSWORD=codeGraph123 \
		-e VERBOSE=true \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Run MCP server via Docker with custom Neo4j
run-docker-custom:
	@echo "🚀 Running MCP server via Docker with custom Neo4j..."
	@echo "Usage: make run-docker-custom NEO4J_URI=bolt://remote:7687 NEO4J_USER=user NEO4J_PASSWORD=pass"
	docker run -i --rm \
		-e NEO4J_URI=$(or $(NEO4J_URI),bolt://localhost:7687) \
		-e NEO4J_USER=$(or $(NEO4J_USER),neo4j) \
		-e NEO4J_PASSWORD=$(or $(NEO4J_PASSWORD),password) \
		-e VERBOSE=$(or $(VERBOSE),true) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

# Run MCP server locally (for development)
run-mcp-local: build-mcp
	@echo "🚀 Running MCP server locally..."
	./$(BINARY_DIR)/mcp-server --verbose

# Start Neo4j via Docker Compose
run-neo4j:
	@echo "🗄️  Starting Neo4j..."
	docker-compose up -d
	@echo "✅ Neo4j started at http://localhost:7474"
	@echo "📊 Default credentials: neo4j/codeGraph123"

# Stop Neo4j
stop-neo4j:
	@echo "🛑 Stopping Neo4j..."
	docker-compose down
	@echo "✅ Neo4j stopped"

# Push Docker image to registry
push-docker: build-docker-registry
	@echo "📤 Pushing Docker image to registry..."
	docker push $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Image pushed to $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"

# Pull Docker image from registry
pull-docker:
	@echo "📥 Pulling Docker image from registry..."
	docker pull $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker tag $(REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "✅ Image pulled and tagged as $(DOCKER_IMAGE):$(DOCKER_TAG)"

# Run tests
test:
	@echo "🧪 Running tests..."
	go test ./...

# Lint the codebase
lint:
	@echo "🔍 Running golangci-lint..."
	golangci-lint run -v --timeout 3m0s
	@echo "✅ Linting complete"

# Lint and auto-fix issues where possible
lint-fix:
	@echo "🔧 Running golangci-lint with auto-fix..."
	golangci-lint run --fix -v --timeout 3m0s
	@echo "✅ Linting and auto-fix complete"


# Clean build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(BINARY_DIR)
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@echo "✅ Cleanup complete"

# Development setup (start everything needed for development)
dev-setup: run-neo4j build-docker
	@echo "🔧 Development environment ready!"
	@echo ""
	@echo "📋 Next steps:"
	@echo "  1. Test MCP server: make run-docker"
	@echo "  2. Or run locally: make run-mcp-local"
	@echo "  3. Configure your MCP client (Claude Desktop) with:"
	@echo ""
	@echo '    "go-code-graph": {'
	@echo '      "command": "docker",'
	@echo '      "args": ['
	@echo '        "run", "-i", "--rm", "--network", "host",'
	@echo '        "-e", "NEO4J_URI=bolt://localhost:7687",'
	@echo '        "-e", "NEO4J_USER=neo4j",'
	@echo '        "-e", "NEO4J_PASSWORD=codeGraph123",'
	@echo '        "-e", "VERBOSE=true",'
	@echo '        "$(DOCKER_IMAGE):$(DOCKER_TAG)"'
	@echo '      ]'
	@echo '    }'

# Generate MCP configuration for Claude Desktop
generate-mcp-config:
	@echo "📄 MCP Configuration for Claude Desktop:"
	@echo ""
	@echo "🔹 RECOMMENDED: LOCAL BINARY (full filesystem access, no Docker complexity):"
	@echo "{"
	@echo '  "mcpServers": {'
	@echo '    "go-code-graph": {'
	@echo '      "command": "$(PWD)/$(BINARY_DIR)/mcp-server",'
	@echo '      "env": {'
	@echo '        "NEO4J_URI": "bolt://localhost:7687",'
	@echo '        "NEO4J_USER": "neo4j",'
	@echo '        "NEO4J_PASSWORD": "codeGraph123",'
	@echo '        "VERBOSE": "false"'
	@echo '      },'
	@echo '      "description": "Code Graph Analysis Server with full workspace access"'
	@echo '    }'
	@echo '  }'
	@echo "}"
	@echo ""
	@echo "🔹 DOCKER QUERIES ONLY (no analyze_workspace):"
	@echo "{"
	@echo '  "mcpServers": {'
	@echo '    "go-code-graph": {'
	@echo '      "command": "docker",'
	@echo '      "args": ['
	@echo '        "run", "-i", "--rm",'
	@echo '        "--network", "host",'
	@echo '        "-e", "NEO4J_URI=bolt://localhost:7687",'
	@echo '        "-e", "NEO4J_USER=neo4j",'
	@echo '        "-e", "NEO4J_PASSWORD=codeGraph123",'
	@echo '        "-e", "VERBOSE=false",'
	@echo '        "$(DOCKER_IMAGE):$(DOCKER_TAG)"'
	@echo '      ],'
	@echo '      "description": "Code Graph Query Server (Docker-based)"'
	@echo '    }'
	@echo '  }'
	@echo "}"
	@echo ""
	@echo "💡 USAGE:"
	@echo "• Local binary: use any absolute path for analyze_workspace"
	@echo "• Docker queries: perfect for cypher_query, natural_query, etc."
	@echo "• Run 'make build' to ensure local binary exists"

# Show help
help:
	@echo "🛠️  Go Code Graph - Available Commands"
	@echo ""
	@echo "📦 Build Commands:"
	@echo "  build              Build all binaries (analyze, import-neo4j, server, mcp-server)"
	@echo "  build-mcp          Build only MCP server binary"
	@echo "  build-analyze      Build only analyze binary"
	@echo "  build-docker       Build Docker image for MCP server"
	@echo ""
	@echo "🚀 Run Commands:"
	@echo "  run-mcp-local      Run MCP server locally (for development)"
	@echo "  run-docker         Run MCP server via Docker (with local Neo4j)"
	@echo "  run-docker-custom  Run MCP server via Docker with custom Neo4j"
	@echo "  run-neo4j          Start Neo4j via Docker Compose"
	@echo "  stop-neo4j         Stop Neo4j"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  push-docker        Push Docker image to registry"
	@echo "  pull-docker        Pull Docker image from registry"
	@echo ""
	@echo "🔧 Development:"
	@echo "  dev-setup          Set up complete development environment"
	@echo "  generate-mcp-config Generate MCP configuration for Claude Desktop"
	@echo "  test               Run tests"
	@echo "  lint               Run golangci-lint"
	@echo "  lint-fix           Run golangci-lint with auto-fix"
	@echo "  clean              Clean build artifacts"
	@echo ""
	@echo "📖 Usage Examples:"
	@echo "  make dev-setup                    # Full development setup"
	@echo "  make run-docker                   # Test MCP server"
	@echo "  make generate-mcp-config          # Get Claude Desktop config"
	@echo "  REGISTRY=myregistry.com make push-docker  # Push to custom registry"
