MATCH (s:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'implements'}]->(i:CodeNode {workspace: $workspace})
WHERE (i.full_name = $name OR i.label = $name)
  AND s.type = 'struct'
  AND i.type = 'interface'
RETURN DISTINCT s.full_name as implementer, s.package as package, s.label as name
ORDER BY s.label
