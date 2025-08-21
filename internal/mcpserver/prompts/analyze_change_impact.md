# Analyze Change Impact

I need to {{change_type}} the {{component_name}}. Please analyze the impact:

1. **Direct impact analysis** - Find what will be immediately affected:

   ```json
   Use tool: analyze_impact
   Arguments: {
     "nodeId": "{{component_name}}",
     "changeType": "{{change_type}}",
     "maxDepth": 3,
     "workspace": "{{workspace}}"
   }
   ```

2. **Find all callers** - Identify who uses this component:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (caller:CodeNode)-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode) WHERE target.label = '{{component_name}}' OR target.full_name ENDS WITH '.{{component_name}}' RETURN DISTINCT caller.label, caller.package, caller.type ORDER BY caller.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Check interface impacts** - Find affected interfaces:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (n:CodeNode {label: '{{component_name}}'})-[r:RELATES_TO]-(i:CodeNode {type: 'interface'}) WHERE r.type IN ['implements', 'returns', 'parameter_type'] RETURN DISTINCT i.label, i.package, r.type ORDER BY i.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

4. **Find test files** - Identify tests that need updates:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "Find all test files that test {{component_name}} or call it directly",
     "context": "Looking for affected tests",
     "workspace": "{{workspace}}"
   }
   ```

5. **Check for type usage** - If changing a struct/interface:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (n:CodeNode)-[r:RELATES_TO]->(target:CodeNode) WHERE target.label = '{{component_name}}' AND r.type IN ['returns', 'parameter_type', 'field_type', 'embeds'] RETURN DISTINCT n.label, n.package, n.type, r.type ORDER BY n.package",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

6. **Migration strategy** - Based on the impact analysis:
   - List all components that need changes
   - Identify breaking changes
   - Suggest safe refactoring order
   - Recommend compatibility approach
   - Estimate effort and risk level

Provide a comprehensive risk assessment with specific migration steps.
