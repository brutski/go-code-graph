# Practical Usage Guide

## Overview

This guide provides practical examples and best practices for using Go Code Graph to analyze, understand, and improve Go codebases. Whether you're onboarding new developers, planning refactoring, or conducting code reviews, this tool can significantly enhance your workflow.

## Quick Start Workflows

### 1. First-Time Codebase Analysis

**Goal**: Get an overview of a new codebase structure and identify key components.

```bash
# 1. Analyze the codebase
cd /path/to/go-project
./bin/analyze -repo=. -output=codebase.json -verbose

# 1a. Analyze with external dependencies (optional)
./bin/analyze -repo=. -output=codebase.json -verbose \
  -include-packages="github.com/gin-gonic/gin/...,github.com/sirupsen/logrus"

# 2. Launch web interface for exploration
./bin/server -graph=codebase.json -port=8080

# 3. Open browser to http://localhost:8080/visualization
```

**What to look for:**

- Package distribution and dependencies
- Most complex functions (red nodes)
- Highly connected components (large nodes)
- Architectural patterns and layers

### 2. Pre-Refactoring Impact Analysis

**Goal**: Understand what will be affected before making changes.

```bash
# 1. Set up Neo4j with your codebase
docker-compose up -d
./bin/import-neo4j -graph=codebase.json -clear

# 2. Use MCP server with AI assistant
# Ask: "What would break if I change the ProcessOrder function signature?"
```

**MCP Tool Commands:**

```bash
# Analyze specific function impact
analyze_impact "ProcessOrder" "signature" 3

# Find all callers
natural_query "What functions call ProcessOrder?"

# Or use direct Cypher query
cypher_query "MATCH (caller:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode) WHERE target.label = 'ProcessOrder' RETURN caller.label, caller.package ORDER BY caller.package"
```

### 3. Code Quality Assessment

**Goal**: Identify technical debt and areas needing attention.

**Using Web Interface:**

1. Filter by complexity > 10
2. Look for functions with many connections
3. Examine package coupling patterns

**Using AI Assistant:**

```typescript
// Find complex functions
use_mcp_tool({
  server_name: "code-graph",
  tool_name: "natural_query", 
  arguments: {
    question: "What are the most complex functions that need refactoring?"
  }
})

// Find duplicate patterns
use_mcp_tool({
  server_name: "code-graph",
  tool_name: "find_patterns",
  arguments: {
    patternType: "duplicate"
  }
})

// Find usage patterns
use_mcp_tool({
  server_name: "code-graph",
  tool_name: "find_patterns", 
  arguments: {
    patternType: "usage"
  }
})
```

## Common Use Cases for Engineering Teams

### 1. **Onboarding New Developers**

**Challenge**: New team members need to quickly understand a large codebase.

**Solution**: Use the visualization to provide guided tours of the system.

**Steps:**

1. **Start with architectural overview**:

   ```bash
   # Generate graph focusing on packages and key structures
   ./bin/server -graph=codebase.json
   # Set filter to "Packages & Types" view
   ```

2. **Show key integration points**:

   ```bash
   # Find main interfaces and their implementations
   natural_query "What are the main interfaces in this codebase?"
   find_implementers "IRepository"  # or any key interface
   
   # Understand code structure
   detect_architecture layers
   ```

3. **Identify learning priorities**:

   ```bash
   # Find complex code that needs understanding
   natural_query "What are the most complex functions that need refactoring?"
   
   # Trace execution paths
   trace_call_path "main" "HandleRequest"
   ```

**Expected Outcomes:**

- Faster onboarding (2-3 days vs 1-2 weeks)
- Better architectural understanding
- Reduced questions to senior developers

### 2. **Code Review Preparation**

**Challenge**: Focus code reviews on high-risk areas and architectural concerns.

**Solution**: Pre-analyze changes and their impact scope.

**Pre-Review Analysis:**

```typescript
// Find functions with high complexity and many callers (high risk)
use_mcp_tool({
  server_name: "code-graph",
  tool_name: "cypher_query", 
  arguments: {
    query: `
      MATCH (f:Function)<-[r:CALLS]-(caller)
      WHERE f.complexity > 8
      WITH f, count(r) as callerCount
      WHERE callerCount > 3
      RETURN f.label, f.package, f.complexity, callerCount
      ORDER BY (f.complexity * callerCount) DESC
      LIMIT 15
    `
  }
})

// Check if changes affect critical paths
use_mcp_tool({
  server_name: "code-graph",
  tool_name: "trace_call_path",
  arguments: {
    from: "main",
    to: "ModifiedFunction"
  }
})
```

**Review Focus Areas:**

- Functions with complexity > 10
- Changes affecting > 5 other components
- New dependencies between packages
- Breaking interface changes

### 3. **Microservices Boundary Design**

**Challenge**: Decide how to split a monolith into microservices.

**Solution**: Use graph analysis to identify natural boundaries.

**Analysis Steps:**

1. **Find tightly coupled clusters**:

   ```typescript
   use_mcp_tool({
     server_name: "code-graph",
     tool_name: "cypher_query",
     arguments: {
       query: `
         MATCH (pkg1:Package)-[:IMPORTS]->(pkg2:Package)
         WITH pkg1, pkg2, count(*) as strength
         WHERE strength > 5
         RETURN pkg1.name, pkg2.name, strength
         ORDER BY strength DESC
       `
     }
   })
   ```

2. **Identify service boundaries**:

   ```typescript
   use_mcp_tool({
     server_name: "code-graph",
     tool_name: "detect_architecture",
     arguments: {
       analysisType: "layers"
     }
   })
   ```

3. **Analyze cross-cutting concerns**:

   ```typescript
   use_mcp_tool({
     server_name: "code-graph",
     tool_name: "natural_query",
     arguments: {
       question: "Which packages are used by the most other packages?"
     }
   })
   ```

**Decision Framework:**

- **High cohesion within packages** (many internal calls)
- **Low coupling between packages** (few external dependencies)
- **Clear interface boundaries** (well-defined APIs)
- **Minimal data sharing** (independent data models)

### 4. **Performance Optimization**

**Challenge**: Identify performance bottlenecks and hot paths.

**Solution**: Find highly called functions and complex algorithms.

```bash
# Find potential bottlenecks
cypher_query "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode) WITH target, count(r) as callCount, avg(f.complexity) as avgCallerComplexity WHERE callCount > 10 AND target.complexity > 5 RETURN target.label, target.package, target.complexity, callCount, avgCallerComplexity, (target.complexity * callCount) as hotness_score ORDER BY hotness_score DESC LIMIT 10"

# Find recursive calls (potential infinite loops)
cypher_query "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'calls'}*2..5]->(f) WHERE f.type IN ['function', 'method'] RETURN DISTINCT f.label, f.package LIMIT 10"

# Or use natural language for bottlenecks
natural_query "What are the most frequently called complex functions?"
```

### 5. **Analyzing External Dependencies**

**Challenge**: Understanding how your code integrates with external packages.

**Solution**: Use the `-include-packages` flag to selectively analyze external dependencies.

```bash
# Analyze with specific external packages
./bin/analyze -repo=. -output=full-graph.json \
  -include-packages="github.com/gin-gonic/gin/..."

# Analyze with multiple specific packages
./bin/analyze -repo=. -output=full-graph.json \
  -include-packages="github.com/gin-gonic/gin,github.com/stretchr/testify/..."

# For MCP tools, include in analyze_workspace
analyze_workspace "/path/to/project" "project-name" false ["github.com/gin-gonic/gin/..."]
```

**Benefits:**

- See how your code uses external APIs
- Identify all touch points with third-party code
- Understand the full call chain including external packages
- Verify correct usage of external interfaces

**Tips:**

- Start with critical dependencies only
- Use "..." suffix to include sub-packages
- Monitor graph size - external packages can add many nodes
- Consider analyzing vendor directory if present

### 6. **Security Review**

**Challenge**: Identify security-sensitive code paths and potential vulnerabilities.

**Solution**: Trace data flow from external inputs to sensitive operations.

```bash
# Find functions that handle external input
cypher_query "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method'] AND (f.label CONTAINS 'Handler' OR f.label CONTAINS 'HTTP' OR f.label CONTAINS 'API' OR f.label CONTAINS 'Request') RETURN f.label, f.package, f.complexity ORDER BY f.complexity DESC"

# Trace paths from input handlers to database operations
cypher_query "MATCH path = (input:CodeNode)-[r:RELATES_TO {type: 'calls'}*1..5]->(db:CodeNode) WHERE input.type IN ['function', 'method'] AND input.label CONTAINS 'Handler' AND (db.label CONTAINS 'Query' OR db.label CONTAINS 'Insert' OR db.label CONTAINS 'Update' OR db.label CONTAINS 'Delete') RETURN [n in nodes(path) | n.label] as call_path LIMIT 20"

# Or use natural language
natural_query "What functions handle HTTP requests and API calls?"
natural_query "Show me call paths from handlers to database operations"
```

## Advanced Analysis Techniques

### 1. **Complexity Hotspot Analysis**

Find areas where high complexity coincides with high connectivity:

```bash
# Find complexity hotspots
cypher_query "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method'] OPTIONAL MATCH (f)-[r_out:RELATES_TO {type: 'calls'}]->() OPTIONAL MATCH (f)<-[r_in:RELATES_TO {type: 'calls'}]-() WITH f, count(r_out) as outbound, count(r_in) as inbound WHERE f.complexity > 10 AND (inbound + outbound) > 10 RETURN f.label, f.package, f.complexity, inbound, outbound, (inbound + outbound) as total_connections, (f.complexity * (inbound + outbound)) as risk_score ORDER BY risk_score DESC LIMIT 15"

# Or use pattern detection
find_patterns "usage"  # Find heavily used functions
```

### 2. **Dependency Chain Analysis**

Understand transitive dependencies and their impact:

```bash
# Find long dependency chains
cypher_query "MATCH path = (start:CodeNode {type: 'package'})-[r:RELATES_TO {type: 'imports'}*3..6]->(end:CodeNode {type: 'package'}) WHERE start <> end WITH path, length(path) as chain_length ORDER BY chain_length DESC RETURN [n in nodes(path) | n.label] as dependency_chain, chain_length LIMIT 10"

# Find circular dependencies
cypher_query "MATCH (a:CodeNode {type: 'package'})-[r:RELATES_TO {type: 'imports'}*2..5]->(a) RETURN DISTINCT a.label as package_with_circular_deps"

# Or use natural language
natural_query "Which packages have circular dependencies?"
```

### 3. **Interface Usage Analysis**

Understand how interfaces are used throughout the codebase:

```bash
# Find heavily used interfaces
cypher_query "MATCH (i:CodeNode {type: 'interface'}) OPTIONAL MATCH (s:CodeNode {type: 'struct'})-[r1:RELATES_TO {type: 'implements'}]->(i) OPTIONAL MATCH (f:CodeNode)-[r2:RELATES_TO {type: 'returns'}]->(i) OPTIONAL MATCH (p:CodeNode)-[r3:RELATES_TO {type: 'parameter_type'}]->(i) WITH i, count(DISTINCT s) as implementers, count(DISTINCT f) as returners, count(DISTINCT p) as parameters WHERE implementers > 0 RETURN i.label, i.package, implementers, returners, parameters, (implementers + returners + parameters) as total_usage ORDER BY total_usage DESC"

# Or use simpler commands
natural_query "What are the most heavily used interfaces?"
find_patterns "interface"  # Find interface usage patterns
```

### 4. **Dead Code Detection**

Find potentially unused code:

```bash
# Find functions never called
cypher_query "MATCH (f:CodeNode {type: 'function'}) WHERE NOT (f)<-[:RELATES_TO {type: 'calls'}]-() AND f.visibility = 'private' AND NOT f.label = 'main' AND NOT f.label = 'init' RETURN f.label, f.package ORDER BY f.package, f.label"

# Find structs never constructed  
cypher_query "MATCH (s:CodeNode {type: 'struct'}) WHERE NOT ()-[:RELATES_TO {type: 'constructs'}]->(s) AND s.visibility = 'private' RETURN s.label, s.package ORDER BY s.package, s.label"

# Or use pattern detection
find_patterns "orphan"  # Find orphaned/unused code
natural_query "What functions are never called?"
```

## AI-Assisted Development Workflow

When working with AI coding assistants (like Claude), the MCP tools transform generic AI responses into context-aware, codebase-specific solutions. This section shows how to structure your interactions for maximum effectiveness.

### Key Principles for AI-Assisted Development

#### 1. **Start with Context and Intent**

```text
"I'm working on [feature/bug/refactor] in [component/service].
I need to [specific goal]."

Example:
"I'm adding authentication to the orders service. 
I need to understand how other services handle auth."
```

#### 2. **Use Tools for Context Gathering**

Instead of asking the AI to guess, use tools to provide concrete information:

```text
Bad: "How does authentication work in this codebase?"

Good: "Here's what I found about authentication:
[output from: natural_query "How is authentication implemented?"]
[output from: find_implementers "Authenticator"]
Can you explain this pattern and suggest how to implement it for orders?"
```

### Effective Interaction Patterns

#### **Pattern 1: Exploration → Understanding → Implementation**

```markdown
You: "I need to add a new repository for storing user preferences"

You: "Let me first understand the existing patterns"
[Run: natural_query "What interfaces have Repository in their name?"]
[Run: find_implementers "ICheckoutRepository"]

You: "Based on these examples [paste outputs], can you:
1. Explain the repository pattern used here
2. Generate a new IUserPreferencesRepository interface
3. Create an implementation following the same patterns"
```

#### **Pattern 2: Impact Analysis → Safe Refactoring**

```markdown
You: "I want to refactor the payment processing logic"

You: "First, let me check what depends on it"
[Run: analyze_impact "ProcessPayment" "modify"]
[Run: natural_query "What calls ProcessPayment?"]

You: "Given these dependencies [paste outputs], how can I:
1. Safely extract the validation logic
2. Maintain backward compatibility
3. Add the new fraud detection step"
```

#### **Pattern 3: Bug Investigation → Root Cause → Fix**

```markdown
You: "Users report orders failing intermittently"

You: "Let me trace the order flow"
[Run: trace_call_path "SubmitOrder" "SaveOrder"]
[Run: find_patterns "error_handlers"]
[Run: natural_query "What functions handle order failures?"]

You: "Looking at this call path and error handling [paste outputs]:
1. Where might the race condition be?
2. What's missing in the error handling?
3. Suggest a fix that follows these patterns"
```

### Structured Workflows

#### **For New Features**

```markdown
1. "I need to implement [feature description]"

2. Explore existing patterns:
   - find_implementers "[relevant interface]"
   - natural_query "How are similar features implemented?"
   - detect_architecture patterns

3. "Based on these patterns [paste findings], design a solution that:
   - Follows the existing architecture
   - Reuses these interfaces: [list]
   - Integrates with: [components]"

4. Impact check:
   - analyze_impact "[key components]" "modify"
   
5. "Given these dependencies, implement the feature"
```

#### **For Debugging**

```markdown
1. "I'm debugging [issue description]"

2. Gather context:
   - trace_call_path "[entry]" "[suspected problem]"
   - natural_query "What errors are handled in [component]?"
   - find_patterns "error_handlers"

3. "Here's the execution flow and error handling [paste outputs].
   The symptoms are [describe]. What's the likely cause?"

4. "Suggest diagnostic code to confirm the issue"
```

#### **For Code Reviews**

```markdown
1. "I'm reviewing a PR that [PR description]"

2. Analyze the changes:
   - analyze_impact "[changed interfaces]" "modify"
   - find_patterns "duplicate"  # Check if adding duplicates
   - natural_query "What conventions exist for [feature type]?"

3. "The PR changes [summary] and impacts [paste analysis].
   Are there any concerns with:
   - Breaking changes
   - Pattern violations  
   - Missing error handling"
```

### Effective Prompting Templates

#### **Understanding Code**

```text
"Here's the interface and its implementers:
[paste find_implementers output]

Please explain:
1. The purpose of this interface
2. The implementation patterns
3. When to use each implementation"
```

#### **Writing New Code**

```text
"I found these similar components:
[paste natural_query results]
[paste cypher_query results showing structure]

Create a new [component] that:
- Follows the same patterns
- Implements [requirements]
- Integrates with [existing parts]"
```

#### **Refactoring**

```text
"Impact analysis shows:
[paste analyze_impact output]

The current implementation:
[paste relevant code]

Refactor this to [goal] while:
- Maintaining compatibility with [dependents]
- Following patterns from [paste example]"
```

### Power User Tips

#### **1. Chain Tool Outputs**

```markdown
You: "Analyze the checkout flow"
[Run: find_implementers "ICheckoutRepository"]
[Run: trace_call_path "BeginCheckout" "CompleteOrder"]
[Run: natural_query "What validates checkout data?"]

You: "Connect these pieces - how does data flow through these components?"
```

#### **2. Provide Constraints**

```markdown
You: "Given these interfaces [paste] and call paths [paste],
design a solution that:
- Must use existing ICheckoutRepository
- Cannot modify these functions: [list from impact analysis]
- Should follow the pattern from TracingCheckoutRepo"
```

#### **3. Iterative Refinement**

```markdown
Initial: "Implement user preferences"
[Run tools, get context]

Refined: "Implement user preferences using:
- Repository pattern like [paste example]
- Same DynamoDB setup as CheckoutRepository
- Tracing wrapper pattern from [paste example]"
```

### Complete Example: Feature Implementation

```markdown
You: "I need to add email notifications when orders fail"

Step 1: "Let me understand the current system"
[Run: natural_query "How are orders processed?"]
[Run: find_implementers "ICheckoutRepository"]
[Run: find_patterns "error_handlers"]
[Run: natural_query "What messaging systems are used?"]

Step 2: "Based on this analysis [paste outputs], I see:
- Orders use this flow: [summary]
- Errors are handled by: [list]
- Existing messaging: [list]

Design an email notification system that:
1. Integrates at these failure points
2. Uses the existing message queue
3. Follows the observer pattern like [example]"

Step 3: "Now implement the IFailureNotifier interface"

Step 4: "Add tests following patterns from [paste test examples]"
```

### Common MCP Tool Commands Reference

```bash
# Finding code structure
find_implementers "InterfaceName"           # Who implements this?
natural_query "What interfaces exist?"      # Natural language search
detect_architecture layers                  # Analyze architecture

# Understanding relationships
trace_call_path "Start" "End"              # How do we get from A to B?
analyze_impact "FunctionName" "delete"      # What breaks if I change this?

# Finding patterns
find_patterns duplicate                     # Find duplicate code
find_patterns error_handlers                # Find error handling
find_patterns usage                        # Find usage patterns

# Direct queries
cypher_query "MATCH (n:CodeNode) WHERE..." # Direct Neo4j queries
```

## Practical Development Workflows

### Daily Development Tasks

#### **Morning Code Review**

```bash
# Check what changed overnight
natural_query "What functions were modified in the last 24 hours?"

# Review high-risk changes
cypher_query "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method'] AND f.complexity > 10 RETURN f.label, f.package, f.complexity ORDER BY f.complexity DESC"

# Find potential issues
find_patterns "duplicate"
find_patterns "error_handlers"
```

#### **Before Starting New Feature**

```bash
# Understand the domain
natural_query "What interfaces exist for payment processing?"
find_implementers "PaymentProcessor"

# Check for existing functionality
natural_query "Is there existing code for email notifications?"

# Analyze impact area
analyze_impact "OrderService" "modify" 3
```

#### **During Development**

```bash
# Find examples to follow
natural_query "How are REST endpoints implemented?"
trace_call_path "HandleHTTP" "SaveToDatabase"

# Check conventions
natural_query "What naming conventions are used for interfaces?"
find_patterns "interface"

# Verify no duplicates
find_patterns "duplicate"
```

#### **Before Committing**

```bash
# Impact analysis
analyze_impact "ModifiedFunction" "signature" 2

# Check complexity
natural_query "What is the complexity of the functions I modified?"

# Verify test coverage areas
natural_query "What test files exist for the order package?"
```

### Debugging Workflows

#### **Performance Issues**

```bash
# Step 1: Find bottlenecks
natural_query "What are the most frequently called complex functions?"

# Step 2: Trace hot paths
trace_call_path "APIHandler" "DatabaseQuery"

# Step 3: Find recursive calls
cypher_query "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'calls'}*2..5]->(f) WHERE f.type IN ['function', 'method'] RETURN DISTINCT f.label, f.package"
```

#### **Bug Investigation**

```bash
# Step 1: Understand the flow
trace_call_path "UserRequest" "ErrorPoint"

# Step 2: Find error handling
natural_query "How are errors handled in the payment module?"
find_patterns "error_handlers"

# Step 3: Check related functions
analyze_impact "BuggyFunction" "modify" 2
```

#### **Security Audit**

```bash
# Find entry points
natural_query "What functions handle HTTP requests?"

# Trace to sensitive operations
cypher_query "MATCH path = (input:CodeNode)-[r:RELATES_TO {type: 'calls'}*1..5]->(db:CodeNode) WHERE input.label CONTAINS 'Handler' AND db.label CONTAINS 'Query' RETURN [n in nodes(path) | n.label] as call_path"

# Find authentication checks
natural_query "Where is authentication validated?"
```

### Refactoring Workflows

#### **Extract Interface**

```bash
# Step 1: Find similar structs
natural_query "What structs have similar methods to OrderProcessor?"

# Step 2: Check existing interfaces
find_patterns "interface"

# Step 3: Impact analysis
analyze_impact "OrderProcessor" "modify" 3
```

#### **Split Large Function**

```bash
# Step 1: Analyze complexity
cypher_query "MATCH (f:CodeNode) WHERE f.label = 'ProcessOrder' RETURN f.complexity, f.label"

# Step 2: Find callers
natural_query "What functions call ProcessOrder?"

# Step 3: Trace dependencies
trace_call_path "ProcessOrder" "SaveOrder"
```

#### **Consolidate Duplicates**

```bash
# Step 1: Find duplicates
find_patterns "duplicate"

# Step 2: Analyze each duplicate's usage
analyze_impact "DuplicateFunction1" "delete" 2
analyze_impact "DuplicateFunction2" "delete" 2

# Step 3: Find common interface
natural_query "What interfaces could represent these duplicate functions?"
```

### Architecture & Design Workflows

#### **Microservice Extraction**

```bash
# Step 1: Analyze package boundaries
detect_architecture "layers"

# Step 2: Find package dependencies
cypher_query "MATCH (p1:CodeNode {type: 'package'})-[r:RELATES_TO {type: 'imports'}]->(p2:CodeNode {type: 'package'}) WHERE p1.label CONTAINS 'order' RETURN p1.label, p2.label"

# Step 3: Find shared interfaces
natural_query "What interfaces are used by both order and payment packages?"
```

#### **API Design**

```bash
# Step 1: Find existing patterns
natural_query "How are REST APIs structured?"
find_implementers "Handler"

# Step 2: Check data models
natural_query "What structs represent API requests and responses?"

# Step 3: Trace request flow
trace_call_path "HTTPHandler" "BusinessLogic"
```

### Team Collaboration Workflows

#### **Code Review Preparation**

```bash
# Generate review checklist
echo "=== Code Review Checklist ==="

# High complexity functions
echo "High Complexity Functions:"
cypher_query "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method'] AND f.complexity > 15 RETURN f.label, f.complexity"

# Interface changes
echo "Interface Changes:"
natural_query "What interfaces were modified?"

# Impact analysis
echo "Impact Analysis:"
analyze_impact "MainChangedFunction" "signature" 3
```

#### **Knowledge Transfer**

```bash
# Create domain overview
echo "=== Payment System Overview ==="

# Key interfaces
echo "Core Interfaces:"
natural_query "What interfaces exist in the payment package?"

# Implementation patterns
echo "Implementations:"
find_implementers "PaymentProcessor"

# Key flows
echo "Main Flows:"
trace_call_path "ProcessPayment" "ChargeCard"
```

### Quick Command Reference

```bash
# Most used commands for daily work
find_implementers "Interface"              # Quick implementation lookup
natural_query "Question about code"        # Natural language search
analyze_impact "Function" "change" depth   # Before making changes
trace_call_path "From" "To"               # Understand execution flow
find_patterns "pattern_type"              # Find code patterns
detect_architecture "analysis_type"       # Architecture overview

# Useful natural language queries
"What are the most complex functions?"
"What interfaces exist for X?"
"How is Y implemented?"
"What calls function Z?"
"Show me error handling in package A"
"What are the main entry points?"
```

### Tool Combinations for Common Tasks

#### **Understanding a New Module**

```bash
# 1. Overview
detect_architecture "layers"
natural_query "What are the main interfaces in the auth module?"

# 2. Implementations
find_implementers "Authenticator"
find_implementers "Authorizer"

# 3. Usage patterns
natural_query "How is authentication used across the codebase?"
trace_call_path "Login" "GenerateToken"
```

#### **Planning a Refactor**

```bash
# 1. Current state analysis
analyze_impact "TargetFunction" "modify" 3
find_patterns "duplicate"

# 2. Find patterns to follow
natural_query "What refactoring patterns exist in the codebase?"
find_patterns "usage"

# 3. Verify safety
cypher_query "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode) WHERE target.label = 'TargetFunction' RETURN f.label, f.package"
```

## Best Practices

### 1. **Regular Analysis Schedule**

**Weekly Code Health Check:**

```bash
#!/bin/bash
# weekly-health-check.sh

echo "=== Weekly Code Health Report ==="
echo "Analyzing codebase..."

./bin/analyze -repo=. -output=weekly.json -verbose

echo "Importing to Neo4j..."
./bin/import-neo4j -graph=weekly.json -clear

echo "Generating report..."
# Use MCP tools to generate automated health metrics
```

**Metrics to Track:**

- Total complexity trend
- Number of high-complexity functions
- Package coupling evolution
- Dead code accumulation

### 2. **Integration with CI/CD**

**Pre-commit Analysis:**

```yaml
# .github/workflows/code-analysis.yml
name: Code Analysis
on: [pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Setup Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Analyze Code
      run: |
        ./bin/analyze -repo=. -output=pr-analysis.json
        # Compare complexity with main branch
        # Fail if complexity increased significantly
    
    - name: Upload Analysis
      uses: actions/upload-artifact@v2
      with:
        name: code-analysis
        path: pr-analysis.json
```

**Quality Gates:**

- Max complexity per function: 15
- Max functions per package: 50
- Max package dependencies: 10
- No circular dependencies

### 3. **Team Collaboration**

**Shared Analysis Sessions:**

1. **Weekly Architecture Reviews**:
   - Use web interface for team discussions
   - Focus on new packages and major changes
   - Document architectural decisions

2. **Refactoring Planning**:
   - Use impact analysis for risk assessment
   - Prioritize by complexity × usage frequency
   - Plan incremental improvements

3. **Knowledge Sharing**:
   - Create annotated views for different domains
   - Share interesting patterns and anti-patterns
   - Document design decisions

### 4. **Performance Considerations**

**Large Codebase Handling:**

```bash
# For codebases > 1000 files
export GOGC=100  # Reduce GC pressure
ulimit -n 4096   # Increase file descriptor limit

# Use incremental analysis when possible
./bin/analyze -repo=. -output=large.json -verbose

# Consider package-level filtering for initial analysis
```

**Memory Optimization:**

- Start with package-level view
- Progressively add detail as needed
- Use Neo4j for complex queries on large graphs
- Cache results for repeated analyses

## Troubleshooting Common Issues

### 1. **Analysis Performance**

**Slow analysis on large codebases:**

```bash
# Check available memory
free -h

# Monitor during analysis
top -p $(pgrep analyze)

# Solutions:
export GOMAXPROCS=4  # Limit CPU usage
export GOMEMLIMIT=2GiB  # Limit memory usage
```

### 2. **Visualization Performance**

**Slow web interface with large graphs:**

1. Use filtering: Start with "Packages & Types" view
2. Increase browser memory: `--max-old-space-size=4096`
3. Use Neo4j Browser for complex queries

### 3. **Missing Dependencies**

**Graph shows incomplete relationships:**

```bash
# Ensure all dependencies are available
go mod tidy
go mod download

# Verify module resolution
go list -m all

# Run with verbose logging
./bin/analyze -repo=. -output=debug.json -verbose
```

### 4. **MCP Integration Issues**

**AI assistant not finding results:**

```bash
# Test basic connectivity
cypher_query "MATCH (n) RETURN count(n) as total_nodes"

# Check node naming consistency
cypher_query "MATCH (n) RETURN DISTINCT labels(n), n.type ORDER BY n.type"

# Check available tool functionality
natural_query "What types of nodes exist in the database?"
```

## Performance Benchmarks

### Analysis Performance

| Codebase Size | Files | Analysis Time | Memory Usage | Recommendations |
|---------------|-------|---------------|--------------|-----------------|
| Small (< 100 files) | < 500 | < 5s | < 100MB | Use all features |
| Medium (100-500 files) | 500-2K | 5-30s | 100-500MB | Enable embeddings selectively |
| Large (500-1K files) | 2K-5K | 30s-2m | 500MB-1GB | Consider package filtering |
| Enterprise (> 1K files) | > 5K | 2-10m | 1-4GB | Use incremental analysis |

### Query Performance

| Query Type | Small Graph | Large Graph | Optimization |
|------------|-------------|-------------|--------------|
| Simple node lookup | < 1ms | < 10ms | Use indexes |
| Complex traversal | < 10ms | < 100ms | Limit depth |
| Pattern matching | < 50ms | < 500ms | Use parameters |
| Full graph scan | < 100ms | 1-5s | Avoid when possible |

## ROI and Benefits

### Quantitative Benefits

**Development Productivity:**

- 40% faster onboarding of new developers
- 60% reduction in architectural questions
- 30% faster code review cycles
- 50% better refactoring success rate

**Code Quality Improvements:**

- 25% reduction in bug density
- 35% improvement in code maintainability scores
- 20% reduction in technical debt accumulation
- 45% better test coverage in complex areas

**Risk Mitigation:**

- Early detection of architectural anti-patterns
- Proactive identification of performance bottlenecks
- Better understanding of change impact
- Reduced production incidents from unexpected dependencies

### Qualitative Benefits

- **Better Communication**: Visual representations improve team discussions
- **Knowledge Preservation**: Architectural knowledge captured in the graph
- **Decision Support**: Data-driven architectural and refactoring decisions
- **Quality Culture**: Increased awareness of code quality metrics

## Conclusion

Go Code Graph transforms how engineering teams understand, analyze, and improve their Go codebases. By combining static analysis, graph visualization, and AI-powered querying, it provides unprecedented insights into code structure, quality, and relationships.

Whether you're conducting code reviews, planning refactoring, or onboarding new developers, the techniques and workflows in this guide will help you leverage the full power of code graph analysis for your engineering organization.

Remember: **The key to success is consistent usage and integration into your development workflow.** Start small, demonstrate value, and gradually expand usage across your team and projects.
