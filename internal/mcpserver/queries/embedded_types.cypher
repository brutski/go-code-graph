// Find all types that embed other types (composition analysis)
MATCH (embedder:CodeNode {workspace: $workspace})-[r:RELATES_TO {type: 'embeds'}]->(embedded:CodeNode {workspace: $workspace})
WHERE embedder.type IN ['struct', 'interface']
  AND embedded.type IN ['struct', 'interface']
RETURN DISTINCT embedder.full_name as embedderType,
       embedder.package as embedderPackage,
       collect(DISTINCT {
         embeddedType: embedded.full_name,
         embeddedPackage: embedded.package,
         embeddedKind: embedded.type
       }) as embeddedTypes,
       count(DISTINCT embedded) as embeddingCount
ORDER BY embeddingCount DESC, embedder.full_name