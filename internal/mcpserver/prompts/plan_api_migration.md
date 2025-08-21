# Plan API Migration

Plan migration from {{old_api}} to {{new_api}}:

1. **Current usage**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find all usages of {{old_api}} in the codebase. Show me how it's currently being used",
     "context": "Understanding current API usage",
     "workspace": "{{workspace}}"
   }
   ```

2. **Impact analysis**:

   ```json
   Use tool: analyze_impact
   Arguments: {
     "nodeId": "{{old_api}}",
     "changeType": "delete",
     "maxDepth": 3,
     "workspace": "{{workspace}}"
   }
   ```

3. **Find adapter patterns**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Show me any adapter or wrapper patterns in the codebase that could help with API migration",
     "context": "Looking for migration patterns",
     "workspace": "{{workspace}}"
   }
   ```

Provide phased migration plan with risk mitigation.
