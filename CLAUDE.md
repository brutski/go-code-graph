# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Build
```bash
# Build all binaries
make build

# Build only specific binaries
make build-analyze      # Build the analyzer CLI
make build-mcp         # Build the MCP server
```

### Test
```bash
# Run all tests
make test
# or
go test ./...
```

### Lint
```bash
# Run linting
make lint

# Run linting with auto-fix
make lint-fix
```

### Development Setup
```bash
# Full development setup (starts Neo4j and builds Docker images)
make dev-setup

# Start Neo4j database
make run-neo4j

# Stop Neo4j database
make stop-neo4j
```

### Running the Tools
```bash
# Analyze a codebase
./bin/analyze -repo=/path/to/project -output=graph.json -verbose

# Start the web visualization server
./bin/server -graph=graph.json -port=8080

# Import to Neo4j
./bin/import-neo4j -graph=graph.json -clear

# Run MCP server locally
make run-mcp-local
```

## High-Level Architecture

This is a Go codebase analysis and visualization system with four main components:

### 1. **Analyzer** (`internal/analyzer/`)
- Parses Go source code using AST analysis
- Builds a graph representation of code structure
- Calculates complexity metrics
- Supports embeddings for semantic analysis

### 2. **Graph Storage**
- **JSON Export**: Serializes graph data for web visualization
- **Neo4j Database**: Stores graph for complex queries and AI analysis
- Supports 9 node types (Package, Struct, Interface, Function, Method, Field, Parameter, Constant, Variable)
- Supports 12+ relationship types (calls, imports, embeds, implements, etc.)

### 3. **Web Visualization** (`internal/server/`, `web/`)
- Interactive web interface using Cytoscape.js
- Real-time search and filtering
- Handles enterprise-scale codebases (4,000+ nodes)
- Responsive design with performance optimizations

### 4. **MCP Server** (`internal/mcpserver/`)
- Provides AI assistants with deep codebase understanding
- Natural language to Cypher query conversion
- Impact analysis and pattern detection tools
- Integrates with Claude Desktop and other MCP clients

## Key Design Patterns

- **Provider-agnostic embeddings**: Factory pattern for OpenAI/Bedrock/Mock providers
- **Batch processing**: Efficient handling of large codebases
- **Graph-based architecture**: All relationships modeled as graph edges
- **Microservice-ready**: Each component can run independently

## Important Considerations

- The analyzer CLI (`cmd/analyze`) does NOT generate embeddings by default (performance)
- The MCP server requires Neo4j to be running with imported graph data
- Default Neo4j credentials: `neo4j/codeGraph123`
- The web server serves static files from the `web/` directory
- All binaries are built with CGO disabled for portability