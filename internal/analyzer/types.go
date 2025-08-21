package analyzer

import (
	"crypto/sha256"
	"fmt"
	"go/token"
	"go/types"
	"time"
)

// Node types - Simplified to essential entities
const (
	// Core node types (95% of usage)
	NodeTypePackage   = "package"   // Package/module level
	NodeTypeStruct    = "struct"    // Struct types
	NodeTypeInterface = "interface" // Interface types
	NodeTypeFunction  = "function"  // Standalone functions
	NodeTypeMethod    = "method"    // Type methods

	// Secondary node types (keep for completeness)
	NodeTypeField     = "field"     // Struct fields
	NodeTypeParameter = "parameter" // Function parameters
	NodeTypeConstant  = "constant"  // Package constants
	NodeTypeVariable  = "variable"  // Package variables

	// Removed node types - never or rarely used
	// NodeTypeGeneric    = "generic"    // Never implemented
	// NodeTypeConstraint = "constraint" // Never implemented
	// NodeTypeAlias      = "alias"      // Never implemented
	// NodeTypeChannel    = "channel"    // Never implemented
	// NodeTypeGoroutine  = "goroutine"  // Never implemented
	// NodeTypeError      = "error"      // Never implemented
)

// Edge types - Simplified to essential relationships for 80/20 efficiency
const (
	// Essential edges (90% of queries)
	EdgeTypeCalls      = "calls"      // Function/method calls: "What calls this?", "What does this call?"
	EdgeTypeImplements = "implements" // Interface implementation: "What implements this interface?"
	EdgeTypeUses       = "uses"       // Type usage/dependency: "Where is this type used?"
	EdgeTypeImports    = "imports"    // Package dependencies: "What does this package depend on?"
	EdgeTypeHasMethod  = "has_method" // Method ownership: "What methods does this type have?"

	// Common analysis queries (70% of use cases)
	EdgeTypeReturns    = "returns"    // Return types: "What functions return this type?"
	EdgeTypeEmbeds     = "embeds"     // Composition: "What does this struct embed?"
	EdgeTypeConstructs = "constructs" // Factory patterns: "How is this type created?"

	// Specialized but important (30% of use cases)
	EdgeTypeHandlesError    = "handles_error"    // Error handling: "Does this handle errors?"
	EdgeTypeSpawnsGoroutine = "spawns_goroutine" // Concurrency: "What spawns goroutines?"
	EdgeTypeHasParameter    = "has_parameter"    // Parameters: "What functions accept this type?"

	// Legacy edge types - kept for backward compatibility but deprecated
	// TODO: Remove these in next major version
	EdgeTypeHasField      = "has_field"      // Use has_method pattern instead
	EdgeTypeMethodOf      = "method_of"      // Redundant with has_method
	EdgeTypeParameterType = "parameter_type" // Use has_parameter
	EdgeTypeWrapsError    = "wraps_error"    // Merged into handles_error

	// Removed edge types - too specialized, low usage
	// EdgeTypeTypeAsserts     = "type_asserts"      // < 5% usage
	// EdgeTypeSendsChannel    = "sends_channel"     // < 5% usage
	// EdgeTypeReceivesChannel = "receives_channel"  // < 5% usage
	// EdgeTypeInstantiates    = "instantiates"      // < 5% usage
	// EdgeTypeConstrains      = "constrains"        // Never used
	// EdgeTypePromotes        = "promotes"          // < 5% usage
	// EdgeTypeShadows         = "shadows"           // < 5% usage
	// EdgeTypeClosureCaptures = "closure_captures"  // < 5% usage
	// EdgeTypeDefers          = "defers"            // < 5% usage
	// EdgeTypePanics          = "panics"            // < 5% usage
	// EdgeTypeRecovers        = "recovers"          // < 5% usage
	// EdgeTypeCreates         = "creates"           // Duplicate of constructs
)

// Visibility constants
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Kind constants for TypeInfo
const (
	KindStruct    = "struct"
	KindInterface = "interface"
)

// EnhancedNode represents a comprehensive code entity
type EnhancedNode struct {
	ID              string                 `json:"id"`
	Label           string                 `json:"label"`
	Type            string                 `json:"type"`
	Package         string                 `json:"package"`
	FullName        string                 `json:"full_name"`
	Size            int                    `json:"size"`
	Level           int                    `json:"level"`
	Position        Position               `json:"position"`
	Signature       string                 `json:"signature,omitempty"`
	TypeInfo        TypeInfo               `json:"type_info,omitempty"`
	Visibility      string                 `json:"visibility"`
	Documentation   string                 `json:"documentation,omitempty"`
	Complexity      int                    `json:"complexity,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	Tags            []string               `json:"tags,omitempty"`
	Embedding       []float32              `json:"embedding,omitempty"`
	EmbeddingModel  string                 `json:"embedding_model,omitempty"`
	SemanticSummary string                 `json:"semantic_summary,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// EnhancedEdge represents a relationship with rich metadata
type EnhancedEdge struct {
	ID          string                 `json:"id"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Type        string                 `json:"type"`
	Weight      int                    `json:"weight"`
	Position    Position               `json:"position"`
	Context     string                 `json:"context,omitempty"`
	Conditional bool                   `json:"conditional,omitempty"` // if in conditional block
	LoopDepth   int                    `json:"loop_depth,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Position represents source code location
type Position struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Offset   int    `json:"offset"`
}

// NodeIdentifier provides standardized node ID generation
type NodeIdentifier struct {
	Type    string
	Package string
	Name    string
}

// ID generates a consistent node ID
func (n NodeIdentifier) ID() string {
	if n.Package == "" {
		return fmt.Sprintf("%s:%s", n.Type, n.Name)
	}
	return fmt.Sprintf("%s:%s.%s", n.Type, n.Package, n.Name)
}

// String returns a human-readable representation
func (n NodeIdentifier) String() string {
	return n.ID()
}

// TypeInfo contains Go type system information
type TypeInfo struct {
	Kind           string            `json:"kind"` // struct, interface, func, etc.
	IsPointer      bool              `json:"is_pointer"`
	IsSlice        bool              `json:"is_slice"`
	IsArray        bool              `json:"is_array"`
	IsChannel      bool              `json:"is_channel"`
	ChannelDir     string            `json:"channel_dir,omitempty"` // send, recv, both
	IsGeneric      bool              `json:"is_generic"`
	TypeParams     []TypeParam       `json:"type_params,omitempty"`
	Constraints    []string          `json:"constraints,omitempty"`
	UnderlyingType string            `json:"underlying_type,omitempty"`
	MethodSet      []MethodSignature `json:"method_set,omitempty"`
	FieldCount     int               `json:"field_count,omitempty"`
	IsExported     bool              `json:"is_exported"`
	IsAlias        bool              `json:"is_alias"`
	EmbeddedTypes  []string          `json:"embedded_types,omitempty"`
}

// TypeParam represents a generic type parameter
type TypeParam struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint"`
}

// MethodSignature represents a method signature
type MethodSignature struct {
	Name       string   `json:"name"`
	Params     []string `json:"params"`
	Returns    []string `json:"returns"`
	IsExported bool     `json:"is_exported"`
}

// ConcurrencyInfo tracks concurrency patterns
type ConcurrencyInfo struct {
	GoroutineSpawns  []GoroutineSpawn  `json:"goroutine_spawns,omitempty"`
	ChannelOps       []ChannelOp       `json:"channel_ops,omitempty"`
	SelectStatements []SelectStatement `json:"select_statements,omitempty"`
	Mutexes          []MutexUsage      `json:"mutexes,omitempty"`
}

// GoroutineSpawn represents a goroutine creation
type GoroutineSpawn struct {
	Position   Position `json:"position"`
	TargetFunc string   `json:"target_func"`
	Context    string   `json:"context"`
}

// ChannelOp represents channel operations
type ChannelOp struct {
	Position    Position `json:"position"`
	Operation   string   `json:"operation"` // send, receive, close, make
	ChannelType string   `json:"channel_type"`
	Direction   string   `json:"direction,omitempty"`
}

// SelectStatement represents select statement
type SelectStatement struct {
	Position   Position     `json:"position"`
	Cases      []SelectCase `json:"cases"`
	HasDefault bool         `json:"has_default"`
}

// SelectCase represents a case in select statement
type SelectCase struct {
	Position  Position `json:"position"`
	Operation string   `json:"operation"` // send, receive
	Channel   string   `json:"channel"`
}

// MutexUsage represents mutex usage
type MutexUsage struct {
	Position  Position `json:"position"`
	Type      string   `json:"type"`      // Mutex, RWMutex
	Operation string   `json:"operation"` // Lock, Unlock, RLock, RUnlock
}

// ErrorInfo tracks error handling patterns
type ErrorInfo struct {
	ErrorTypes    []ErrorType     `json:"error_types,omitempty"`
	ErrorHandling []ErrorHandling `json:"error_handling,omitempty"`
	PanicRecovery []PanicRecovery `json:"panic_recovery,omitempty"`
}

// ErrorType represents custom error types
type ErrorType struct {
	Name     string   `json:"name"`
	Package  string   `json:"package"`
	Position Position `json:"position"`
	Methods  []string `json:"methods"`
}

// ErrorHandling represents error handling patterns
type ErrorHandling struct {
	Position    Position `json:"position"`
	Pattern     string   `json:"pattern"` // check, wrap, ignore, propagate
	ErrorVar    string   `json:"error_var"`
	Wrapping    bool     `json:"wrapping"`
	WrapMessage string   `json:"wrap_message,omitempty"`
}

// PanicRecovery represents panic/recover usage
type PanicRecovery struct {
	Position Position `json:"position"`
	Type     string   `json:"type"` // panic, recover
	InDefer  bool     `json:"in_defer"`
	Message  string   `json:"message,omitempty"`
}

// Graph represents the complete enhanced code graph
type Graph struct {
	Nodes       []EnhancedNode  `json:"nodes"`
	Edges       []EnhancedEdge  `json:"edges"`
	Stats       GraphStats      `json:"stats"`
	Metadata    GraphMetadata   `json:"metadata"`
	Concurrency ConcurrencyInfo `json:"concurrency,omitempty"`
	ErrorInfo   ErrorInfo       `json:"error_info,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	Version     string          `json:"version"`
}

// GraphStats contains comprehensive statistics
type GraphStats struct {
	TotalNodes            int              `json:"total_nodes"`
	TotalEdges            int              `json:"total_edges"`
	NodesByType           map[string]int   `json:"nodes_by_type"`
	EdgesByType           map[string]int   `json:"edges_by_type"`
	PackageCount          int              `json:"package_count"`
	MaxDepth              int              `json:"max_depth"`
	Dependencies          []string         `json:"dependencies"`
	CyclomaticComplexity  int              `json:"cyclomatic_complexity"`
	CouplingMetrics       CouplingMetrics  `json:"coupling_metrics"`
	InterfaceCompliance   []InterfaceMatch `json:"interface_compliance,omitempty"`
	ArchitecturalPatterns []Pattern        `json:"architectural_patterns,omitempty"`
}

// CouplingMetrics represents coupling analysis
type CouplingMetrics struct {
	AfferentCoupling map[string]int     `json:"afferent_coupling"` // Ca - incoming dependencies
	EfferentCoupling map[string]int     `json:"efferent_coupling"` // Ce - outgoing dependencies
	Instability      map[string]float64 `json:"instability"`       // I = Ce / (Ca + Ce)
	Abstractness     map[string]float64 `json:"abstractness"`      // A = Abstract classes / Total classes
}

// InterfaceMatch represents interface satisfaction
type InterfaceMatch struct {
	Interface       string   `json:"interface"`
	Implementations []string `json:"implementations"`
	PartialMatches  []string `json:"partial_matches,omitempty"`
}

// Pattern represents detected architectural patterns
type Pattern struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`       // singleton, factory, observer, etc.
	Confidence  float64                `json:"confidence"` // 0.0 - 1.0
	Components  []string               `json:"components"` // node IDs involved
	Description string                 `json:"description"`
	Position    Position               `json:"position"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// GraphMetadata contains graph-level metadata
type GraphMetadata struct {
	SourcePath      string            `json:"source_path"`
	GoVersion       string            `json:"go_version"`
	ModulePath      string            `json:"module_path"`
	BuildTags       []string          `json:"build_tags,omitempty"`
	AnalysisOptions AnalysisOptions   `json:"analysis_options"`
	Performance     AnalysisPerf      `json:"performance"`
	Warnings        []AnalysisWarning `json:"warnings,omitempty"`
}

// AnalysisOptions represents analysis configuration
type AnalysisOptions struct {
	IncludeConcurrency bool     `json:"include_concurrency"`
	IncludeErrorFlow   bool     `json:"include_error_flow"`
	IncludePatterns    bool     `json:"include_patterns"`
	IncludeGenerics    bool     `json:"include_generics"`
	MaxDepth           int      `json:"max_depth"`
	ExcludePaths       []string `json:"exclude_paths,omitempty"`
	IncludeTests       bool     `json:"include_tests"`
}

// AnalysisPerf contains performance metrics
type AnalysisPerf struct {
	Duration      time.Duration `json:"duration"`
	FilesAnalyzed int           `json:"files_analyzed"`
	LinesOfCode   int           `json:"lines_of_code"`
	MemoryUsed    int64         `json:"memory_used"`
}

// AnalysisWarning represents analysis warnings
type AnalysisWarning struct {
	Type     string   `json:"type"`
	Message  string   `json:"message"`
	Position Position `json:"position"`
	Severity string   `json:"severity"`
}

// NewGraph creates a new enhanced graph
func NewGraph() *Graph {
	return &Graph{
		Nodes:     make([]EnhancedNode, 0),
		Edges:     make([]EnhancedEdge, 0),
		CreatedAt: time.Now(),
		Version:   "2.0.0",
		Stats: GraphStats{
			NodesByType:           make(map[string]int),
			EdgesByType:           make(map[string]int),
			Dependencies:          make([]string, 0),
			InterfaceCompliance:   make([]InterfaceMatch, 0),
			ArchitecturalPatterns: make([]Pattern, 0),
			CouplingMetrics: CouplingMetrics{
				AfferentCoupling: make(map[string]int),
				EfferentCoupling: make(map[string]int),
				Instability:      make(map[string]float64),
				Abstractness:     make(map[string]float64),
			},
		},
		Metadata: GraphMetadata{
			AnalysisOptions: AnalysisOptions{
				IncludeConcurrency: true,
				IncludeErrorFlow:   true,
				IncludePatterns:    true,
				IncludeGenerics:    true,
				MaxDepth:           10,
				IncludeTests:       false,
			},
			Warnings: make([]AnalysisWarning, 0),
		},
	}
}

// AddNode adds an enhanced node to the graph
func (g *Graph) AddNode(node EnhancedNode) {
	node.CreatedAt = time.Now()
	g.Nodes = append(g.Nodes, node)
	g.Stats.TotalNodes++
	g.Stats.NodesByType[node.Type]++
}

// AddEdge adds an enhanced edge to the graph
func (g *Graph) AddEdge(edge EnhancedEdge) {
	edge.CreatedAt = time.Now()
	edge.ID = fmt.Sprintf("%s->%s:%s", edge.Source, edge.Target, edge.Type)
	g.Edges = append(g.Edges, edge)
	g.Stats.TotalEdges++
	g.Stats.EdgesByType[edge.Type]++
}

// GetNodeByID finds a node by its ID
func (g *Graph) GetNodeByID(id string) *EnhancedNode {
	for i := range g.Nodes {
		if g.Nodes[i].ID == id {
			return &g.Nodes[i]
		}
	}
	return nil
}

// GenerateNodeID creates a unique node ID
// GenerateNodeID creates a consistent node ID using NodeIdentifier
func GenerateNodeID(nodeType, pkg, name string) string {
	return NodeIdentifier{
		Type:    nodeType,
		Package: pkg,
		Name:    name,
	}.ID()
}

// PositionFromToken converts token.Position to our Position
func PositionFromToken(fset *token.FileSet, pos token.Pos) Position {
	if !pos.IsValid() {
		return Position{}
	}
	position := fset.Position(pos)
	return Position{
		Filename: position.Filename,
		Line:     position.Line,
		Column:   position.Column,
		Offset:   position.Offset,
	}
}

// TypeInfoFromGoTypes extracts type information from go/types
func TypeInfoFromGoTypes(t types.Type) TypeInfo {
	info := TypeInfo{}

	switch typ := t.(type) {
	case *types.Named:
		extractNamedTypeInfo(&info, typ)
	case *types.Pointer:
		info.IsPointer = true
		info.Kind = "pointer"
	case *types.Slice:
		info.IsSlice = true
		info.Kind = "slice"
	case *types.Array:
		info.IsArray = true
		info.Kind = "array"
	case *types.Chan:
		extractChannelTypeInfo(&info, typ)
	case *types.Interface:
		extractInterfaceTypeInfo(&info, typ)
	case *types.Struct:
		extractStructTypeInfo(&info, typ)
	case *types.Signature:
		info.Kind = NodeTypeFunction
	case *types.Alias:
		info.IsAlias = true
		info.Kind = "alias"
	default:
		info.Kind = "basic"
	}

	return info
}

// extractNamedTypeInfo extracts information from named types
func extractNamedTypeInfo(info *TypeInfo, typ *types.Named) {
	info.Kind = "named"
	info.IsExported = typ.Obj().Exported()
	if underlying := typ.Underlying(); underlying != nil {
		info.UnderlyingType = underlying.String()
	}

	// Check for generic type parameters
	if typeParams := typ.TypeParams(); typeParams != nil && typeParams.Len() > 0 {
		info.IsGeneric = true
		for i := 0; i < typeParams.Len(); i++ {
			param := typeParams.At(i)
			info.TypeParams = append(info.TypeParams, TypeParam{
				Name:       param.String(),
				Constraint: param.Constraint().String(),
			})
		}
	}
}

// extractChannelTypeInfo extracts information from channel types
func extractChannelTypeInfo(info *TypeInfo, typ *types.Chan) {
	info.IsChannel = true
	info.Kind = "channel"
	switch typ.Dir() {
	case types.SendRecv:
		info.ChannelDir = "both"
	case types.SendOnly:
		info.ChannelDir = "send"
	case types.RecvOnly:
		info.ChannelDir = "recv"
	}
}

// extractInterfaceTypeInfo extracts information from interface types
func extractInterfaceTypeInfo(info *TypeInfo, typ *types.Interface) {
	info.Kind = KindInterface
	// Extract method signatures
	for i := 0; i < typ.NumMethods(); i++ {
		method := typ.Method(i)
		sig := method.Type().(*types.Signature)

		methodSig := MethodSignature{
			Name:       method.Name(),
			IsExported: method.Exported(),
		}

		// Extract parameters
		if params := sig.Params(); params != nil {
			for j := 0; j < params.Len(); j++ {
				methodSig.Params = append(methodSig.Params, params.At(j).Type().String())
			}
		}

		// Extract returns
		if results := sig.Results(); results != nil {
			for j := 0; j < results.Len(); j++ {
				methodSig.Returns = append(methodSig.Returns, results.At(j).Type().String())
			}
		}

		info.MethodSet = append(info.MethodSet, methodSig)
	}
}

// extractStructTypeInfo extracts information from struct types
func extractStructTypeInfo(info *TypeInfo, typ *types.Struct) {
	info.Kind = KindStruct
	info.FieldCount = typ.NumFields()

	// Extract embedded types
	for i := 0; i < typ.NumFields(); i++ {
		field := typ.Field(i)
		if field.Embedded() {
			info.EmbeddedTypes = append(info.EmbeddedTypes, field.Type().String())
		}
	}
}

// Legacy support for existing code
type (
	Node  = EnhancedNode
	Edge  = EnhancedEdge
	Stats = GraphStats
)

// nodeHasher handles hash computation for nodes
// This is internal to the v2 package and used by EmbeddingsGenerator
type nodeHasher struct{}

// computeContentHash creates a hash based on stable code properties
func (h *nodeHasher) computeContentHash(node *Node) string {
	// Only hash stable, code-based properties that indicate real changes
	data := fmt.Sprintf("%s|%s|%s|%d|%d|%s",
		node.Signature,         // Function signature
		node.TypeInfo.Kind,     // Type information
		node.Visibility,        // Public/private
		node.Size,              // Lines of code
		node.Complexity,        // Cyclomatic complexity
		node.Position.Filename, // File location
	)
	return h.hashString(data)
}

// computeSemanticHash creates a hash for semantic summary changes
func (h *nodeHasher) computeSemanticHash(semanticSummary string) string {
	return h.hashString(semanticSummary)
}

// hashString creates SHA256 hash of a string
func (h *nodeHasher) hashString(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// addHashMetadata adds hash metadata to a node
func (h *nodeHasher) addHashMetadata(node *Node) {
	if node.Metadata == nil {
		node.Metadata = make(map[string]interface{})
	}

	// Generate hashes
	contentHash := h.computeContentHash(node)
	semanticHash := h.computeSemanticHash(node.SemanticSummary)

	// Add metadata
	node.Metadata["content_hash"] = contentHash
	node.Metadata["semantic_hash"] = semanticHash
	node.Metadata["has_embedding"] = len(node.Embedding) > 0
	node.Metadata["last_analyzed"] = node.CreatedAt.Unix()

	if len(node.Embedding) > 0 {
		node.Metadata["embedding_generated_at"] = node.CreatedAt.Unix()
	}
}
