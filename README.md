# Go Code Graph

A powerful Go codebase analyzer that transforms code into explorable graphs, enabling AI assistants to understand complex codebases through the Model Context Protocol (MCP).

## What is Go Code Graph?

Go Code Graph analyzes Go source code to create a comprehensive graph representation of your codebase structure, including packages, types, functions, and their relationships. This graph can be:

- Visualized in an interactive web interface
- Stored in Neo4j for complex queries
- Queried by AI assistants through MCP for intelligent code understanding

## Key Features

- **🔍 Deep Code Analysis**: Extracts 9 node types and 12 relationship types from Go AST
- **🎨 Interactive Visualization**: Web interface handling 4,000+ nodes with real-time filtering
- **🧠 AI-Powered Understanding**: MCP server enables natural language queries about your code
- **📊 Graph Database**: Neo4j integration for complex architectural analysis
- **🚀 Enterprise Scale**: Successfully tested on codebases with 97 packages and 4,000+ nodes
- **⚡ Semantic Search**: Optional embeddings for enhanced code understanding

## Quick Start with MCP

### 1. Setup with Docker (Recommended)

```bash
# Clone and setup
git clone https://github.com/brutski/go-code-graph.git
cd go-code-graph
make dev-setup  # Starts Neo4j and builds MCP server image
```

### 2. Configure Claude Desktop

```bash
make generate-mcp-config  # Generates exact configuration needed
```

Copy the output to your Claude Desktop config file:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

### 3. Start Using with AI

The MCP server is now ready! Your AI assistant can analyze any Go codebase.

## MCP Tools and Prompts

The MCP server provides two ways to analyze your codebase:

### MCP Prompts (Guided Workflows)

The MCP server includes 15 comprehensive prompts that guide you through common analysis tasks. These prompts show the AI assistant exactly which tools to use and in what order,
ensuring best practices and comprehensive coverage. See [MCP Prompts Guide](docs/MCP_PROMPTS.md) for detailed documentation.

Example prompts:

- `analyze_new_codebase` - Complete initial analysis with architecture overview
- `assess_code_quality` - Comprehensive quality assessment with actionable insights
- `analyze_change_impact` - Understand effects before making changes
- `debug_execution_flow` - Trace execution paths for debugging
- `plan_refactoring` - Get step-by-step refactoring guidance

**When to use prompts**: Starting new analysis tasks, following best practices, ensuring comprehensive coverage.

### MCP Tools (Building Blocks)

The MCP server provides 8 powerful tools for direct codebase analysis. These are the building blocks that prompts use, but you can also use them directly for specific queries and custom exploration:

#### 🔍 **analyze_workspace**

Imports and analyzes a Go codebase into the graph database.

- **When to use**: Initial setup or updating codebase analysis
- **Supports**: Selective inclusion of external dependencies
- **Example usage**:

  ```json
  {
    "workspacePath": "/path/to/project",
    "workspaceName": "my-project",
    "incremental": false,
    "allowedPackages": [
      "github.com/gin-gonic/gin/...",
      "github.com/neo4j/neo4j-go-driver/v5"
    ]
  }
  ```

#### 💬 **natural_query**

Converts natural language questions into graph queries.

- **When to use**: Exploring codebase structure and relationships
- **Example questions**:
  - "What are the most complex functions?"
  - "Which structs have the most fields?"
  - "Show me unused interfaces"
  - "What packages depend on the auth module?"
  - "Which functions handle errors?"

#### 🔗 **cypher_query**

Direct access to the graph database for complex queries.

- **When to use**: Advanced analysis requiring custom graph traversal
- **Example usage**: Finding circular dependencies, analyzing call chains, custom metrics

#### 📊 **analyze_impact**

Analyzes what would be affected by changing a component.

- **When to use**: Before refactoring or making breaking changes
- **Example questions**: "What breaks if I change the ProcessOrder function signature?"

#### 🎯 **find_patterns**

Detects code patterns and anti-patterns across the codebase.

- **When to use**: Code quality assessment, finding technical debt
- **Pattern types**: duplicate functions, high complexity, god objects, unused code

#### 🔌 **find_implementers**

Finds all types implementing a given interface.

- **When to use**: Understanding interface usage and polymorphism
- **Example**: "What implements the io.Writer interface?"

#### 🛤️ **trace_call_path**

Finds execution paths between two functions.

- **When to use**: Understanding how components connect
- **Example**: "How does main() reach the SaveToDatabase function?"

#### 🏗️ **detect_architecture**

Analyzes architectural patterns and layers.

- **When to use**: Understanding system design and structure
- **Analysis types**: layer detection, pattern recognition

**When to use tools directly**: Specific queries, interactive exploration, custom analysis workflows. See [Tools vs Prompts Guide](docs/TOOLS_VS_PROMPTS.md) for detailed comparison.

## Testing and Development

### Test Scripts

The `scripts/` directory contains utility scripts for testing and development:

- **test-mcp-tool.sh**: Test individual MCP tools with custom arguments
- **test-mcp-tools.sh**: Comprehensive test suite for all MCP tools
- **setup-test-workspace.sh**: Initialize and analyze a workspace for testing

Example usage:

```bash
# Test a specific tool
./scripts/test-mcp-tool.sh natural_query go-code-graph '{"question":"What are the most complex functions?"}'

# Run comprehensive test suite
./scripts/test-mcp-tools.sh go-code-graph

# Setup a new workspace
./scripts/setup-test-workspace.sh my-project /path/to/project
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/analyzer -v
```

## Documentation

For detailed information, see the docs directory:

- **[Technical Architecture](docs/ARCHITECTURE.md)**: System design, components, and embeddings
- **[MCP Tools Reference](docs/TOOLS.md)**: Detailed tool documentation with examples
- **[MCP Prompts Guide](docs/MCP_PROMPTS.md)**: Comprehensive prompts for guided analysis
- **[Tools vs Prompts](docs/TOOLS_VS_PROMPTS.md)**: When to use tools vs prompts
- **[Practical Usage Guide](docs/GUIDE.md)**: Common workflows and best practices
- **[Docker MCP Setup](docs/DOCKER_MCP_SETUP.md)**: Detailed Docker deployment guide

## Traditional Usage (Without MCP)

If you prefer to use the tools directly:

```bash
# Build tools
make build

# Analyze codebase (local packages only)
./bin/analyze -repo=. -output=graph.json

# View in browser
./bin/server -graph=graph.json -port=8080
# Open http://localhost:8080/visualization

# Import to Neo4j
./bin/import-neo4j -graph=graph.json -clear
```

## Analyzing External Dependencies

By default, the analyzer only includes packages from your module to keep the graph focused and manageable.
However, you can selectively include external dependencies to understand how your code interacts with third-party libraries.

### Why Include External Packages?

- **API Usage**: See exactly how you're using third-party libraries
- **Integration Points**: Identify all touchpoints with external code
- **Call Chains**: Trace complete execution paths including external calls
- **Interface Implementation**: See which external interfaces your code implements

### Command Line Usage

Use the `-include-packages` flag with the `analyze` command:

```bash
# Include a single external package and all its sub-packages
./bin/analyze -repo=. -output=graph.json \
  -include-packages="github.com/gin-gonic/gin/..."

# Include multiple specific packages
./bin/analyze -repo=. -output=graph.json \
  -include-packages="github.com/gin-gonic/gin,github.com/stretchr/testify/assert"

# Include packages with different patterns
./bin/analyze -repo=. -output=graph.json \
  -include-packages="github.com/sirupsen/logrus,golang.org/x/sync/..."
```

### MCP Usage

When using the MCP server with AI assistants, specify external packages in the `analyze_workspace` tool:

```json
{
  "workspacePath": "/path/to/project",
  "workspaceName": "my-project",
  "incremental": false,
  "allowedPackages": [
    "github.com/gin-gonic/gin/...",
    "github.com/neo4j/neo4j-go-driver/v5"
  ]
}
```

Or simply tell your AI assistant:

- "Analyze /path/to/project including github.com/gin-gonic/gin"
- "Include the database driver packages when analyzing"

### Package Patterns

- **Exact package**: `github.com/sirupsen/logrus` - Only this specific package
- **With sub-packages**: `github.com/gin-gonic/gin/...` - Package and all sub-packages
- **Multiple packages**: Comma-separated list for CLI, array for MCP

### Example Use Cases

#### 1. Database Driver Analysis

```bash
./bin/analyze -repo=. -output=db-integration.json \
  -include-packages="github.com/lib/pq,database/sql"
```

#### 2. Web Framework Integration

```bash
./bin/analyze -repo=. -output=web-app.json \
  -include-packages="github.com/gin-gonic/gin/...,github.com/gorilla/mux"
```

#### 3. Internal Shared Libraries

```bash
./bin/analyze -repo=. -output=full-analysis.json \
  -include-packages="github.com/mycompany/shared-lib/...,github.com/mycompany/common/..."
```

### Best Practices

1. **Start Small**: Begin with your most critical dependencies
2. **Use Patterns**: The `...` suffix includes all sub-packages efficiently
3. **Monitor Size**: External packages can significantly increase graph size
4. **Focus on Interfaces**: Prioritize packages whose interfaces you implement
5. **Vendor Directory**: If using vendoring, you can analyze `./vendor/...` directly

### Performance Considerations

Including external packages increases:

- Analysis time (more packages to parse)
- Memory usage (more nodes and edges)
- Graph complexity (more relationships to track)
- Visualization performance (more elements to render)

### Troubleshooting

**Package Not Found**:

- Ensure the package is in your `go.mod` or run `go mod download`
- Check the exact import path matches what's in your code
- Try using `go list` to verify the package path

**Graph Too Large**:

- Be more selective with packages
- Use specific packages instead of `...` where possible
- Consider analyzing external packages separately

## Example Natural Language Queries

Ask your AI assistant questions like:

**Code Quality**:

- "What are the most complex functions that need refactoring?"
- "Find functions with too many parameters"
- "Show me potential god objects with too many responsibilities"

**Architecture Understanding**:

- "How do the payment and order services interact?"
- "What are the main architectural layers?"
- "Which packages have circular dependencies?"

**Impact Analysis**:

- "What would break if I delete the User struct?"
- "Which services depend on the authentication module?"
- "Find all callers of the deprecated API"

**Code Patterns**:

- "Find duplicate function implementations"
- "Which interfaces have only one implementation?"
- "Show me unused private functions"

**Development Support**:

- "What are the main entry points to understand this service?"
- "Which functions spawn goroutines?"
- "Find all error handling patterns"

**External Dependencies** (when analyzed with -include-packages):

- "How does our code use the wp-golang-packages library?"
- "What external interfaces do we implement?"
- "Show all calls to the database driver"

## Requirements

- **Docker** (for Neo4j and MCP server)
- **Go 1.21+** (for local development)
- **MCP-compatible client** (Claude Desktop, or other MCP clients)

## License

MIT License - see [LICENSE](LICENSE) file for details
