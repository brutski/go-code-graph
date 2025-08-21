# Analyze Interface Usage

Analyze the {{interface_name}} interface:

1. **Find implementations**:

   ```json
   Use tool: find_implementers
   Arguments: {
     "interfaceName": "{{interface_name}}",
     "workspace": "{{workspace}}"
   }
   ```

2. **Usage as parameter**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'has_parameter'}]->(p:CodeNode) WHERE p.type_name = '{{interface_name}}' OR p.type_name ENDS WITH '.{{interface_name}}' RETURN DISTINCT f.label, f.package, f.type ORDER BY f.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Functions returning interface**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (f:CodeNode)-[r:RELATES_TO {type: 'returns'}]->(i:CodeNode) WHERE i.label = '{{interface_name}}' RETURN DISTINCT f.label, f.package, f.signature ORDER BY f.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

Provide design insights and improvement suggestions.
