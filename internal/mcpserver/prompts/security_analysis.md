# Security Analysis

Security analysis for {{entry_points}} entry points:

1. **Find entry points**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find all {{entry_points}} handlers and entry points in the codebase",
     "context": "Security analysis entry points",
     "workspace": "{{workspace}}"
   }
   ```

2. **Trace to sensitive operations**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH path = (entry:CodeNode)-[r:RELATES_TO {type: 'calls'}*1..5]->(sensitive:CodeNode) WHERE entry.label CONTAINS 'Handler' AND (sensitive.label CONTAINS 'Query' OR sensitive.label CONTAINS 'Execute' OR sensitive.label CONTAINS 'Write') RETURN [n in nodes(path) | n.label] as call_path LIMIT 20",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Authentication checks**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Where are authentication and authorization checks performed in {{entry_points}} handlers?",
     "context": "Security validation",
     "workspace": "{{workspace}}"
   }
   ```

Provide security recommendations.
