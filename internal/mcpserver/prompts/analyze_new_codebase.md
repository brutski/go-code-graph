# Analyze New Codebase

I need to analyze the Go codebase at {{workspace_path}}. Please perform a comprehensive analysis following these steps:

1. **Analyze the workspace** - Import and analyze the codebase:

   ```json
   Use tool: analyze_workspace
   Arguments: {
     "workspacePath": "{{workspace_path}}",
     "workspaceName": "{{workspace_name}}",
     "incremental": false{{#if include_external}},
     "allowedPackages": [{{include_external}}]{{/if}}
   }
   ```

2. **Detect architectural patterns** - Understand the system design:

   ```json
   Use tool: detect_architecture
   Arguments: {
     "analysisType": "layers",
     "workspace": "{{#if workspace_name}}{{workspace_name}}{{else}}{{workspace_path}}{{/if}}"
   }
   ```

3. **Find complex functions** - Identify refactoring candidates:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "What are the most complex functions that need refactoring? Show me functions with complexity > 10",
     "context": "Looking for refactoring candidates",
     "workspace": "{{#if workspace_name}}{{workspace_name}}{{else}}{{workspace_path}}{{/if}}"
   }
   ```

4. **Identify code quality issues** - Find potential problems:

   ```json
   Use tool: find_patterns
   Arguments: {
     "patternType": "duplicate",
     "workspace": "{{#if workspace_name}}{{workspace_name}}{{else}}{{workspace_path}}{{/if}}"
   }
   ```

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "What are the main code quality issues? Check for god objects, long parameter lists, and unused code",
     "context": "Code quality assessment",
     "workspace": "{{#if workspace_name}}{{workspace_name}}{{else}}{{workspace_path}}{{/if}}"
   }
   ```

5. **Summary** - After analysis, provide:
   - Overall codebase structure and size
   - Main architectural patterns found
   - Top 5 complex functions to refactor
   - Key code quality issues to address
   - Recommendations for improvement

Please execute these tools in order and provide insights based on the results.
