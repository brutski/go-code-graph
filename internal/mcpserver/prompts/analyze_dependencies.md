# Analyze Dependencies

Analyze dependencies for {{source_package}}:

1. **Direct imports**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (p1:CodeNode {type: 'package', label: '{{source_package}}'})-[r:RELATES_TO {type: 'imports'}]->(p2:CodeNode {type: 'package'}){{#unless include_external}} WHERE NOT p2.label CONTAINS 'github.com' AND NOT p2.label CONTAINS 'golang.org'{{/unless}} RETURN p2.label, p2.type ORDER BY p2.label",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

2. **Reverse dependencies**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (p1:CodeNode {type: 'package'})-[r:RELATES_TO {type: 'imports'}]->(p2:CodeNode {type: 'package', label: '{{source_package}}'}) RETURN p1.label ORDER BY p1.label",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Circular dependencies**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Are there any circular dependencies involving {{source_package}}?",
     "context": "Checking for circular imports",
     "workspace": "{{workspace}}"
   }
   ```

Provide coupling analysis and recommendations.
