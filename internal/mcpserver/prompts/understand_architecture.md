# Understand Architecture

Help me understand the architecture{{#if component}} of {{component}}{{/if}}:

1. **Architectural layers**:

   ```json
   Use tool: detect_architecture
   Arguments: {
     "analysisType": "layers",
     "workspace": "{{workspace}}"
   }
   ```

2. **Component interactions**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "How do the main components interact{{#if component}}, specifically around {{component}}{{/if}}? Show me the key interfaces and their relationships",
     "context": "Understanding component architecture",
     "workspace": "{{workspace}}"
   }
   ```

3. **Main interfaces**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (i:CodeNode {type: 'interface'}){{#if component}} WHERE i.package CONTAINS '{{component}}'{{/if}} OPTIONAL MATCH (i)<-[r:RELATES_TO {type: 'implements'}]-(s) WITH i, count(s) as implCount RETURN i.label, i.package, implCount ORDER BY implCount DESC LIMIT 20",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

Provide architectural insights and design patterns used.
