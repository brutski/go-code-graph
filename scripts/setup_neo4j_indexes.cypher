// Neo4j Vector Index Setup for Code Graph Analysis
// This script creates the necessary vector indexes for efficient semantic search
// Run this script in Neo4j Browser or using cypher-shell

// 1. Create vector index for CodeNode embeddings (Titan V2 embeddings are 1024-dimensional)
CREATE VECTOR INDEX code_embeddings_index IF NOT EXISTS
FOR (n:CodeNode) 
ON n.embedding
OPTIONS {
  indexConfig: {
    `vector.dimensions`: 1024,
    `vector.similarity_function`: 'cosine'
  }
};

// 2. Create composite index for efficient filtering during hybrid search
CREATE INDEX code_type_package_index IF NOT EXISTS
FOR (n:CodeNode) 
ON (n.type, n.package);

// 3. Create index for full-text search on labels and names
CREATE FULLTEXT INDEX code_fulltext_index IF NOT EXISTS
FOR (n:CodeNode)
ON EACH [n.label, n.full_name, n.semantic_summary];

// 4. Create index for complexity-based filtering
CREATE INDEX code_complexity_index IF NOT EXISTS
FOR (n:CodeNode)
ON n.complexity;

// 5. Create index for relationship types for efficient traversal
CREATE INDEX relationship_type_index IF NOT EXISTS
FOR ()-[r:RELATES_TO]-()
ON r.type;

// 6. Create constraint to ensure unique node IDs
CREATE CONSTRAINT unique_node_id IF NOT EXISTS
FOR (n:CodeNode)
REQUIRE n.id IS UNIQUE;

// Display created indexes
SHOW INDEXES;
