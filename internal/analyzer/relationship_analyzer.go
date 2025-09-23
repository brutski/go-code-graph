package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	nilIdent = "nil"
	errIdent = "err"
)

// RelationshipAnalyzer handles all relationship analysis for the code graph
type RelationshipAnalyzer struct {
	graph       *Graph
	nodeIndex   map[string]bool
	packages    map[string]*packages.Package
	nodeBuilder *NodeBuilder
}

// NewRelationshipAnalyzer creates a new relationship analyzer
func NewRelationshipAnalyzer(graph *Graph, nodeBuilder *NodeBuilder) *RelationshipAnalyzer {
	return &RelationshipAnalyzer{
		graph:       graph,
		nodeIndex:   make(map[string]bool),
		packages:    make(map[string]*packages.Package),
		nodeBuilder: nodeBuilder,
	}
}

// SetNodeIndex sets the node index for checking node existence
func (r *RelationshipAnalyzer) SetNodeIndex(index map[string]bool) {
	r.nodeIndex = index
}

// AddPackage adds a package for relationship analysis
func (r *RelationshipAnalyzer) AddPackage(pkg *packages.Package) {
	r.packages[pkg.PkgPath] = pkg
}

// AnalyzeAllRelationships analyzes all relationships in the registered packages
func (r *RelationshipAnalyzer) AnalyzeAllRelationships() {
	for _, pkg := range r.packages {
		r.analyzeImportRelationships(pkg)
		r.analyzeTypeRelationships(pkg)
		r.analyzeCallRelationships(pkg)
		r.analyzeFieldUsageRelationships(pkg)
		r.analyzeInterfaceImplementations(pkg)
		r.analyzeConstructorPatterns(pkg)
		r.analyzeConcurrencyPatterns(pkg)
		r.analyzeErrorHandling(pkg)
		// Removed rarely used analysis:
		// r.analyzeShadowing(pkg)        // Variable/field shadowing < 5% usage
		// r.analyzeClosureCaptures(pkg)  // Closure captures < 5% usage
	}
}

// analyzeImportRelationships creates edges for package imports
func (r *RelationshipAnalyzer) analyzeImportRelationships(pkg *packages.Package) {
	sourceNodeID := GenerateNodeID(NodeTypePackage, "", pkg.PkgPath)

	for _, imp := range pkg.Imports {
		targetNodeID := GenerateNodeID(NodeTypePackage, "", imp.PkgPath)
		if r.nodeIndex[targetNodeID] {
			edge := Edge{
				Source: sourceNodeID,
				Target: targetNodeID,
				Type:   EdgeTypeImports,
				Weight: 1,
			}
			r.graph.AddEdge(edge)
		}
	}
}

// analyzeTypeRelationships analyzes struct embedding and interface relationships
func (r *RelationshipAnalyzer) analyzeTypeRelationships(pkg *packages.Package) {
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			switch t := typeName.Type().Underlying().(type) {
			case *types.Struct:
				r.analyzeStructRelationships(pkg, typeName, t)
			case *types.Interface:
				r.analyzeInterfaceRelationships(pkg, typeName, t)
			}
		}
	}
}

// analyzeStructRelationships analyzes struct embedding and field relationships
func (r *RelationshipAnalyzer) analyzeStructRelationships(pkg *packages.Package, typeName *types.TypeName, s *types.Struct) {
	sourceNodeID := GenerateNodeID(NodeTypeStruct, pkg.PkgPath, typeName.Name())

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)

		// Handle embedded fields (struct embedding)
		if field.Embedded() {
			if named, ok := field.Type().(*types.Named); ok {
				targetPkg := ""
				if named.Obj().Pkg() != nil {
					targetPkg = named.Obj().Pkg().Path()
				}

				targetNodeID := GenerateNodeID(NodeTypeStruct, targetPkg, named.Obj().Name())
				if r.nodeIndex[targetNodeID] {
					edge := Edge{
						Source: sourceNodeID,
						Target: targetNodeID,
						Type:   EdgeTypeEmbeds,
						Weight: 3,
					}
					r.graph.AddEdge(edge)

					// Analyze method promotion
					r.analyzeMethodPromotion(pkg, sourceNodeID, named)
				}
			}
		}
	}
}

// analyzeInterfaceRelationships analyzes interface embedding
func (r *RelationshipAnalyzer) analyzeInterfaceRelationships(pkg *packages.Package, typeName *types.TypeName, iface *types.Interface) {
	sourceNodeID := GenerateNodeID(NodeTypeInterface, pkg.PkgPath, typeName.Name())

	for i := 0; i < iface.NumEmbeddeds(); i++ {
		embedded := iface.EmbeddedType(i)
		if named, ok := embedded.(*types.Named); ok {
			targetPkg := ""
			if named.Obj().Pkg() != nil {
				targetPkg = named.Obj().Pkg().Path()
			}

			targetNodeID := GenerateNodeID(NodeTypeInterface, targetPkg, named.Obj().Name())
			if r.nodeIndex[targetNodeID] {
				edge := Edge{
					Source: sourceNodeID,
					Target: targetNodeID,
					Type:   EdgeTypeEmbeds,
					Weight: 2,
				}
				r.graph.AddEdge(edge)
			}
		}
	}
}

// analyzeMethodPromotion creates edges for promoted methods through embedding
func (r *RelationshipAnalyzer) analyzeMethodPromotion(_ *packages.Package, _ string, embeddedType *types.Named) {
	// Get the method set of the embedded type
	// Use pointer type to get all methods (both value and pointer receivers)
	methodSet := types.NewMethodSet(types.NewPointer(embeddedType))

	for i := 0; i < methodSet.Len(); i++ {
		method := methodSet.At(i)
		methodObj := method.Obj()

		// Skip methods that aren't actually promoted (private methods)
		if !methodObj.Exported() {
			continue
		}

		// Method promotion edges removed - rarely used
		// // Create edge from embedder to the promoted method
		// targetPkg := ""
		// if methodObj.Pkg() != nil {
		// 	targetPkg = methodObj.Pkg().Path()
		// }
		// // Generate the method node ID
		// // Need to include the receiver type prefix to match how methods are generated
		// receiverName := embeddedType.Obj().Name()
		// sig := methodObj.Type().(*types.Signature)
		// recv := sig.Recv()
		//
		// // Determine receiver type string (pointer or value)
		// var recvTypeStr string
		// if _, ok := recv.Type().(*types.Pointer); ok {
		// 	recvTypeStr = "*" + targetPkg + "." + receiverName
		// } else {
		// 	recvTypeStr = targetPkg + "." + receiverName
		// }
		//
		// methodNodeID := GenerateNodeID(NodeTypeMethod, targetPkg, fmt.Sprintf("%s.%s", recvTypeStr, methodObj.Name()))
		//
		// if r.nodeIndex[methodNodeID] {
		// 	edge := Edge{
		// 		Source: embedderNodeID,
		// 		Target: methodNodeID,
		// 		Type:   EdgeTypePromotes,
		// 		Weight: 2,
		// 	}
		// 	r.graph.AddEdge(edge)
		// }
	}
}

// analyzeCallRelationships creates edges for function calls
func (r *RelationshipAnalyzer) analyzeCallRelationships(pkg *packages.Package) {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				r.analyzeCallsInFunction(pkg, fn)
				// Also analyze parameter and return types
				r.analyzeFunctionTypes(pkg, fn)
			}
			return true
		})
	}
}

// analyzeFunctionTypes analyzes parameter and return types for a function
func (r *RelationshipAnalyzer) analyzeFunctionTypes(pkg *packages.Package, fn *ast.FuncDecl) {
	if fn.Name == nil {
		return
	}

	funcID := r.getFunctionID(pkg, fn)
	if !r.nodeIndex[funcID] {
		return
	}

	// Analyze parameter types
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			if t, ok := pkg.TypesInfo.Types[param.Type]; ok {
				r.AnalyzeTypeUsage(pkg, funcID, t.Type, EdgeTypeParameterType)
			}
		}
	}

	// Analyze return types
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			if t, ok := pkg.TypesInfo.Types[result.Type]; ok {
				r.AnalyzeTypeUsage(pkg, funcID, t.Type, EdgeTypeReturns)
			}
		}
	}
}

// functionAnalysisVisitor encapsulates the state for analyzing a function's body
type functionAnalysisVisitor struct {
	r        *RelationshipAnalyzer
	pkg      *packages.Package
	callerID string
}

// visit processes a single AST node during function body traversal
func (v *functionAnalysisVisitor) visit(n ast.Node) bool {
	switch node := n.(type) {
	case *ast.CallExpr:
		v.handleCallExpression(node)
	case *ast.FuncLit:
		// Skip - CallExpr inside FuncLit will be handled by the outer ast.Inspect
		return true
	case *ast.CompositeLit:
		v.handleCompositeLiteral(node)
	case *ast.IndexExpr:
		// Generic instantiation edges removed - rarely used
	case *ast.BinaryExpr:
		v.handleBinaryExpr(node)
	case *ast.GoStmt:
		v.handleGoStatement(node)
	case *ast.DeferStmt:
		// Defer edges removed - rarely used
	case *ast.GenDecl:
		v.handleGenDecl(node)
	case *ast.TypeAssertExpr:
		// Type assertion edges removed - rarely used
	case *ast.TypeSwitchStmt:
		// Type switch analysis removed - rarely used
	}
	return true
}

// handleCallExpression handles regular function/method calls
func (v *functionAnalysisVisitor) handleCallExpression(node *ast.CallExpr) {
	v.r.analyzeCallExpression(v.pkg, v.callerID, node)
}

// handleCompositeLiteral handles composite literals like &Service{} or User{}
func (v *functionAnalysisVisitor) handleCompositeLiteral(node *ast.CompositeLit) {
	if t, ok := v.pkg.TypesInfo.Types[node]; ok {
		v.r.AnalyzeTypeUsage(v.pkg, v.callerID, t.Type, EdgeTypeUses)
	}
}

// handleBinaryExpr handles nil checks like u == nil or u != nil
func (v *functionAnalysisVisitor) handleBinaryExpr(node *ast.BinaryExpr) {
	if node.Op == token.EQL || node.Op == token.NEQ {
		// Check if one side is nil and the other is an identifier
		if ident, ok := node.X.(*ast.Ident); ok && v.r.isNil(node.Y) {
			// Check if the identifier refers to a parameter or variable of a named type
			if obj := v.pkg.TypesInfo.Uses[ident]; obj != nil {
				if t := obj.Type(); t != nil {
					v.r.AnalyzeTypeUsage(v.pkg, v.callerID, t, EdgeTypeUses)
				}
			}
		} else if ident, ok := node.Y.(*ast.Ident); ok && v.r.isNil(node.X) {
			// Check the other way around (nil == u)
			if obj := v.pkg.TypesInfo.Uses[ident]; obj != nil {
				if t := obj.Type(); t != nil {
					v.r.AnalyzeTypeUsage(v.pkg, v.callerID, t, EdgeTypeUses)
				}
			}
		}
	}
}

// handleGoStatement handles goroutine spawning
func (v *functionAnalysisVisitor) handleGoStatement(node *ast.GoStmt) {
	if calleeID := v.r.resolveCallTarget(v.pkg, node.Call); calleeID != "" {
		if v.r.nodeIndex[calleeID] {
			edge := Edge{
				Source: v.callerID,
				Target: calleeID,
				Type:   EdgeTypeSpawnsGoroutine,
				Weight: 2,
			}
			v.r.graph.AddEdge(edge)
		}
	}
}

// handleGenDecl handles variable declarations with explicit types
func (v *functionAnalysisVisitor) handleGenDecl(node *ast.GenDecl) {
	if node.Tok.String() == "var" {
		for _, spec := range node.Specs {
			if vspec, ok := spec.(*ast.ValueSpec); ok {
				if vspec.Type != nil {
					if t, ok := v.pkg.TypesInfo.Types[vspec.Type]; ok {
						v.r.AnalyzeTypeUsage(v.pkg, v.callerID, t.Type, EdgeTypeUses)
					}
				}
			}
		}
	}
}

// setupCallerID extracts the caller ID setup logic to reduce complexity
func (r *RelationshipAnalyzer) setupCallerID(pkg *packages.Package, fn *ast.FuncDecl) string {
	callerID := GenerateNodeID(NodeTypeFunction, pkg.PkgPath, fn.Name.Name)
	if fn.Recv != nil {
		// It's a method
		if t, ok := pkg.TypesInfo.Types[fn.Recv.List[0].Type]; ok {
			receiverType := types.TypeString(t.Type, nil)
			callerID = GenerateNodeID(NodeTypeMethod, pkg.PkgPath, receiverType+"."+fn.Name.Name)

			// Create method_of relationship
			r.AnalyzeTypeUsage(pkg, callerID, t.Type, EdgeTypeMethodOf)
		}
	}
	return callerID
}

// analyzeCallsInFunction analyzes function calls within a function
func (r *RelationshipAnalyzer) analyzeCallsInFunction(pkg *packages.Package, fn *ast.FuncDecl) {
	if fn.Name == nil || fn.Body == nil {
		return
	}

	callerID := r.setupCallerID(pkg, fn)
	if !r.nodeIndex[callerID] {
		return
	}

	// Create visitor and walk the function body
	visitor := &functionAnalysisVisitor{
		r:        r,
		pkg:      pkg,
		callerID: callerID,
	}
	ast.Inspect(fn.Body, visitor.visit)
}

// analyzeCallExpression analyzes a single call expression
func (r *RelationshipAnalyzer) analyzeCallExpression(pkg *packages.Package, callerID string, call *ast.CallExpr) {
	calleeID := r.resolveCallTarget(pkg, call)
	if calleeID != "" {
		if r.nodeIndex[calleeID] {
			edge := Edge{
				Source: callerID,
				Target: calleeID,
				Type:   EdgeTypeCalls,
				Weight: 1,
			}
			r.graph.AddEdge(edge)
		}
	}

	// Check for error handling patterns
	r.analyzeErrorHandlingInCall(pkg, callerID, call)
}

// resolveCallTarget resolves the target of a function call
func (r *RelationshipAnalyzer) resolveCallTarget(pkg *packages.Package, call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		// Direct function call
		if obj := pkg.TypesInfo.Uses[fn]; obj != nil {
			if fnObj, ok := obj.(*types.Func); ok {
				if fnObj.Pkg() != nil {
					return GenerateNodeID(NodeTypeFunction, fnObj.Pkg().Path(), fnObj.Name())
				}
			} else if varObj, ok := obj.(*types.Var); ok {
				// Handle function variables (e.g., var f = someFunc; f())
				if _, ok := varObj.Type().Underlying().(*types.Signature); ok {
					// This is a function-typed variable, but we can't resolve its target
					return ""
				}
			}
		}

	case *ast.SelectorExpr:
		// Method call or qualified function call
		if sel := pkg.TypesInfo.Selections[fn]; sel != nil {
			// Method call
			recv := sel.Recv()
			// Check if it's actually a method (could be a field holding a function)
			method, ok := sel.Obj().(*types.Func)
			if !ok {
				// This might be a field that holds a function value
				if varObj, ok := sel.Obj().(*types.Var); ok {
					// Try to resolve the type of the variable
					if _, ok := varObj.Type().Underlying().(*types.Signature); ok {
						// This is a function-typed field, but we can't resolve its target
						return ""
					}
				}
				return ""
			}

			if method.Pkg() != nil {
				receiverType := types.TypeString(recv, nil)
				switch recv.(type) {
				case *types.Named:
					if sig, _ := method.Type().(*types.Signature); sig != nil {
						if _, isPtr := deref(sig.Recv().Type()); isPtr {
							receiverType = "*" + receiverType
						}
					}
				}

				methodName := receiverType + "." + method.Name()
				return GenerateNodeID(NodeTypeMethod, method.Pkg().Path(), methodName)
			}
		} else if obj := pkg.TypesInfo.Uses[fn.Sel]; obj != nil {
			// Qualified function call
			if fnObj, ok := obj.(*types.Func); ok && fnObj.Pkg() != nil {
				return GenerateNodeID(NodeTypeFunction, fnObj.Pkg().Path(), fnObj.Name())
			}
		}
	}

	return ""
}

// analyzeFieldUsageRelationships analyzes field usage patterns
func (r *RelationshipAnalyzer) analyzeFieldUsageRelationships(pkg *packages.Package) {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if sel, ok := n.(*ast.SelectorExpr); ok {
				r.analyzeFieldAccess(pkg, sel)
			}
			return true
		})
	}
}

// analyzeFieldAccess analyzes a field access expression
func (r *RelationshipAnalyzer) analyzeFieldAccess(pkg *packages.Package, sel *ast.SelectorExpr) {
	if selection := pkg.TypesInfo.Selections[sel]; selection != nil {
		if selection.Kind() == types.FieldVal {
			// Find the struct type
			recv := selection.Recv()
			if named, ok := recv.(*types.Named); ok {
				if named.Obj().Pkg() != nil {
					structNodeID := GenerateNodeID(NodeTypeStruct,
						named.Obj().Pkg().Path(), named.Obj().Name())

					// Find the containing function
					funcID := r.findContainingFunction(pkg, sel)
					if funcID != "" && r.nodeIndex[funcID] && r.nodeIndex[structNodeID] {
						edge := Edge{
							Source: funcID,
							Target: structNodeID,
							Type:   EdgeTypeUses,
							Weight: 1,
						}
						r.graph.AddEdge(edge)
					}
				}
			}
		}
	}
}

// findContainingFunction finds the function containing a given node
func (r *RelationshipAnalyzer) findContainingFunction(pkg *packages.Package, targetNode ast.Node) string {
	// We need to walk through all files to find the containing function
	for _, file := range pkg.Syntax {
		var result string
		var currentFunc *ast.FuncDecl

		ast.Inspect(file, func(n ast.Node) bool {
			// Track current function
			if fn, ok := n.(*ast.FuncDecl); ok {
				currentFunc = fn
				return true
			}

			// Check if we found our target node
			if n == targetNode && currentFunc != nil {
				result = r.getFunctionID(pkg, currentFunc)
				return false // Stop searching
			}

			return true
		})

		if result != "" {
			return result
		}
	}

	return ""
}

// CreateImportEdges creates import relationship edges for a package
func (r *RelationshipAnalyzer) CreateImportEdges(pkg *packages.Package) {
	r.analyzeImportRelationships(pkg)
}

// AnalyzeTypeRelationships analyzes relationships for a specific type (public method for AST visitor)
func (r *RelationshipAnalyzer) AnalyzeTypeRelationships(pkg *packages.Package, typeName *types.TypeName) {
	switch t := typeName.Type().Underlying().(type) {
	case *types.Struct:
		r.analyzeStructRelationships(pkg, typeName, t)
	case *types.Interface:
		r.analyzeInterfaceRelationships(pkg, typeName, t)
	}
}

// AnalyzeConcurrencyPatterns analyzes concurrency patterns in a function (public method for AST visitor)
func (r *RelationshipAnalyzer) AnalyzeConcurrencyPatterns(pkg *packages.Package, fn *ast.FuncDecl) {
	if fn.Name == nil || fn.Body == nil {
		return
	}

	funcID := r.getFunctionID(pkg, fn)
	if !r.nodeIndex[funcID] {
		return
	}

	// Detect goroutine spawning
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GoStmt:
			if targetID := r.resolveCallTarget(pkg, node.Call); targetID != "" {
				if r.nodeIndex[targetID] {
					edge := Edge{
						Source: funcID,
						Target: targetID,
						Type:   EdgeTypeSpawnsGoroutine,
						Weight: 2,
					}
					r.graph.AddEdge(edge)
				}
			}
			// Channel operations removed - rarely used
			// case *ast.SendStmt: - removed (previously called analyzeChannelSend)
			// case *ast.UnaryExpr with token.ARROW: - removed (previously called analyzeChannelReceive)
		}
		return true
	})
}

// AnalyzeFunctionRelationships analyzes relationships for a function declaration
func (r *RelationshipAnalyzer) AnalyzeFunctionRelationships(pkg *packages.Package, fn *ast.FuncDecl, funcNode *Node) {
	if fn.Name == nil || fn.Body == nil {
		return
	}

	// Analyze function calls
	r.analyzeCallsInFunction(pkg, fn)

	// Analyze type usage in parameters and returns
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			if t, ok := pkg.TypesInfo.Types[param.Type]; ok {
				r.AnalyzeTypeUsage(pkg, funcNode.ID, t.Type, EdgeTypeParameterType)
			}
		}
	}

	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			if t, ok := pkg.TypesInfo.Types[result.Type]; ok {
				r.AnalyzeTypeUsage(pkg, funcNode.ID, t.Type, EdgeTypeReturns)
			}
		}
	}

	// For methods, create method_of relationship
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		if t, ok := pkg.TypesInfo.Types[fn.Recv.List[0].Type]; ok {
			r.AnalyzeTypeUsage(pkg, funcNode.ID, t.Type, EdgeTypeMethodOf)
		}
	}
}

// getFunctionID generates the appropriate node ID for a function or method
func (r *RelationshipAnalyzer) getFunctionID(pkg *packages.Package, fn *ast.FuncDecl) string {
	funcName := fn.Name.Name

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		// It's a method
		if t, ok := pkg.TypesInfo.Types[fn.Recv.List[0].Type]; ok {
			receiverType := types.TypeString(t.Type, nil)
			return GenerateNodeID(NodeTypeMethod, pkg.PkgPath, receiverType+"."+funcName)
		}
	}

	return GenerateNodeID(NodeTypeFunction, pkg.PkgPath, funcName)
}

// analyzeInterfaceImplementations checks which types implement interfaces
func (r *RelationshipAnalyzer) analyzeInterfaceImplementations(pkg *packages.Package) {
	// Collect all interfaces
	var interfaces []*types.Interface
	var interfaceNodes []string

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			if iface, ok := typeName.Type().Underlying().(*types.Interface); ok {
				interfaces = append(interfaces, iface)
				nodeID := GenerateNodeID(NodeTypeInterface, pkg.PkgPath, typeName.Name())
				interfaceNodes = append(interfaceNodes, nodeID)
			}
		}
	}

	// Check each type against interfaces
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			typ := typeName.Type()

			// Check against each interface
			for i, iface := range interfaces {
				if types.Implements(typ, iface) || types.Implements(types.NewPointer(typ), iface) {
					sourceNodeID := ""

					switch typ.Underlying().(type) {
					case *types.Struct:
						sourceNodeID = GenerateNodeID(NodeTypeStruct, pkg.PkgPath, typeName.Name())
					default:
						// Other types can also implement interfaces
						continue
					}

					if sourceNodeID != "" && r.nodeIndex[sourceNodeID] && r.nodeIndex[interfaceNodes[i]] {
						edge := Edge{
							Source: sourceNodeID,
							Target: interfaceNodes[i],
							Type:   EdgeTypeImplements,
							Weight: 2,
						}
						r.graph.AddEdge(edge)
					}
				}
			}
		}
	}
}

// analyzeConstructorPatterns detects constructor/factory patterns
func (r *RelationshipAnalyzer) analyzeConstructorPatterns(pkg *packages.Package) {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				// Skip methods - we want functions only
				if fn.Name == nil || fn.Recv != nil {
					return true
				}

				// Check if function name suggests constructor pattern
				funcName := fn.Name.Name
				if strings.HasPrefix(funcName, "New") || strings.HasPrefix(funcName, "Create") ||
					strings.HasPrefix(funcName, "Make") || strings.HasPrefix(funcName, "Build") {
					r.analyzeConstructorFunction(pkg, fn)
				}
			}
			return true
		})
	}
}

// analyzeConstructorFunction analyzes a potential constructor function
func (r *RelationshipAnalyzer) analyzeConstructorFunction(pkg *packages.Package, fn *ast.FuncDecl) {
	funcID := GenerateNodeID(NodeTypeFunction, pkg.PkgPath, fn.Name.Name)
	if !r.nodeIndex[funcID] {
		return
	}

	// Check return types for constructed types
	if fn.Type.Results != nil {
		for _, result := range fn.Type.Results.List {
			if t, ok := pkg.TypesInfo.Types[result.Type]; ok {
				// Handle pointer types
				targetType := t.Type
				if ptr, ok := targetType.(*types.Pointer); ok {
					targetType = ptr.Elem()
				}

				// Check if it's a named type (struct or interface)
				if named, ok := targetType.(*types.Named); ok {
					targetPkg := ""
					if named.Obj().Pkg() != nil {
						targetPkg = named.Obj().Pkg().Path()
					}

					// Determine target node type
					targetNodeType := ""
					switch named.Underlying().(type) {
					case *types.Struct:
						targetNodeType = NodeTypeStruct
					case *types.Interface:
						targetNodeType = NodeTypeInterface
					}

					if targetNodeType != "" {
						targetID := GenerateNodeID(targetNodeType, targetPkg, named.Obj().Name())
						if r.nodeIndex[targetID] {
							edge := Edge{
								Source: funcID,
								Target: targetID,
								Type:   EdgeTypeConstructs,
								Weight: 2,
							}
							r.graph.AddEdge(edge)
						}
					}
				}
			}
		}
	}
}

// analyzeConcurrencyPatterns analyzes goroutines and channel operations
func (r *RelationshipAnalyzer) analyzeConcurrencyPatterns(pkg *packages.Package) {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				r.AnalyzeConcurrencyPatterns(pkg, fn)
			}
			return true
		})
	}
}

// analyzeChannelSend analyzes channel send operations
// Channel send edges removed - rarely used
// func (r *RelationshipAnalyzer) analyzeChannelSend(pkg *packages.Package, send *ast.SendStmt) {
// 	if t := pkg.TypesInfo.TypeOf(send.Chan); t != nil {
// 		if chanType, ok := t.Underlying().(*types.Chan); ok {
// 			funcID := r.findContainingFunction(pkg, send)
// 			if funcID != "" {
// 				r.createChannelEdge(pkg, funcID, chanType, EdgeTypeSendsChannel)
// 			}
// 		}
// 	}
// }

// analyzeChannelReceive analyzes channel receive operations
// Channel receive edges removed - rarely used
// func (r *RelationshipAnalyzer) analyzeChannelReceive(pkg *packages.Package, recv *ast.UnaryExpr) {
// 	if t := pkg.TypesInfo.TypeOf(recv.X); t != nil {
// 		if chanType, ok := t.Underlying().(*types.Chan); ok {
// 			funcID := r.findContainingFunction(pkg, recv)
// 			if funcID != "" {
// 				r.createChannelEdge(pkg, funcID, chanType, EdgeTypeReceivesChannel)
// 			}
// 		}
// 	}
// }

// createChannelEdge creates an edge for channel operations
// Channel operation edges removed - rarely used
// func (r *RelationshipAnalyzer) createChannelEdge(_ *packages.Package, funcID string, chanType *types.Chan, edgeType string) {
// 	elemType := chanType.Elem()
//
// 	// Create a pseudo-target based on the channel element type
// 	var targetID string
// 	switch t := elemType.(type) {
// 	case *types.Named:
// 		// Named types (structs, interfaces)
// 		if t.Obj().Pkg() != nil {
// 			targetID = GenerateNodeID(NodeTypeStruct, t.Obj().Pkg().Path(), t.Obj().Name())
// 		}
// 	case *types.Basic:
// 		// Basic types like int, string, etc.
// 		targetID = GenerateNodeID(NodeTypeStruct, "builtin", "chan_"+t.Name())
// 	default:
// 		// For other types, create a generic channel edge
// 		targetID = GenerateNodeID(NodeTypeStruct, "builtin", "chan_unknown")
// 	}
//
// 	if targetID != "" {
// 		edge := Edge{
// 			Source: funcID,
// 			Target: targetID,
// 			Type:   edgeType,
// 			Weight: 1,
// 		}
// 		r.graph.AddEdge(edge)
// 	}
// }

// analyzeErrorHandling analyzes error handling patterns
func (r *RelationshipAnalyzer) analyzeErrorHandling(pkg *packages.Package) {
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				r.analyzeErrorHandlingInFunction(pkg, fn)
			}
			return true
		})
	}
}

// analyzeErrorHandlingInFunction detects error handling patterns in a function
func (r *RelationshipAnalyzer) analyzeErrorHandlingInFunction(pkg *packages.Package, fn *ast.FuncDecl) {
	if fn.Name == nil || fn.Body == nil {
		return
	}

	funcID := r.getFunctionID(pkg, fn)
	if !r.nodeIndex[funcID] {
		return
	}

	// Detect error handling patterns
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			// Check for: if err != nil
			r.detectErrorCheck(pkg, funcID, node)
		case *ast.CallExpr:
			// Check for error wrapping functions
			r.detectErrorWrapping(pkg, funcID, node)
		}
		return true
	})
}

// detectErrorCheck detects error checking patterns
func (r *RelationshipAnalyzer) detectErrorCheck(_ *packages.Package, funcID string, ifStmt *ast.IfStmt) {
	// Look for pattern: if err != nil
	if binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr); ok {
		if binExpr.Op == token.NEQ {
			// Check if left side is "err" and right side is "nil"
			if ident, ok := binExpr.X.(*ast.Ident); ok && ident.Name == errIdent {
				if ident, ok := binExpr.Y.(*ast.Ident); ok && ident.Name == nilIdent {
					// Found error checking pattern
					edge := Edge{
						Source: funcID,
						Target: funcID, // Self-edge for now
						Type:   EdgeTypeHandlesError,
						Weight: 1,
					}
					r.graph.AddEdge(edge)
				}
			}
		}
	}
}

// detectErrorWrapping detects error wrapping patterns
func (r *RelationshipAnalyzer) detectErrorWrapping(_ *packages.Package, funcID string, call *ast.CallExpr) {
	// Check for common error wrapping functions
	errorWrappers := []string{"fmt.Errorf", "errors.Wrap", "errors.WithMessage", "errors.WithStack"}

	callName := r.getCallName(call)
	for _, wrapper := range errorWrappers {
		if strings.HasSuffix(callName, wrapper) {
			edge := Edge{
				Source: funcID,
				Target: funcID, // Self-edge for now
				Type:   EdgeTypeWrapsError,
				Weight: 1,
			}
			r.graph.AddEdge(edge)
			break
		}
	}
}

// getCallName extracts the full name of a function call
func (r *RelationshipAnalyzer) getCallName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		if x, ok := fn.X.(*ast.Ident); ok {
			return x.Name + "." + fn.Sel.Name
		}
	}
	return ""
}

// analyzeErrorHandlingInCall checks for error handling in function calls
func (r *RelationshipAnalyzer) analyzeErrorHandlingInCall(_ *packages.Package, callerID string, call *ast.CallExpr) {
	// Check if it's an error wrapping call
	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "Errorf" || strings.HasSuffix(ident.Name, "Wrap") {
			edge := Edge{
				Source: callerID,
				Target: callerID, // Self-edge for now
				Type:   EdgeTypeWrapsError,
				Weight: 1,
			}
			r.graph.AddEdge(edge)
		}
	}
}

// Panic/recover edges removed - rarely used
// func (r *RelationshipAnalyzer) analyzePanicRecoverCall(pkg *packages.Package, callerID string, call *ast.CallExpr) {
// 	if ident, ok := call.Fun.(*ast.Ident); ok {
// 		// Check for built-in panic/recover by name
// 		switch ident.Name {
// 		case "panic":
// 			// Verify it's actually the builtin (not a user-defined function)
// 			if obj := pkg.TypesInfo.Uses[ident]; obj != nil {
// 				if _, isBuiltin := obj.(*types.Builtin); isBuiltin {
// 					edge := Edge{
// 						Source: callerID,
// 						Target: GenerateNodeID(NodeTypeFunction, "builtin", "panic"),
// 						Type:   EdgeTypePanics,
// 						Weight: 1,
// 					}
// 					r.graph.AddEdge(edge)
// 				}
// 			}
// 		case "recover":
// 			// Verify it's actually the builtin (not a user-defined function)
// 			if obj := pkg.TypesInfo.Uses[ident]; obj != nil {
// 				if _, isBuiltin := obj.(*types.Builtin); isBuiltin {
// 					edge := Edge{
// 						Source: callerID,
// 						Target: GenerateNodeID(NodeTypeFunction, "builtin", "recover"),
// 						Type:   EdgeTypeRecovers,
// 						Weight: 1,
// 					}
// 					r.graph.AddEdge(edge)
// 				}
// 			}
// 		}
// 	}
// }

// Closure capture analysis removed - rarely used
// func (r *RelationshipAnalyzer) analyzeClosureCaptures(pkg *packages.Package) {
// 	for _, file := range pkg.Syntax {
// 		ast.Inspect(file, func(n ast.Node) bool {
// 			switch node := n.(type) {
// 			case *ast.FuncLit:
// 				r.analyzeClosureCapturesInFuncLit(pkg, node)
// 			case *ast.GoStmt:
// 				if funcLit, ok := node.Call.Fun.(*ast.FuncLit); ok {
// 					r.analyzeClosureCapturesInFuncLit(pkg, funcLit)
// 				}
// 			}
// 			return true
// 		})
// 	}
// }

// Closure capture analysis removed - rarely used
// func (r *RelationshipAnalyzer) analyzeClosureCapturesInFuncLit(pkg *packages.Package, funcLit *ast.FuncLit) {
// 	containingFunc := r.findContainingFunction(pkg, funcLit)
// 	if containingFunc == "" {
// 		return
// 	}
//
// 	// Track captured variables
// 	capturedVars := make(map[string]bool)
//
// 	// Walk the function literal body to find variable uses
// 	ast.Inspect(funcLit.Body, func(n ast.Node) bool {
// 		if ident, ok := n.(*ast.Ident); ok {
// 			// Check if this is a use of a variable (not a definition)
// 			if obj := pkg.TypesInfo.Uses[ident]; obj != nil {
// 				if _, isVar := obj.(*types.Var); isVar {
// 					// Check if this variable is defined outside the closure
// 					if !r.isDefinedInScope(funcLit, obj) {
// 						capturedVars[ident.Name] = true
// 					}
// 				}
// 			}
// 		}
// 		return true
// 	})
//
// 	// Create closure capture edges
// 	for varName := range capturedVars {
// 		edge := Edge{
// 			Source: containingFunc,
// 			Target: containingFunc,
// 			Type:   EdgeTypeClosureCaptures,
// 			Weight: 1,
// 			Metadata: map[string]any{
// 				"captured_variable": varName,
// 			},
// 		}
// 		r.graph.AddEdge(edge)
// 	}
// }

// isDefinedInScope removed - unused after closure capture removal

// isNil checks if an expression is the nil identifier
func (r *RelationshipAnalyzer) isNil(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "nil"
	}
	return false
}

// AnalyzeTypeUsage analyzes type usage relationships
func (r *RelationshipAnalyzer) AnalyzeTypeUsage(pkg *packages.Package, sourceID string, typ types.Type, edgeType string) {
	switch t := typ.(type) {
	case *types.Named:
		// Named types (structs, interfaces, type aliases)
		targetPkg := ""
		if t.Obj().Pkg() != nil {
			targetPkg = t.Obj().Pkg().Path()
		}

		// Special handling for builtin error type
		if targetPkg == "" && t.Obj().Name() == "error" {
			targetID := GenerateNodeID(NodeTypeInterface, "builtin", "error")
			if r.nodeIndex[targetID] {
				edge := Edge{
					Source: sourceID,
					Target: targetID,
					Type:   edgeType,
					Weight: 1,
				}
				r.graph.AddEdge(edge)
			}
			return
		}

		// Determine node type
		targetID := ""
		switch t.Underlying().(type) {
		case *types.Struct:
			targetID = GenerateNodeID(NodeTypeStruct, targetPkg, t.Obj().Name())
		case *types.Interface:
			targetID = GenerateNodeID(NodeTypeInterface, targetPkg, t.Obj().Name())
		}

		if targetID != "" && r.nodeIndex[targetID] {
			edge := Edge{
				Source: sourceID,
				Target: targetID,
				Type:   edgeType,
				Weight: 1,
			}
			r.graph.AddEdge(edge)
		}
	case *types.Basic:
		// For basic types like string, int, etc., create edge with type name
		targetID := GenerateNodeID(NodeTypeStruct, "builtin", t.Name())
		if r.nodeIndex[targetID] {
			edge := Edge{
				Source: sourceID,
				Target: targetID,
				Type:   edgeType,
				Weight: 1,
			}
			r.graph.AddEdge(edge)
		}
	case *types.Pointer:
		// For pointer types, analyze the element type
		r.AnalyzeTypeUsage(pkg, sourceID, t.Elem(), edgeType)
	}
}

// Shadow analysis removed - rarely used
// func (r *RelationshipAnalyzer) analyzeShadowing(pkg *packages.Package) {
// 	// Analyze variable shadowing in each file
// 	for _, file := range pkg.Syntax {
// 		// Analyze variable shadowing
// 		r.analyzeVariableShadowing(pkg, file)
//
// 		// Analyze field shadowing through embedding
// 		r.analyzeFieldShadowing(pkg, file)
// 	}
// }

// analyzeVariableShadowing removed - unused after shadow detection removal

// analyzeBlockShadowing removed - unused after shadow detection removal

// analyzeFuncLitShadowing removed - unused after shadow detection removal

// findOuterScopeVariable removed - unused after shadow detection removal

// analyzeFieldShadowing removed - unused after shadow detection removal

// checkStructFieldShadowing removed - unused after shadow detection removal

// deref safely dereferences a type and returns whether it was a pointer
func deref(typ types.Type) (types.Type, bool) {
	if p, _ := types.Unalias(typ).(*types.Pointer); p != nil {
		// p.base should never be nil, but be conservative
		if p.Elem() == nil {
			return types.Typ[types.Invalid], true
		}
		return p.Elem(), true
	}
	return typ, false
}
