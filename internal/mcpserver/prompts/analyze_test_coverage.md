# Analyze Test Coverage

Analyze test coverage for {{package}}:

1. **All public functions**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method'] AND f.package = '{{package}}' AND f.visibility = 'public' RETURN f.label, f.complexity ORDER BY f.complexity DESC",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

2. **Functions with tests**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (test:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(f:CodeNode) WHERE f.package = '{{package}}' AND test.label STARTS WITH 'Test' RETURN DISTINCT f.label ORDER BY f.label",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Complex untested functions**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Which complex functions (complexity > 5) in {{package}} don't have any test coverage?",
     "context": "Finding untested complex code",
     "workspace": "{{workspace}}"
   }
   ```

Provide prioritized list of functions needing tests.
