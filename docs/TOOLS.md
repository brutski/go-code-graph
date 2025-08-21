# MCP Tools Reference

## Overview

The Go Code Graph MCP Server provides 9 powerful tools that give AI assistants deep understanding of Go codebases through graph database queries. Each tool serves specific analysis needs, from natural language exploration to precise impact analysis.

## Tool Categories

### 🔍 **Query Tools**

- [`natural_query`](#natural_query) - Natural language codebase exploration
- [`cypher_query`](#cypher_query) - Direct graph database access

### 📊 **Analysis Tools**

- [`analyze_impact`](#analyze_impact) - Change impact analysis
- [`find_patterns`](#find_patterns) - Code pattern detection
- [`detect_architecture`](#detect_architecture) - Architectural analysis

### 🔗 **Relationship Tools**

- [`find_implementers`](#find_implementers) - Interface implementation discovery
- [`trace_call_path`](#trace_call_path) - Call chain analysis

### 🏗️ **Workspace Tools**

- [`analyze_workspace`](#analyze_workspace) - Codebase import and analysis
- [`list_workspaces`](#list_workspaces) - List available workspaces

---

## Tool Details

### `natural_query`

**Purpose**: Ask questions about the codebase in natural language, automatically converted to graph queries.

**Use Cases**:

- Exploratory codebase understanding
- Quick fact-finding
- Non-technical stakeholder queries
- Learning about unfamiliar codebases

#### Natural Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `question` | string | ✅ | Natural language question about the codebase |
| `context` | string | ❌ | Additional context for better query generation |
| `workspace` | string | ✅ | Workspace to query |

#### Supported Question Patterns

**Complexity Analysis:**

- "What are the most complex functions?"
- "Which functions have high complexity?"
- "Show me complex methods that need refactoring"

**Structural Queries:**

- "What interfaces are defined?"
- "Show me all packages"
- "Which structs have the most methods?"

**Usage Patterns:**

- "What functions are called the most?"
- "Show me unused functions"
- "Which packages depend on X?"

**Field & Parameter Analysis:**

- "Which structs have the most fields?"
- "Show me functions with many parameters"
- "What are the parameter types for function X?"

**Interface & Implementation:**

- "What interfaces are implemented by structs?"
- "Show me all interface implementations"
- "Which structs implement interface X?"

**Constructor Patterns:**

- "Show me types with multiple constructors"
- "What constructors exist for struct X?"
- "Which types have factory methods?"

#### Examples

```typescript
// Basic complexity query
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "natural_query",
  arguments: {
    question: "What are the most complex functions in the auth package?",
    workspace: "my-project"
  }
})

// Interface exploration
use_mcp_tool({
  server_name: "go-code-graph", 
  tool_name: "natural_query",
  arguments: {
    question: "Show me all interfaces and their implementations",
    context: "I'm trying to understand the abstraction layers",
    workspace: "my-project"
  }
})

// Structural analysis
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "natural_query",
  arguments: {
    question: "Which packages have the most dependencies?",
    workspace: "my-project"
  }
})
```

#### Response Format

- **Summary**: Human-readable explanation of the query
- **Results**: Structured data matching the question intent
- **Query Type**: The generated Cypher pattern used

---

### `cypher_query`

**Purpose**: Execute direct Cypher queries against the code graph database for precise analysis.

**Use Cases**:

- Complex multi-step analysis
- Custom reporting and metrics
- Performance-critical queries
- Advanced graph algorithms

#### Cypher Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | ✅ | Cypher query to execute |
| `parameters` | object | ❌ | Query parameters for parameterized queries |
| `explain` | boolean | ❌ | Return query execution plan instead of results |
| `workspace` | string | ✅ | Workspace to query |

#### Graph Schema Reference

**Node Types:**

```cypher
(:Package {name, path, id})
(:Struct {name, package, id, complexity, embedding})
(:Interface {name, package, id})
(:Function {name, package, id, complexity, signature, embedding})
(:Method {name, package, id, complexity, signature, embedding})
(:Field {name, package, id, type_name})
(:Parameter {name, package, id, type_name, index})
```

**Relationship Types:**

```cypher
()-[:IMPORTS]->()         // Package dependencies
()-[:CALLS]->()          // Function/method calls
()-[:EMBEDS]->()         // Struct embedding
()-[:HAS_FIELD]->()      // Struct fields
()-[:HAS_METHOD]->()     // Type methods
()-[:IMPLEMENTS]->()     // Interface implementation
()-[:CONSTRUCTS]->()     // Type instantiation
()-[:RETURNS]->()        // Return types
```

#### Cypher Query Examples

**Basic Node Queries:**

```typescript
// Count nodes by type
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH (n)
      RETURN n.type as node_type, count(n) as count
      ORDER BY count DESC
    `,
    workspace: "my-project"
  }
})

// Find high-complexity functions
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query", 
  arguments: {
    query: `
      MATCH (f:Function)
      WHERE f.complexity > $minComplexity
      RETURN f.name, f.package, f.complexity
      ORDER BY f.complexity DESC
      LIMIT $limit
    `,
    parameters: {
      "minComplexity": 10,
      "limit": 20
    },
    workspace: "my-project"
  }
})
```

**Relationship Analysis:**

```typescript
// Package dependency analysis
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH (p1:Package)-[:IMPORTS]->(p2:Package)
      WITH p1, count(p2) as dependencies
      RETURN p1.name as package, dependencies
      ORDER BY dependencies DESC
      LIMIT 15
    `,
    workspace: "my-project"
  }
})

// Interface implementation patterns
use_mcp_tool({
  server_name: "go-code-graph", 
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH (i:Interface)<-[:IMPLEMENTS]-(s:Struct)
      WITH i, count(s) as implementers, collect(s.name) as structs
      WHERE implementers > 1
      RETURN i.name as interface, implementers, structs
      ORDER BY implementers DESC
    `,
    workspace: "my-project"
  }
})
```

**Advanced Patterns:**

```typescript
// Find potential god objects
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH (s:Struct)-[:HAS_FIELD]->(f:Field)
      MATCH (s)-[:HAS_METHOD]->(m:Method)
      WITH s, count(f) as fields, count(m) as methods
      WHERE fields > 10 OR methods > 15
      RETURN s.name, s.package, fields, methods,
             (fields + methods) as total_members
      ORDER BY total_members DESC
    `,
    workspace: "my-project"
  }
})

// Circular dependency detection
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH cycle = (p:Package)-[:IMPORTS*2..5]->(p)
      RETURN [n in nodes(cycle) | n.name] as circular_dependency
      LIMIT 10
    `,
    workspace: "my-project"
  }
})
```

---

### `analyze_impact`

**Purpose**: Analyze what would be affected by changing a specific code component.

**Use Cases**:

- Pre-refactoring risk assessment
- Change impact estimation
- Dependency analysis
- Breaking change evaluation

#### Analyze Impact Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `nodeId` | string | ✅ | Node identifier to analyze |
| `changeType` | string | ❌ | Type of change: `signature`, `delete`, `modify` |
| `maxDepth` | number | ❌ | Maximum relationship depth (default: 5) |
| `workspace` | string | ✅ | Workspace to analyze |

#### Change Types

**`signature`** (default): Analyze impact of changing function/method signature

- Finds all callers that would need updates
- Identifies interface contract violations
- Shows parameter usage patterns

**`delete`**: What would break if this component is removed

- Direct dependencies that would fail
- Transitive impact analysis
- Orphaned code identification

**`modify`**: General modification impact

- All related components
- Bidirectional relationships
- Structural dependencies

#### Analyze Impact Examples

```typescript
// Function signature change impact
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_impact",
  arguments: {
    nodeId: "function:internal/auth.ValidateToken",
    changeType: "signature",
    maxDepth: 3,
    workspace: "my-project"
  }
})

// Struct deletion impact
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_impact", 
  arguments: {
    nodeId: "struct:models.User",
    changeType: "delete",
    maxDepth: 4,
    workspace: "my-project"
  }
})

// Interface modification impact
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_impact",
  arguments: {
    nodeId: "interface:storage.Repository",
    changeType: "modify",
    workspace: "my-project"
  }
})
```

#### Analyze Impact Response Format

- **Summary**: Human-readable impact description
- **Affected Components**: List of impacted nodes with relationship types
- **Depth Analysis**: Impact at each relationship level
- **Risk Assessment**: Qualitative risk indicators

---

### `find_patterns`

**Purpose**: Detect code patterns, anti-patterns, and usage statistics across the codebase.

**Use Cases**:

- Code quality assessment
- Technical debt identification
- Refactoring opportunity discovery
- Architectural pattern analysis

#### Find Patterns Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `patternType` | string | ✅ | Pattern type to detect |
| `filter` | object | ❌ | Additional filtering criteria |
| `workspace` | string | ✅ | Workspace to analyze |

#### Supported Pattern Types

**`duplicate`**: Functions with identical names across packages

- Identifies potential code duplication
- Shows distribution patterns
- Helps find consolidation opportunities

**`usage`**: Most frequently used components

- Finds highly coupled functions
- Identifies critical dependencies
- Shows usage hotspots

**`field`**: Structs with excessive fields (>10)

- God object detection
- Data model complexity analysis
- Refactoring candidates

**`parameter`**: Functions with too many parameters (>5)

- Interface complexity indicators
- Potential design issues
- Refactoring opportunities

**`interface`**: Interface usage patterns

- Unused interfaces
- Single-implementation interfaces
- Polymorphism analysis

**`constructor`**: Types with multiple constructors

- Factory pattern identification
- Construction complexity
- Initialization patterns

**`embedding`**: Struct embedding relationships

- Composition over inheritance usage
- Embedding patterns
- Structural relationships

**`orphan`**: Isolated components with no relationships

- Dead code candidates
- Unused components
- Cleanup opportunities

#### Find Patterns Examples

```typescript
// Find duplicate function names
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "duplicate",
    workspace: "my-project"
  }
})

// Find most used functions  
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "usage",
    workspace: "my-project"
  }
})

// Find structs with too many fields
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "field",
    filter: {"min_fields": 15},
    workspace: "my-project"
  }
})

// Find functions with too many parameters
use_mcp_tool({
  server_name: "go-code-graph", 
  tool_name: "find_patterns",
  arguments: {
    patternType: "parameter",
    workspace: "my-project"
  }
})

// Analyze interface usage
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "interface",
    workspace: "my-project"
  }
})
```

#### Find Patterns Response Format

- **Pattern Type**: The detected pattern category
- **Instances**: Specific occurrences with details
- **Statistics**: Quantitative analysis
- **Recommendations**: Suggested actions

---

### `find_implementers`

**Purpose**: Find all types that implement a given interface.

**Use Cases**:

- Interface usage analysis
- Polymorphism understanding
- Implementation discovery
- Contract verification

#### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `interfaceName` | string | ✅ | Full name of the interface (e.g., "io.Writer") |
| `workspace` | string | ✅ | Workspace to search |

#### Find Implementers Examples

```typescript
// Find io.Writer implementations
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_implementers",
  arguments: {
    interfaceName: "io.Writer",
    workspace: "my-project"
  }
})

// Find custom interface implementations
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_implementers",
  arguments: {
    interfaceName: "internal/storage.Repository",
    workspace: "my-project"
  }
})

// Check interface adoption
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_implementers", 
  arguments: {
    interfaceName: "context.Context",
    workspace: "my-project"
  }
})
```

#### Find Implementers Response Format

- **Interface**: The queried interface details
- **Implementers**: List of implementing types
- **Implementation Details**: Package locations and signatures
- **Usage Patterns**: How implementations are used

---

### `trace_call_path`

**Purpose**: Find the call path between two functions or methods.

**Use Cases**:

- Call chain analysis
- Dependency path discovery
- Execution flow understanding
- Debugging assistance

#### Trace Call Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `from` | string | ✅ | Starting function/method name |
| `to` | string | ✅ | Ending function/method name |
| `workspace` | string | ✅ | Workspace to search |

#### Trace Call Path Examples

```typescript
// Find path from main to specific function
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "trace_call_path",
  arguments: {
    from: "main",
    to: "ProcessOrder",
    workspace: "my-project"
  }
})

// Trace initialization flow
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "trace_call_path",
  arguments: {
    from: "init",
    to: "StartServer",
    workspace: "my-project"
  }
})

// Find connection between services
use_mcp_tool({
  server_name: "go-code-graph", 
  tool_name: "trace_call_path",
  arguments: {
    from: "HandleUserRequest",
    to: "SaveToDatabase",
    workspace: "my-project"
  }
})
```

#### Trace Call Path Response Format

- **Call Path**: Sequence of function/method calls
- **Path Length**: Number of hops in the call chain
- **Full Names**: Complete function identifiers
- **Shortest Paths**: Multiple paths if they exist

---

### `detect_architecture`

**Purpose**: Perform structural analysis of the codebase to identify architectural patterns.

**Use Cases**:

- Architecture documentation
- Pattern recognition
- Layer analysis
- Design verification

#### Detect Architecture Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `analysisType` | string | ✅ | Type of architectural analysis |
| `workspace` | string | ✅ | Workspace to analyze |

#### Analysis Types

**`layers`**: Package-based layer analysis

- Shows distribution of components across packages
- Identifies architectural layers
- Reveals package size patterns

**`patterns`**: Architectural pattern detection

- Factory patterns
- Singleton patterns
- Observer patterns
- Strategy patterns

#### Detect Architecture Examples

```typescript
// Analyze architectural layers
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "detect_architecture",
  arguments: {
    analysisType: "layers",
    workspace: "my-project"
  }
})

// Detect design patterns
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "detect_architecture", 
  arguments: {
    analysisType: "patterns",
    workspace: "my-project"
  }
})
```

#### Detect Architecture Response Format

- **Architecture Type**: The analysis performed
- **Layers/Patterns**: Discovered architectural elements
- **Metrics**: Quantitative measurements
- **Observations**: Qualitative insights

---

### `analyze_workspace`

**Purpose**: Import and analyze a new Go workspace/repository into the graph database.

**Use Cases**:

- Initial codebase setup
- Incremental analysis updates
- Multi-repository analysis
- CI/CD integration

#### Analyze Workspace Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workspacePath` | string | ❌ | Path to Go workspace (default: current directory) |
| `workspaceName` | string | ❌ | Unique name for the workspace |
| `incremental` | boolean | ❌ | Perform incremental update vs full refresh |

#### Analyze Workspace Examples

```typescript
// Analyze current directory
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_workspace",
  arguments: {}
})

// Analyze specific project
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_workspace",
  arguments: {
    workspacePath: "/path/to/go/project",
    workspaceName: "my-service",
    incremental: false
  }
})

// Incremental update
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_workspace",
  arguments: {
    workspaceName: "my-service", 
    incremental: true
  }
})
```

#### Analyze Workspace Response Format

- **Analysis Summary**: Node and edge counts
- **Import Status**: Success/failure indicators
- **Performance Metrics**: Processing time and memory usage
- **Error Details**: Any issues encountered

---

### `list_workspaces`

**Purpose**: List all available workspaces in the graph database.

**Use Cases**:

- Discover available workspaces
- Workspace management
- Multi-project analysis setup
- Validation before workspace operations

#### List Workspaces Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| No parameters required | | | |

#### List Workspaces Examples

```typescript
// List all available workspaces
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "list_workspaces",
  arguments: {}
})
```

#### List Workspaces Response Format

- **Workspaces**: Array of workspace names
- **Count**: Total number of workspaces
- **Details**: Additional metadata per workspace

---

## Advanced Usage Patterns

### 1. **Multi-Step Analysis Workflow**

Combine multiple tools for comprehensive analysis:

```typescript
// 1. First, understand the codebase structure
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "natural_query",
  arguments: {
    question: "What are the main packages and their purposes?",
    workspace: "my-project"
  }
})

// 2. Find potential problem areas
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns", 
  arguments: {
    patternType: "usage",
    workspace: "my-project"
  }
})

// 3. Analyze impact of changing high-usage components
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_impact",
  arguments: {
    nodeId: "function:utils.ProcessData",  // From step 2 results
    changeType: "signature",
    workspace: "my-project"
  }
})
```

### 2. **Quality Assessment Pipeline**

Systematic code quality evaluation:

```typescript
// Step 1: Find complex functions
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "usage",
    workspace: "my-project"
  }
})

// Step 2: Check for god objects
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "field",
    workspace: "my-project"
  }
})

// Step 3: Find interface violations
use_mcp_tool({
  server_name: "go-code-graph", 
  tool_name: "find_patterns",
  arguments: {
    patternType: "interface",
    workspace: "my-project"
  }
})

// Step 4: Custom complexity analysis
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH (f:Function)
      WHERE f.complexity > 15
      OPTIONAL MATCH (f)<-[r:CALLS]-()
      WITH f, count(r) as callers
      RETURN f.name, f.package, f.complexity, callers,
             (f.complexity * callers) as risk_score
      ORDER BY risk_score DESC
      LIMIT 10
    `,
    workspace: "my-project"
  }
})
```

### 3. **Refactoring Planning**

Structured approach to refactoring decisions:

```typescript
// 1. Identify refactoring candidates
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "cypher_query",
  arguments: {
    query: `
      MATCH (f:Function)
      WHERE f.complexity > 12
      RETURN f.name, f.package, f.complexity
      ORDER BY f.complexity DESC
      LIMIT 5
    `,
    workspace: "my-project"
  }
})

// 2. For each candidate, analyze impact
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "analyze_impact",
  arguments: {
    nodeId: "function:service.ProcessComplexLogic",
    changeType: "modify",
    maxDepth: 3,
    workspace: "my-project"
  }
})

// 3. Find call paths to understand usage
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "trace_call_path", 
  arguments: {
    from: "main",
    to: "ProcessComplexLogic",
    workspace: "my-project"
  }
})

// 4. Check for similar patterns
use_mcp_tool({
  server_name: "go-code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "duplicate",
    workspace: "my-project"
  }
})
```

## Best Practices

### 1. **Tool Selection Guidelines**

**Start with `natural_query`** for:

- Initial exploration
- Quick questions
- Non-technical stakeholders
- Learning new codebases

**Use `cypher_query`** for:

- Complex multi-step analysis
- Custom metrics and reports
- Performance-critical queries
- Advanced graph algorithms

**Apply `analyze_impact`** before:

- Making significant changes
- Refactoring critical components
- Breaking API changes
- Architecture modifications

### 2. **Query Optimization**

**For Large Codebases:**

- Use `LIMIT` clauses to control result size
- Add `WHERE` filters early in queries
- Use parameterized queries for safety
- Consider using `EXPLAIN` for performance tuning

**Memory Management:**

- Start with targeted queries
- Progressively expand scope
- Use pagination for large result sets
- Monitor Neo4j memory usage

### 3. **Error Handling**

**Common Issues:**

- **Node not found**: Check exact node identifiers with `natural_query`
- **Timeout errors**: Reduce query complexity or add limits
- **Memory errors**: Use more specific filters or pagination
- **Connection errors**: Verify Neo4j is running and accessible

**Debugging Steps:**

1. Test with simple `cypher_query` first
2. Verify node existence: `MATCH (n) WHERE n.label = 'YourFunction' RETURN n`
3. Check relationships: `MATCH (n)-[r]-() WHERE n.label = 'YourFunction' RETURN type(r), count(r)`
4. Use `EXPLAIN` to understand query execution

### 4. **Performance Monitoring**

**Track Query Performance:**

- Response times for different query types
- Memory usage patterns
- Result set sizes
- Error rates

**Optimize Based on Usage:**

- Cache frequently used results
- Pre-compute common metrics
- Use indexes for hot query paths
- Monitor Neo4j query logs

## Integration Examples

### With CI/CD Systems

```yaml
# .github/workflows/code-analysis.yml
- name: Analyze Code Quality
  run: |
    # Use MCP tools in automated fashion
    echo "use_mcp_tool('go-code-graph', 'find_patterns', {'patternType': 'usage'})" | mcp-client
    echo "use_mcp_tool('go-code-graph', 'find_patterns', {'patternType': 'duplicate'})" | mcp-client
```

### With Development IDEs

```python
# IDE plugin example
def analyze_function_impact(function_name):
    return mcp_client.call_tool(
        server="go-code-graph",
        tool="analyze_impact", 
        args={"nodeId": f"function:{function_name}"}
    )
```

### With Documentation Systems

```markdown
<!-- Auto-generated architecture docs -->
{{mcp_tool "go-code-graph" "detect_architecture" analysisType="layers"}}
```

This comprehensive toolset provides everything needed for deep codebase analysis, from quick exploration to detailed architectural assessment. Each tool is designed to work independently or in combination with others for maximum analytical power.
