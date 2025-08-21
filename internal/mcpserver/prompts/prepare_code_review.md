# Prepare Code Review

Prepare code review for: {{changed_components}}

1. **Complexity analysis**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "What is the complexity of these components: {{changed_components}}? Show me their metrics",
     "context": "Code review preparation",
     "workspace": "{{workspace}}"
   }
   ```

2. **Impact analysis**:

   ```json
   Use tool: analyze_impact
   Arguments: {
     "nodeId": "[first component from {{changed_components}}]",
     "changeType": "modify",
     "maxDepth": 2,
     "workspace": "{{workspace}}"
   }
   ```

3. **Similar code patterns**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find code similar to {{changed_components}} that should follow the same patterns", 
     "context": "Ensuring consistency",
     "workspace": "{{workspace}}"
   }
   ```

Provide focused review checklist.
