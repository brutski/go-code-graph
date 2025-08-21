# Plan Refactoring

I want to refactor {{target}} to {{goal}}. Please create a detailed plan:

1. **Analyze current implementation** - Understand what we're changing:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Show me the current implementation of {{target}} including its complexity, dependencies, and structure",
     "context": "Understanding current state before refactoring",
     "workspace": "{{workspace}}"
   }
   ```

2. **Find all dependencies** - What depends on this:

   ```json
   Use tool: analyze_impact
   Arguments: {
     "nodeId": "{{target}}",
     "changeType": "modify",
     "maxDepth": 3,
     "workspace": "{{workspace}}"
   }
   ```

3. **Identify similar patterns** - Learn from existing code:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find examples of {{goal}} in the codebase. Show me well-implemented patterns I can follow",
     "context": "Looking for patterns to follow",
     "workspace": "{{workspace}}"
   }
   ```

4. **Check existing tests** - Ensure safety:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (test:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode) WHERE target.label = '{{target}}' AND test.label CONTAINS 'Test' RETURN test.label, test.package ORDER BY test.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

5. **Find related code** - What else might need changes:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find code similar to {{target}} that might need the same refactoring",
     "context": "Finding related code patterns",
     "workspace": "{{workspace}}"
   }
   ```

6. **Refactoring plan** - Step by step approach:

   Based on the analysis, create a plan that includes:

   **Phase 1: Preparation**
   - List all tests that need to be verified/updated
   - Identify interfaces that might need changes
   - Create compatibility layer if needed

   **Phase 2: Core Refactoring**
   - Specific code changes needed
   - Order of operations to maintain functionality
   - How to achieve "{{goal}}"

   **Phase 3: Cleanup**
   - Update all callers
   - Remove deprecated code
   - Update documentation

   **Risk Assessment:**
   - Breaking changes identified
   - Mitigation strategies
   - Rollback plan

Provide code examples for key changes where helpful.
