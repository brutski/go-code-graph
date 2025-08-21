# Find Performance Issues

Find potential performance issues{{#if focus_area}} in {{focus_area}}{{/if}}:

1. **Complex functions**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode) WHERE f.type IN ['function', 'method']{{#if focus_area}} AND f.package CONTAINS '{{focus_area}}'{{/if}} AND f.complexity > 15 RETURN f.label, f.package, f.complexity ORDER BY f.complexity DESC LIMIT 20",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

2. **Frequently called functions**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode)<-[r:RELATES_TO {type: 'calls'}]-(caller) WHERE f.type IN ['function', 'method']{{#if focus_area}} AND f.package CONTAINS '{{focus_area}}'{{/if}} WITH f, count(r) as callCount WHERE callCount > 10 RETURN f.label, f.package, f.complexity, callCount ORDER BY (f.complexity * callCount) DESC LIMIT 20", 
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Recursive calls**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'calls'}*2..5]->(f) WHERE f.type IN ['function', 'method']{{#if focus_area}} AND f.package CONTAINS '{{focus_area}}'{{/if}} RETURN DISTINCT f.label, f.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

Provide optimization suggestions for each issue.
