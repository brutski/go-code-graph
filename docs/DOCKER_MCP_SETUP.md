# Docker-based MCP Server Setup

This guide explains how to set up and use the Code Graph MCP Server via Docker for seamless integration with MCP clients like Claude Desktop.

## 🚀 Quick Start

### 1. Development Setup

```bash
# Clone and setup everything in one command
make dev-setup

# This will:
# - Start Neo4j via Docker Compose
# - Build the MCP server Docker image
# - Show configuration instructions
```

### 2. Generate MCP Configuration

```bash
# Get the exact configuration for Claude Desktop
make generate-mcp-config
```

### 3. Test the MCP Server

```bash
# Test with local Neo4j
make run-docker
```

## 🐳 Docker Architecture

### Multi-stage Build with Go Toolchain

The `Dockerfile.mcp` uses a multi-stage build with Go toolchain for full functionality:

- **Builder stage**: Uses Go 1.24.3 to compile the binary
- **Runtime stage**: golang:alpine image with Go toolchain (for analyze_workspace)
- **Security**: Runs as non-root user, includes git and ca-certificates

### Environment Variables

The MCP server supports configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NEO4J_URI` | `bolt://localhost:7687` | Neo4j connection string |
| `NEO4J_USER` | `neo4j` | Neo4j username |
| `NEO4J_PASSWORD` | `password` | Neo4j password |
| `VERBOSE` | `false` | Enable detailed logging |

## 📋 MCP Client Configuration

### Claude Desktop Configuration

Add this to your `claude_desktop_config.json`:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`  
**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "code-graph": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "--network", "host",
        "-e", "NEO4J_URI=bolt://localhost:7687",
        "-e", "NEO4J_USER=neo4j", 
        "-e", "NEO4J_PASSWORD=codeGraph123",
        "-e", "VERBOSE=false",
        "code-graph-mcp:latest"
      ],
      "env": {},
      "description": "Code Graph Analysis Server for enhanced code understanding"
    }
  }
}
```

### Alternative: Using Binary Path

If you prefer to use the compiled binary directly:

```json
{
  "mcpServers": {
    "code-graph": {
      "command": "/path/to/your/project/bin/mcp-server",
      "args": [
        "--neo4j-uri", "bolt://localhost:7687",
        "--neo4j-user", "neo4j",
        "--neo4j-password", "codeGraph123",
        "--verbose"
      ],
      "env": {},
      "description": "Code Graph Analysis Server"
    }
  }
}
```

## 🛠️ Development Workflows

### Local Development

```bash
# Build and run locally (no Docker)
make build-mcp
make run-mcp-local

# Or build and run via Docker
make build-docker
make run-docker
```

### Testing Different Neo4j Instances

```bash
# Local Neo4j (default)
make run-docker

# Remote Neo4j
make run-docker-custom \
  NEO4J_URI=bolt://remote.neo4j.com:7687 \
  NEO4J_USER=myuser \
  NEO4J_PASSWORD=mypassword

# Cloud Neo4j (AuraDB)
make run-docker-custom \
  NEO4J_URI=neo4j+s://xxxxx.databases.neo4j.io \
  NEO4J_USER=neo4j \
  NEO4J_PASSWORD=your-aura-password
```

### Production Deployment Setup

```bash
# Build and push to registry
REGISTRY=myregistry.com make push-docker

# Or use GitHub Container Registry
REGISTRY=ghcr.io/username make push-docker

# Users can then pull and use:
docker pull ghcr.io/username/code-graph-mcp:latest
```

## 🔧 Available Make Commands

### Build Commands

- `make build` - Build all binaries
- `make build-mcp` - Build only MCP server binary
- `make build-docker` - Build Docker image for MCP server

### Run Commands  

- `make run-mcp-local` - Run MCP server locally (for development)
- `make run-docker` - Run MCP server via Docker (with local Neo4j)
- `make run-docker-custom` - Run MCP server via Docker with custom Neo4j
- `make run-neo4j` - Start Neo4j via Docker Compose
- `make stop-neo4j` - Stop Neo4j

### Docker Commands

- `make push-docker` - Push Docker image to registry
- `make pull-docker` - Pull Docker image from registry

### Development

- `make dev-setup` - Set up complete development environment
- `make generate-mcp-config` - Generate MCP configuration for Claude Desktop
- `make test` - Run tests
- `make clean` - Clean build artifacts

## 🚦 Usage Examples

### Basic Workflow

```bash
# 1. Start development environment
make dev-setup

# 2. Test the MCP server
make run-docker

# 3. Get configuration for Claude Desktop
make generate-mcp-config

# 4. Copy the generated JSON to claude_desktop_config.json

# 5. Restart Claude Desktop

# 6. Use the new tools in Claude:
#    - analyze_workspace
#    - natural_query  
#    - analyze_impact
#    - find_patterns
#    - cypher_query
```

### Production Deployment

```bash
# Build and deploy to container registry
REGISTRY=myregistry.com make push-docker

# Update MCP client configuration to use registry image:
# "command": "docker",
# "args": ["run", "-i", "--rm", "myregistry.com/code-graph-mcp:latest"]
```

### Cloud Neo4j (AuraDB) Integration

```bash
# Run with Neo4j AuraDB
make run-docker-custom \
  NEO4J_URI="neo4j+s://xxxxx.databases.neo4j.io" \
  NEO4J_USER="neo4j" \
  NEO4J_PASSWORD="your-aura-password"
```

## 🔍 Available MCP Tools

Once configured, you'll have access to these tools in your MCP client:

### 1. `analyze_workspace`

Analyzes a workspace/repository and imports it to the graph database.

```json
{
  "workspacePath": "./",
  "workspaceName": "my-project", 
  "incremental": false
}
```

### 2. `natural_query`

Ask questions about the codebase in natural language.

```json
{
  "question": "What are the most complex functions?",
  "context": "Looking for refactoring candidates"
}
```

### 3. `analyze_impact`

Analyze the impact of changing a specific code component.

```json
{
  "nodeId": "function:pkg.FunctionName",
  "changeType": "signature",
  "maxDepth": 5
}
```

### 4. `find_patterns`

Find code patterns, duplicates, or anti-patterns.

```json
{
  "patternType": "duplicate",
  "filter": {}
}
```

### 5. `cypher_query`

Execute Cypher queries directly against the code graph.

```json
{
  "query": "MATCH (f:CodeNode) WHERE f.complexity > 10 RETURN f.label, f.complexity",
  "parameters": {},
  "explain": false
}
```

## 🐛 Troubleshooting

### MCP Server Won't Start

```bash
# Check Docker image exists
docker images | grep code-graph-mcp

# Rebuild if missing
make build-docker

# Test Neo4j connection
docker run --rm --network host \
  -e NEO4J_URI=bolt://localhost:7687 \
  -e NEO4J_USER=neo4j \
  -e NEO4J_PASSWORD=codeGraph123 \
  -e VERBOSE=true \
  code-graph-mcp:latest
```

### Neo4j Connection Issues

```bash
# Check Neo4j is running
docker-compose ps

# Start Neo4j if needed
make run-neo4j

# Check credentials in docker-compose.yml
grep -A 5 NEO4J_AUTH docker-compose.yml
```

### Claude Desktop Not Recognizing Server

1. Verify JSON syntax in `claude_desktop_config.json`
2. Completely quit and restart Claude Desktop
3. Check that Docker image is available: `docker images`
4. Test server manually: `make run-docker`

## 🔒 Security Considerations

### Production Security

- Use environment variables for sensitive credentials
- Consider using Docker secrets for Neo4j passwords
- Run containers with read-only root filesystem when possible
- Use specific image tags instead of `latest` in production

### Network Security

- The `--network host` flag is used for local development
- In production, consider using custom Docker networks
- Restrict Neo4j access to only necessary services

## 📈 Performance Tuning

### Docker Resource Limits

```bash
# Run with resource limits
docker run -i --rm \
  --network host \
  --memory="1g" \
  --cpus="2" \
  -e NEO4J_URI=bolt://localhost:7687 \
  code-graph-mcp:latest
```

### Neo4j Performance

- Adjust `NEO4J_MEMORY_HEAP_INITIAL_SIZE` and `NEO4J_MEMORY_HEAP_MAX_SIZE` in docker-compose.yml
- Consider using Neo4j Enterprise for production workloads
- Configure appropriate Neo4j indexes for your query patterns
