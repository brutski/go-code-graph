# Test Scripts

This directory contains test and utility scripts for the code-graph MCP server.

## MCP Testing Scripts

### test-mcp-tool.sh

Test individual MCP tools with custom arguments.

**Usage:**

```bash
./scripts/test-mcp-tool.sh <tool_name> [workspace] [arguments]
```

**Examples:**

```bash
# List all workspaces
./scripts/test-mcp-tool.sh list_workspaces

# Query with specific workspace
./scripts/test-mcp-tool.sh natural_query go-code-graph '{"question":"What are the most complex functions?"}'

# Analyze a new workspace
./scripts/test-mcp-tool.sh analyze_workspace my-project '{"workspacePath":"/path/to/code","workspaceName":"my-project"}'
```

### test-mcp-tools.sh

Comprehensive test suite that runs all MCP tools with various test cases.

**Usage:**

```bash
./scripts/test-mcp-tools.sh [workspace]
```

**Example:**

```bash
# Test with default workspace (go-code-graph)
./scripts/test-mcp-tools.sh

# Test with custom workspace
./scripts/test-mcp-tools.sh my-project
```

### setup-test-workspace.sh

Initialize and analyze a workspace for MCP tools testing.

**Usage:**

```bash
./scripts/setup-test-workspace.sh [workspace_name] [workspace_path]
```

**Example:**

```bash
# Setup default workspace
./scripts/setup-test-workspace.sh

# Setup custom workspace
./scripts/setup-test-workspace.sh my-project /path/to/my-project
```

### test-all-tools-working.sh

Legacy test script for verifying all tools are functional.

## Database Scripts

### setup_neo4j_indexes.cypher

Cypher script to create indexes and constraints in Neo4j for optimal graph query performance.

## Prerequisites

- Neo4j running on `localhost:7687` with credentials `neo4j/codeGraph123`
- MCP server binary built at `bin/mcp-server`
- AWS credentials configured for Bedrock (if using embeddings)
- Analyzed workspace data in Neo4j

## Common Issues

1. **No data found**: Run `setup-test-workspace.sh` first to analyze a codebase
2. **Connection refused**: Ensure Neo4j is running (`make run-neo4j`)
3. **Tool errors**: Check that workspace name matches an analyzed workspace
4. **Missing binaries**: Build the MCP server first (`make build-mcp`)
