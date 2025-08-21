MATCH (n:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'calls'}]->(target:CodeNode {workspace: $workspace})
WITH target, count(r) as usageCount
WHERE usageCount > 1
RETURN DISTINCT target.label as targetName,
       target.package as package,
       target.type as nodeType,
       usageCount
ORDER BY usageCount DESC
