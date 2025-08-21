// Find all functions that construct specific types
MATCH (f:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'constructs'}]->(t:CodeNode {workspace: $workspace})
WHERE f.type IN ['function', 'method']
  AND t.type IN ['struct', 'interface']
RETURN DISTINCT t.full_name as constructedType,
       t.package as typePackage,
       collect(DISTINCT {
         constructor: f.full_name,
         constructorType: f.type,
         constructorPackage: f.package
       }) as constructors,
       count(DISTINCT f) as constructorCount
ORDER BY constructorCount DESC, t.full_name