# Onboard New Developer

Create onboarding guide{{#if focus_area}} for {{focus_area}}{{/if}}:

1. **Main entry points**:

   ```json
   Use tool: natural_query
   Arguments: {
     "question": "What are the main entry points{{#if focus_area}} for {{focus_area}}{{/if}}? Show me where to start understanding the code",
     "context": "Developer onboarding",
     "workspace": "{{workspace}}"
   }
   ```

2. **Key interfaces**:

   ```json
   Use tool: cypher_query
   Arguments: {
     "query": "MATCH (i:CodeNode {type: 'interface'}){{#if focus_area}} WHERE i.package CONTAINS '{{focus_area}}'{{/if}} OPTIONAL MATCH (i)<-[r:RELATES_TO {type: 'implements'}]-(s) WITH i, count(s) as implementations WHERE implementations > 0 RETURN i.label, i.package, implementations ORDER BY implementations DESC LIMIT 10",
     "parameters": {},
     "workspace": "{{workspace}}"
   }
   ```

3. **Important packages**:

   ```json
   Use tool: detect_architecture
   Arguments: {
     "analysisType": "layers",
     "workspace": "{{workspace}}"
   }
   ```

Provide structured learning path.
