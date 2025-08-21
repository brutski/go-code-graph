# Debug Execution Flow

Help me debug the execution flow from {{entry_point}} to {{target_point}}:

1. **Trace the call path** - Find how we get from A to B:

   ```json
   Use tool: trace_call_path
   Arguments: {
     "from": "{{entry_point}}",
     "to": "{{target_point}}",
     "workspace": "{{workspace}}"
   }
   ```

2. **Analyze intermediate functions** - For each function in the path:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH path = shortestPath((start:CodeNode {label: '{{entry_point}}'})-[r:RELATES_TO {type: 'calls'}*1..10]->(end:CodeNode {label: '{{target_point}}'})) UNWIND nodes(path) as n WHERE n.type IN ['function', 'method'] RETURN n.label, n.package, n.complexity, n.signature ORDER BY n.label",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Find error handling** - Check error handling in the flow:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Show me all error handling between {{entry_point}} and {{target_point}}. Which functions return errors and how are they handled?",
     "context": "Debugging error propagation",
     "workspace": "{{workspace}}"
   }
   ```

4. **Check for goroutines** - Find concurrency in the path:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH path = shortestPath((start:CodeNode {label: '{{entry_point}}'})-[*1..10]->(end:CodeNode {label: '{{target_point}}'})) UNWIND nodes(path) as n OPTIONAL MATCH (n)-[r:RELATES_TO {type: 'spawns_goroutine'}]->(g) WHERE n.type IN ['function', 'method'] RETURN n.label, n.package, g.label as spawned_goroutine",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

5. **External calls** - Find I/O operations:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "What external calls (database, HTTP, file I/O) happen between {{entry_point}} and {{target_point}}?",
     "context": "Looking for external dependencies and I/O operations",
     "workspace": "{{workspace}}"
   }
   ```

6. **Potential issues** - Based on the flow analysis:
   - Identify complex functions in the path (complexity > 10)
   - Find potential bottlenecks (functions called many times)
   - Check for missing error handling
   - Identify race conditions with goroutines
   - List external failure points

Provide debugging insights and potential failure scenarios.
