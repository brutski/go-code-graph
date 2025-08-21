# MCP Prompts Guide

This guide describes the comprehensive set of MCP prompts available in the Go Code Graph MCP server.
These prompts provide structured templates with actual tool usage examples for common code analysis tasks, making it easier to leverage the full power of the tool.

## What are MCP Prompts?

MCP prompts are predefined templates that:

- **Show exact tool usage** - Each prompt includes the specific tools to call with their arguments
- **Chain multiple tools** - Combine tools in the right order for comprehensive analysis  
- **Provide structure** - Guide through complex analysis tasks step-by-step
- **Include best practices** - Use optimal queries and patterns for each task

The prompts don't automatically execute tools, but they show the AI assistant exactly which tools to use and how to use them effectively.

## Available Prompts

### 1. Initial Codebase Analysis

**Prompt**: `analyze_new_codebase`

**Purpose**: Get a comprehensive overview when analyzing a Go codebase for the first time.

**Arguments**:

- `workspace_path` (required): Path to the Go project directory
- `include_external` (optional): External packages to include (comma-separated)

**What it does**:

1. Analyzes the workspace and imports it to the graph database
2. Detects architectural patterns and layers
3. Identifies complex functions needing refactoring
4. Highlights code quality issues
5. Optionally includes external dependencies in analysis

**Example usage**:

```text
Use the analyze_new_codebase prompt with:
- workspace_path: /home/user/my-go-project
- include_external: github.com/gin-gonic/gin,github.com/stretchr/testify
```

**What the prompt provides**:
The prompt will show the AI assistant the exact sequence of tools to use:

1. `analyze_workspace` with the provided path and external packages
2. `detect_architecture` to understand system design  
3. `natural_query` to find complex functions
4. `find_patterns` to identify code quality issues

Each tool call includes the exact arguments and parameters needed.

### 2. Code Quality Assessment

**Prompt**: `assess_code_quality`

**Purpose**: Perform a comprehensive code quality assessment with actionable insights.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `focus_area` (optional): Specific package or component to focus on

**What it does**:

1. Finds complex functions that need refactoring
2. Identifies duplicate code patterns
3. Detects unused code (dead code)
4. Finds god objects with too many responsibilities
5. Identifies functions with too many parameters
6. Checks for error handling issues

**Example usage**:

```text
Use the assess_code_quality prompt with:
- workspace: my-go-project
- focus_area: internal/api
```

### 3. Architecture Understanding

**Prompt**: `understand_architecture`

**Purpose**: Get a comprehensive understanding of the codebase architecture.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `component` (optional): Specific component or service to analyze

**What it does**:

1. Identifies main architectural layers
2. Shows how key components interact
3. Lists main interfaces and their implementations
4. Detects circular dependencies
5. Maps external dependencies
6. Identifies main entry points

**Example usage**:

```text
Use the understand_architecture prompt with:
- workspace: my-go-project
- component: order-service
```

### 4. Change Impact Analysis

**Prompt**: `analyze_change_impact`

**Purpose**: Analyze the impact of changing a specific component before making changes.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `component_name` (required): Name of the function, struct, or interface to change
- `change_type` (required): Type of change (signature, delete, modify, rename)

**What it does**:

1. Analyzes direct and indirect impact
2. Finds all callers and users
3. Identifies affected interfaces
4. Checks for test files needing updates
5. Suggests safe migration approach
6. Lists potential breaking changes

**Example usage**:

```text
Use the analyze_change_impact prompt with:
- workspace: my-go-project
- component_name: ProcessOrder
- change_type: signature
```

### 5. Debugging Support

**Prompt**: `debug_execution_flow`

**Purpose**: Understand execution flow for debugging purposes.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `entry_point` (required): Starting function or handler
- `target_point` (required): Target function or suspected problem area

**What it does**:

1. Traces the call path between functions
2. Identifies all intermediate functions
3. Finds error handling along the path
4. Checks for goroutine spawning
5. Identifies external calls and I/O operations
6. Highlights potential bottlenecks

**Example usage**:

```text
Use the debug_execution_flow prompt with:
- workspace: my-go-project
- entry_point: HandleHTTPRequest
- target_point: SaveToDatabase
```

### 6. Interface Usage Analysis

**Prompt**: `analyze_interface_usage`

**Purpose**: Understand how interfaces are used throughout the codebase.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `interface_name` (required): Name of the interface to analyze

**What it does**:

1. Finds all implementations
2. Shows where interface is used as parameter
3. Finds functions returning the interface
4. Checks for single-implementation interfaces
5. Verifies all methods are used
6. Suggests interface segregation opportunities

**Example usage**:

```text
Use the analyze_interface_usage prompt with:
- workspace: my-go-project
- interface_name: Repository
```

### 7. Refactoring Planning

**Prompt**: `plan_refactoring`

**Purpose**: Get detailed guidance for refactoring a specific component.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `target` (required): Function, package, or component to refactor
- `goal` (required): Refactoring goal (e.g., reduce complexity, extract interface)

**What it does**:

1. Analyzes current implementation complexity
2. Maps all dependencies and callers
3. Finds similar patterns to follow
4. Identifies affected tests
5. Provides step-by-step approach
6. Highlights risks and breaking changes

**Example usage**:

```text
Use the plan_refactoring prompt with:
- workspace: my-go-project
- target: PaymentProcessor
- goal: extract payment gateway interface
```

### 8. Code Pattern Search

**Prompt**: `find_code_patterns`

**Purpose**: Search for specific code patterns or implementations.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `pattern` (required): Pattern to search for (e.g., error handling, HTTP handlers)
- `package_filter` (optional): Limit search to specific package

**What it does**:

1. Searches for matching functions
2. Finds similar implementations
3. Identifies anti-patterns
4. Shows common approaches
5. Highlights inconsistencies
6. Provides best practice recommendations

**Example usage**:

```text
Use the find_code_patterns prompt with:
- workspace: my-go-project
- pattern: database queries
- package_filter: internal/repository
```

### 9. Dependency Analysis

**Prompt**: `analyze_dependencies`

**Purpose**: Analyze dependencies between packages or components.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `source_package` (required): Source package to analyze
- `include_external` (optional): Include external dependencies (true/false)

**What it does**:

1. Lists packages imported by source
2. Finds packages importing source
3. Identifies circular dependencies
4. Shows dependency depth
5. Finds tightly coupled packages
6. Provides decoupling recommendations

**Example usage**:

```text
Use the analyze_dependencies prompt with:
- workspace: my-go-project
- source_package: internal/auth
- include_external: true
```

### 10. Performance Analysis

**Prompt**: `find_performance_issues`

**Purpose**: Identify potential performance bottlenecks.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `focus_area` (optional): Specific area to analyze

**What it does**:

1. Finds high-complexity functions
2. Identifies frequently called functions
3. Detects recursive calls
4. Finds inefficient nested loops
5. Identifies goroutine spawning patterns
6. Detects potential N+1 query patterns

**Example usage**:

```text
Use the find_performance_issues prompt with:
- workspace: my-go-project
- focus_area: api/handlers
```

### 11. Test Coverage Analysis

**Prompt**: `analyze_test_coverage`

**Purpose**: Analyze test coverage and identify gaps.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `package` (required): Package to analyze

**What it does**:

1. Lists all public functions
2. Identifies functions with tests
3. Finds complex untested functions
4. Checks interface test implementations
5. Identifies integration test patterns
6. Prioritizes functions needing tests

**Example usage**:

```text
Use the analyze_test_coverage prompt with:
- workspace: my-go-project
- package: internal/service
```

### 12. API Migration Planning

**Prompt**: `plan_api_migration`

**Purpose**: Plan migration from one API/interface to another.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `old_api` (required): Current API or interface being used
- `new_api` (required): New API or interface to migrate to

**What it does**:

1. Finds all current API usages
2. Analyzes API differences
3. Identifies high-risk areas
4. Suggests wrapper patterns
5. Provides phased migration approach
6. Lists affected tests

**Example usage**:

```text
Use the plan_api_migration prompt with:
- workspace: my-go-project
- old_api: v1.PaymentAPI
- new_api: v2.PaymentGateway
```

### 13. Code Review Preparation

**Prompt**: `prepare_code_review`

**Purpose**: Prepare for code review by analyzing changes comprehensively.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `changed_components` (required): List of changed functions/structs (comma-separated)

**What it does**:

1. Analyzes complexity of changes
2. Finds affected callers
3. Checks interface impacts
4. Identifies breaking changes
5. Finds related code needing updates
6. Checks security-sensitive paths

**Example usage**:

```text
Use the prepare_code_review prompt with:
- workspace: my-go-project
- changed_components: ProcessOrder,OrderValidator,OrderRepository.Save
```

### 14. Developer Onboarding

**Prompt**: `onboard_new_developer`

**Purpose**: Create an onboarding guide for new developers.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `focus_area` (optional): Specific area the developer will work on

**What it does**:

1. Identifies main entry points
2. Lists key interfaces and purposes
3. Prioritizes packages to study
4. Highlights complex areas
5. Shows architectural patterns
6. Provides code examples

**Example usage**:

```text
Use the onboard_new_developer prompt with:
- workspace: my-go-project
- focus_area: payment-processing
```

### 15. Security Analysis

**Prompt**: `security_analysis`

**Purpose**: Analyze security-sensitive code paths.

**Arguments**:

- `workspace` (required): Workspace to analyze
- `entry_points` (required): Type of entry points to analyze (HTTP, API, CLI)

**What it does**:

1. Finds all entry points
2. Traces paths to sensitive operations
3. Identifies auth checks
4. Finds direct database queries
5. Checks input validation
6. Identifies exposed errors

**Example usage**:

```text
Use the security_analysis prompt with:
- workspace: my-go-project
- entry_points: HTTP
```

## Best Practices for Using Prompts

### 1. Start with Overview Prompts

Begin with `analyze_new_codebase` or `understand_architecture` to get context before diving into specific areas.

### 2. Chain Prompts for Comprehensive Analysis

Use multiple prompts in sequence:

1. `understand_architecture` → Get overview
2. `assess_code_quality` → Identify issues
3. `plan_refactoring` → Create improvement plan

### 3. Use Focused Analysis

When working on specific features, use targeted prompts:

- `analyze_interface_usage` for API design
- `debug_execution_flow` for troubleshooting
- `analyze_change_impact` before refactoring

### 4. Combine with Direct Tool Usage

Prompts provide structured analysis, but you can always use individual tools for specific queries:

- Use prompts for comprehensive analysis
- Use tools for quick specific questions

### 5. Iterative Refinement

Start with broad prompts and narrow down:

1. `assess_code_quality` → Find problem areas
2. `find_code_patterns` → Deep dive into specific issues
3. `plan_refactoring` → Create action plan

## How Prompts Guide Tool Usage

Each prompt contains explicit tool usage instructions like:

```json
Use tool: analyze_workspace
Arguments: {
  "workspacePath": "{{workspace_path}}",
  "workspaceName": "{{workspace_name}}", 
  "incremental": false
}
```

This tells the AI assistant:

- **Which tool** to use (`analyze_workspace`)
- **What arguments** to pass (with template variables replaced)
- **In what order** to execute tools
- **How to interpret** the results

## Tips for AI Assistants

When using these prompts with AI assistants like Claude:

1. **Provide Context**: Give the AI context about your project and goals
2. **Be Specific**: Use precise component names and clear change types
3. **Execute Tools**: The AI will execute each tool shown in the prompt
4. **Follow Up**: Ask the AI to elaborate on specific findings
5. **Request Examples**: Ask for code examples when implementing suggestions
6. **Verify Impact**: Always run impact analysis before making changes

## Example Workflow

Here's a complete workflow for analyzing and improving a codebase:

```text
1. "Use the analyze_new_codebase prompt for /my/project"
   → Get initial overview

2. "Use the assess_code_quality prompt focusing on the api package"
   → Identify specific issues

3. "Use the analyze_interface_usage prompt for the Repository interface"
   → Understand current design

4. "Use the plan_refactoring prompt to reduce complexity in ProcessOrder"
   → Get detailed refactoring plan

5. "Use the analyze_change_impact prompt before changing ProcessOrder signature"
   → Ensure safe changes

6. "Use the prepare_code_review prompt for the changes"
   → Get review checklist
```

## Integration with Development Workflow

### During Development

- Use `debug_execution_flow` when troubleshooting
- Use `find_code_patterns` to maintain consistency
- Use `analyze_dependencies` to avoid coupling

### Before Changes

- Use `analyze_change_impact` for risk assessment
- Use `plan_refactoring` for complex changes
- Use `analyze_interface_usage` for API changes

### Code Review

- Use `prepare_code_review` for thorough reviews
- Use `assess_code_quality` for objective metrics
- Use `security_analysis` for sensitive changes

### Team Collaboration

- Use `onboard_new_developer` for new team members
- Use `understand_architecture` for design discussions
- Use `analyze_dependencies` for module boundaries

## Conclusion

MCP prompts transform the Go Code Graph tool from a powerful but complex analysis system into an intuitive, guided experience.
By using these prompts, you can ensure comprehensive analysis, maintain consistency, and make informed decisions about your codebase.

Remember: These prompts are templates that guide analysis, but you can always customize the approach based on your specific needs.
The combination of structured prompts and flexible tools provides the best of both worlds for code analysis.
