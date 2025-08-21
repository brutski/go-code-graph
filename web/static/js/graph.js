// Global variables
let cy;
let graphData = null;
let allNodes = [];
let allEdges = [];
let currentSearchResults = [];
let selectedNode = null;

// Performance settings
let performanceMode = false;
let webglEnabled = false;
let initialLoadComplete = false;

console.log('✅ graph.js loaded successfully');

// Default filter settings for large graphs
const DEFAULT_PERFORMANCE_FILTER = {
    hideFields: true,
    hideParameters: true,
    minComplexity: 2,
    showOnlyPackagesAndTypes: true
};

// Initialize the graph visualization
document.addEventListener('DOMContentLoaded', function() {
    console.log('📄 DOMContentLoaded event fired');
    
    // Check if Cytoscape is loaded
    if (typeof cytoscape === 'undefined') {
        console.error('❌ Cytoscape library not loaded!');
        document.getElementById('cy').innerHTML = '<div class="loading" style="color: #dc3545;">Error: Cytoscape library failed to load. Please check your internet connection.</div>';
        return;
    }
    console.log('✅ Cytoscape library loaded');
    
    initializeGraph();
    setupEventListeners();
    setupPanelTabs();
});

async function initializeGraph() {
    try {
        console.log('🚀 Starting graph initialization...');
        
        // Show loading with performance info
        document.getElementById('cy').innerHTML = '<div class="loading"><div class="loading-spinner"></div>Loading enhanced graph data...</div>';
        
        console.log('📡 Fetching graph data from /api/graph...');
        
        // Fetch graph data from API
        const response = await fetch('/api/graph');
        console.log('📡 Response status:', response.status, response.statusText);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        console.log('🔍 Raw data received:', {
            dataKeys: Object.keys(data),
            hasGraph: !!data.graph,
            dataStructure: data.graph ? {
                hasNodes: !!data.graph.nodes,
                hasEdges: !!data.graph.edges,
                nodeCount: data.graph.nodes ? data.graph.nodes.length : 0,
                edgeCount: data.graph.edges ? data.graph.edges.length : 0
            } : 'No graph property'
        });
        
        graphData = data;
        
        // Check data structure and provide fallback
        if (!data.graph) {
            console.error('❌ No "graph" property found in data. Available keys:', Object.keys(data));
            // Try to find nodes and edges at root level
            if (data.nodes && data.edges) {
                console.log('✅ Found nodes/edges at root level, using those');
                data.graph = { nodes: data.nodes, edges: data.edges };
            } else {
                throw new Error('Invalid data structure: no graph.nodes or graph.edges found');
            }
        }
        
        if (!data.graph.nodes || !Array.isArray(data.graph.nodes)) {
            throw new Error('Invalid data structure: graph.nodes is not an array');
        }
        
        if (!data.graph.edges || !Array.isArray(data.graph.edges)) {
            throw new Error('Invalid data structure: graph.edges is not an array');
        }
        
        console.log('📊 Data validation passed:', {
            nodeCount: data.graph.nodes.length,
            edgeCount: data.graph.edges.length
        });
        
        // Transform data for Cytoscape with enhanced information
        allNodes = data.graph.nodes.map(node => ({
            data: {
                ...node,
                complexity: node.complexity || 1,
                visibility: node.visibility || 'public',
                is_exported: node.is_exported !== false,
                documentation: node.documentation || '',
                signature: node.signature || '',
                type_info: node.type_info || {},
                position_info: node.position || {}
            }
        }));
        
        // Create a set of valid node IDs for quick lookup
        const validNodeIds = new Set(allNodes.map(node => node.data.id));
        
        // Map edges and filter out those referencing external packages
        const unfilteredEdges = data.graph.edges.map(edge => ({
            data: {
                id: edge.id || `${edge.source}-${edge.target}`,
                source: edge.source,
                target: edge.target,
                type: edge.type,
                weight: edge.weight || 1,
                context: edge.context || '',
                conditional: edge.conditional || false
            }
        }));
        
        // Filter edges to only include those where both source and target exist
        allEdges = unfilteredEdges.filter(edge => {
            const hasSource = validNodeIds.has(edge.data.source);
            const hasTarget = validNodeIds.has(edge.data.target);
            
            if (!hasSource || !hasTarget) {
                // Log filtered edges for debugging (this should be rare with --include-packages)
                console.log(`⚠️ Filtering out edge with missing node(s) (use --include-packages to analyze external dependencies):`, {
                    edgeId: edge.data.id,
                    source: edge.data.source,
                    target: edge.data.target,
                    hasSource,
                    hasTarget,
                    type: edge.data.type
                });
                return false;
            }
            return true;
        });
        
        const filteredCount = unfilteredEdges.length - allEdges.length;
        if (filteredCount > 0) {
            console.log(`🔧 Filtered out ${filteredCount} edges referencing external packages`);
        }
        
        // Performance check - but don't filter for now
        const nodeCount = allNodes.length;
        const edgeCount = allEdges.length;
        performanceMode = false; // Disable performance mode completely for now
        
        console.log(`📈 Graph size: ${nodeCount} nodes, ${edgeCount} edges`);
        console.log('🔧 Using all nodes and filtered edges');
        
        // Use all nodes and filtered edges
        const initialNodes = allNodes;
        const initialEdges = allEdges;
        
        // Clear loading message
        document.getElementById('cy').innerHTML = '';
        
        console.log('⚙️ Computing layout on client...');
        const layoutOptions = getOptimizedLayoutOptions(performanceMode, nodeCount);
        
        console.log('🔧 Initializing Cytoscape with:', {
            nodeCount: initialNodes.length,
            edgeCount: initialEdges.length,
            sampleNode: initialNodes[0],
            sampleEdge: initialEdges[0],
            hasConfig: !!data.config
        });
        
        try {
            cy = cytoscape({
                container: document.getElementById('cy'),
                
                elements: {
                    nodes: initialNodes,
                    edges: initialEdges
                },
                
                style: generateCytoscapeStyle(data.config || {}),
                layout: layoutOptions,
                
                // Performance optimizations
                pixelRatio: performanceMode ? 1 : 'auto',
                
                // Disable expensive features for large graphs
                autolock: performanceMode,
                autoungrabify: performanceMode,
                boxSelectionEnabled: !performanceMode
            });
        } catch (cytoscapeError) {
            console.error('❌ Cytoscape initialization error:', cytoscapeError);
            throw new Error(`Cytoscape failed to initialize: ${cytoscapeError.message}`);
        }
        
        // Add performance warning if needed
        if (performanceMode) {
            addPerformanceWarning(nodeCount, edgeCount);
            // Set up performance-friendly controls
            setupPerformanceControls();
        }
        
        setupGraphEvents();
        initializeAnalytics();
        
        // Set proper initial zoom and viewport for massive graphs
        if (nodeCount > 5000) {
            console.log('🔍 Setting initial zoom for massive graph');
            // Set a reasonable initial zoom instead of trying to fit everything
            cy.zoom(0.5);
            cy.center();
            
            // Add zoom controls for large graphs
            addZoomControls();
        } else if (performanceMode) {
            // For large graphs, zoom out a bit but still show structure
            cy.zoom(0.8);
            cy.center();
        }
        
        initialLoadComplete = true;
        console.log(`✅ Enhanced graph initialized with ${initialNodes.length}/${nodeCount} nodes and ${initialEdges.length}/${edgeCount} edges`);
        
        // Show performance info
        if (performanceMode) {
            showPerformanceInfo(nodeCount, edgeCount, initialNodes.length, initialEdges.length);
        }
        
    } catch (error) {
        console.error('❌ Error loading graph data:', error);
        console.error('Error details:', {
            message: error.message,
            stack: error.stack
        });
        
        // More detailed error message
        const errorDetails = error.message.includes('Invalid data structure') 
            ? `<br><small>Data structure issue: ${error.message}</small>`
            : '';
            
        document.getElementById('cy').innerHTML = `
            <div class="loading" style="color: #dc3545;">
                <strong>Error loading graph data</strong>${errorDetails}
                <br><small>Check browser console for details</small>
            </div>
        `;
    }
}


function applyPerformanceFiltering(nodes, edges) {
    if (!performanceMode) {
        console.log('🔄 No performance filtering needed');
        return { initialNodes: nodes, initialEdges: edges };
    }
    
    console.log('🔄 Applying performance filtering...');
    
    // More intelligent filtering for better visualization
    const priorityTypes = ['package', 'struct', 'interface'];
    const secondaryTypes = ['function', 'method'];
    const hiddenTypes = ['field', 'parameter', 'variable', 'constant']; // Hide these initially
    
    // Filter nodes based on type and importance
    const filteredNodes = nodes.filter(node => {
        const type = node.data.type;
        
        // Always show architectural types (packages, structs, interfaces)
        if (priorityTypes.includes(type)) {
            return true;
        }
        
        // Hide low-priority types initially
        if (hiddenTypes.includes(type)) {
            return false;
        }
        
        // Show high-complexity or exported functions/methods only
        if (secondaryTypes.includes(type)) {
            const isExported = node.data.is_exported !== false;
            const isComplex = (node.data.complexity || 1) >= 3;
            return isExported && isComplex;
        }
        
        return false;
    });
    
    console.log('🔄 Node filtering results:', {
        original: nodes.length,
        filtered: filteredNodes.length,
        typeBreakdown: filteredNodes.reduce((acc, node) => {
            const type = node.data.type;
            acc[type] = (acc[type] || 0) + 1;
            return acc;
        }, {})
    });
    
    // Filter edges to show meaningful relationships
    const visibleNodeIds = new Set(filteredNodes.map(n => n.data.id));
    const meaningfulEdgeTypes = ['imports', 'implements', 'embeds', 'calls', 'has_method', 'constructs'];
    
    const filteredEdges = edges.filter(edge => {
        // Both nodes must be visible
        if (!visibleNodeIds.has(edge.data.source) || !visibleNodeIds.has(edge.data.target)) {
            return false;
        }
        
        // Show only meaningful edge types
        return meaningfulEdgeTypes.includes(edge.data.type);
    });
    
    console.log('🔄 Edge filtering results:', {
        original: edges.length,
        filtered: filteredEdges.length,
        edgeTypes: filteredEdges.reduce((acc, edge) => {
            const type = edge.data.type;
            acc[type] = (acc[type] || 0) + 1;
            return acc;
        }, {})
    });
    
    return { 
        initialNodes: filteredNodes, 
        initialEdges: filteredEdges 
    };
}

function getOptimizedLayoutOptions(performanceMode, nodeCount) {
    // For massive graphs (>5000 nodes), use COSE layout (works better than concentric)
    if (nodeCount > 5000) {
        console.log(`🎯 Using COSE layout for massive graph (${nodeCount} nodes)`);
        return {
            name: 'cose',
            animate: false,
            fit: false, // Let user control zoom manually
            padding: 100,
            nodeRepulsion: function(node) {
                // Different repulsion for different node types
                const type = node.data('type');
                if (type === 'package') return 20000;
                if (type === 'struct' || type === 'interface') return 15000;
                return 10000;
            },
            nodeOverlap: 50,
            idealEdgeLength: function(edge) {
                // Longer edges for important relationships
                const type = edge.data('type');
                if (type === 'imports') return 150;
                if (type === 'implements' || type === 'embeds') return 130;
                return 100;
            },
            edgeElasticity: 16,
            nestingFactor: 2,
            gravity: 20,
            numIter: 300, // Reduced for faster loading on massive graphs
            initialTemp: 80,
            coolingFactor: 0.9,
            minTemp: 5.0,
            randomize: true,
            avoidOverlap: true
        };
    } else if (performanceMode) {
        // Better layout for large graphs - use cose with reduced iterations
        return {
            name: 'cose',
            animate: false,
            fit: false, // Let user control zoom manually
            padding: 50,
            nodeRepulsion: function(node) {
                // Different repulsion for different node types
                const type = node.data('type');
                if (type === 'package') return 20000;
                if (type === 'struct' || type === 'interface') return 15000;
                return 10000;
            },
            nodeOverlap: 30,
            idealEdgeLength: function(edge) {
                // Longer edges for important relationships
                const type = edge.data('type');
                if (type === 'imports') return 200;
                if (type === 'implements' || type === 'embeds') return 180;
                return 120;
            },
            edgeElasticity: 16,
            nestingFactor: 2,
            gravity: 30,
            numIter: 150, // Further reduced for massive graphs
            initialTemp: 50,
            coolingFactor: 0.85,
            minTemp: 3.0,
            randomize: true,
            avoidOverlap: true
        };
    } else {
        // Full-featured layout for smaller graphs
        return {
            name: 'cose',
            animate: true,
            animationDuration: 1000,
            fit: true,
            padding: 30,
            nodeRepulsion: 8000,
            nodeOverlap: 10,
            idealEdgeLength: 100,
            edgeElasticity: 32,
            nestingFactor: 5,
            gravity: 80,
            numIter: 1000,
            initialTemp: 200,
            coolingFactor: 0.95,
            minTemp: 1.0
        };
    }
}

function addPerformanceWarning(nodeCount, edgeCount) {
    const warningDiv = document.createElement('div');
    warningDiv.className = 'performance-warning';
    warningDiv.style.cssText = `
        position: absolute;
        top: 10px;
        right: 10px;
        background: #fff3cd;
        border: 1px solid #ffeaa7;
        border-radius: 4px;
        padding: 8px 12px;
        font-size: 12px;
        z-index: 1000;
        max-width: 300px;
    `;
    warningDiv.innerHTML = `
        <strong>🚀 Performance Mode</strong><br>
        Large graph detected (${nodeCount.toLocaleString()} nodes)<br>
        Showing simplified view for better performance
    `;
    
    document.querySelector('.viz-content').appendChild(warningDiv);
    
    // Auto-hide after 5 seconds
    setTimeout(() => {
        warningDiv.style.display = 'none';
    }, 5000);
}

function setupPerformanceControls() {
    // Add performance toggle buttons
    const controlsDiv = document.createElement('div');
    controlsDiv.className = 'performance-controls';
    controlsDiv.style.cssText = `
        position: absolute;
        bottom: 10px;
        left: 10px;
        display: flex;
        gap: 8px;
        z-index: 1000;
    `;
    
    // Show all nodes button
    const showAllBtn = document.createElement('button');
    showAllBtn.textContent = '👁️ Show All';
    showAllBtn.className = 'btn btn-sm';
    showAllBtn.onclick = () => showAllNodes();
    
    // Show fields toggle
    const showFieldsBtn = document.createElement('button');
    showFieldsBtn.textContent = '🔗 Fields';
    showFieldsBtn.className = 'btn btn-sm';
    showFieldsBtn.onclick = () => toggleFields();
    
    controlsDiv.appendChild(showAllBtn);
    controlsDiv.appendChild(showFieldsBtn);
    document.querySelector('.viz-content').appendChild(controlsDiv);
}

function addZoomControls() {
    // Add zoom control buttons for massive graphs
    const zoomControlsDiv = document.createElement('div');
    zoomControlsDiv.className = 'zoom-controls';
    zoomControlsDiv.style.cssText = `
        position: absolute;
        top: 60px;
        right: 10px;
        display: flex;
        flex-direction: column;
        gap: 4px;
        z-index: 1000;
    `;
    
    // Zoom in button
    const zoomInBtn = document.createElement('button');
    zoomInBtn.textContent = '🔍+';
    zoomInBtn.className = 'btn btn-sm';
    zoomInBtn.title = 'Zoom In';
    zoomInBtn.onclick = () => {
        if (cy) {
            cy.zoom(cy.zoom() * 1.2);
        }
    };
    
    // Zoom out button  
    const zoomOutBtn = document.createElement('button');
    zoomOutBtn.textContent = '🔍-';
    zoomOutBtn.className = 'btn btn-sm';
    zoomOutBtn.title = 'Zoom Out';
    zoomOutBtn.onclick = () => {
        if (cy) {
            cy.zoom(cy.zoom() * 0.8);
        }
    };
    
    // Fit to screen button
    const fitBtn = document.createElement('button');
    fitBtn.textContent = '📐';
    fitBtn.className = 'btn btn-sm';
    fitBtn.title = 'Fit to Screen';
    fitBtn.onclick = () => {
        if (cy) {
            cy.fit(cy.nodes(':visible'), 30);
        }
    };
    
    // Reset view button
    const resetBtn = document.createElement('button');
    resetBtn.textContent = '🎯';
    resetBtn.className = 'btn btn-sm';
    resetBtn.title = 'Reset View';
    resetBtn.onclick = () => {
        if (cy) {
            cy.zoom(1);
            cy.center();
        }
    };
    
    zoomControlsDiv.appendChild(zoomInBtn);
    zoomControlsDiv.appendChild(zoomOutBtn);
    zoomControlsDiv.appendChild(fitBtn);
    zoomControlsDiv.appendChild(resetBtn);
    
    document.querySelector('.viz-content').appendChild(zoomControlsDiv);
}

function showPerformanceInfo(totalNodes, totalEdges, visibleNodes, visibleEdges) {
    const hiddenNodes = totalNodes - visibleNodes;
    const hiddenEdges = totalEdges - visibleEdges;
    
    if (hiddenNodes > 0) {
        console.log(`📊 Performance filtering: ${hiddenNodes} nodes and ${hiddenEdges} edges hidden for performance`);
    }
}

function showAllNodes() {
    if (!cy) return;
    
    const confirmation = confirm(`This will show all ${allNodes.length} nodes and may impact performance. Continue?`);
    if (confirmation) {
        document.getElementById('cy').innerHTML = '<div class="loading">Loading all nodes...</div>';
        
        setTimeout(() => {
            cy.elements().remove();
            cy.add(allNodes);
            cy.add(allEdges);
            cy.layout({ 
                name: 'cose', 
                animate: false,
                nodeRepulsion: 8000,
                idealEdgeLength: 80,
                numIter: 200
            }).run();
            updateAnalytics();
        }, 100);
    }
}

function toggleFields() {
    if (!cy) return;
    
    const fieldNodes = cy.nodes('[type="field"]');
    const areFieldsVisible = fieldNodes.length > 0 && fieldNodes[0].visible();
    
    if (areFieldsVisible) {
        // Hide fields and their edges
        fieldNodes.style('display', 'none');
        cy.edges('[type="has_field"]').style('display', 'none');
    } else {
        // Show fields from allNodes if not already added
        const currentNodeIds = new Set(cy.nodes().map(n => n.id()));
        const fieldNodesToAdd = allNodes.filter(n => 
            n.data.type === 'field' && !currentNodeIds.has(n.data.id)
        );
        
        if (fieldNodesToAdd.length > 0) {
            cy.add(fieldNodesToAdd);
            
            // Add corresponding edges
            const fieldNodeIds = new Set(fieldNodesToAdd.map(n => n.data.id));
            const fieldEdgesToAdd = allEdges.filter(e => 
                e.data.type === 'has_field' && fieldNodeIds.has(e.data.target)
            );
            cy.add(fieldEdgesToAdd);
        }
        
        // Show all field nodes and edges
        cy.nodes('[type="field"]').style('display', 'element');
        cy.edges('[type="has_field"]').style('display', 'element');
    }
    
    updateAnalytics();
}

function generateCytoscapeStyle(config) {
    const nodeColors = config.node_colors || {};
    const edgeColors = config.edge_colors || {};
    
    return [
        // Default node style
        {
            selector: 'node',
            style: {
                'background-color': '#cccccc',
                'label': 'data(label)',
                'width': 'mapData(complexity, 1, 20, 20, 80)',
                'height': 'mapData(complexity, 1, 20, 20, 80)',
                'font-size': 'mapData(complexity, 1, 10, 10, 16)',
                'font-family': 'Arial, sans-serif',
                'color': '#333333',
                'text-valign': 'center',
                'text-halign': 'center',
                'overlay-opacity': 0,
                'border-width': 2,
                'border-color': '#ffffff',
                'text-outline-width': 1,
                'text-outline-color': '#ffffff',
                'text-outline-opacity': 0.8
            }
        },
        
        // Package nodes
        {
            selector: 'node[type="package"]',
            style: {
                'background-color': nodeColors.package || '#4CAF50',
                'shape': 'round-rectangle',
                'width': 'mapData(size, 1, 10, 80, 140)',
                'height': 'mapData(size, 1, 10, 50, 90)',
                'font-size': '14px',
                'font-weight': 'bold'
            }
        },
        
        // Builtin package - special styling
        {
            selector: 'node[package="builtin"]',
            style: {
                'background-color': '#9E9E9E',
                'opacity': 0.8,
                'border-style': 'dashed',
                'border-color': '#616161',
                'font-size': '11px',
                'width': 35,
                'height': 35
            }
        },
        
        // Struct nodes
        {
            selector: 'node[type="struct"]',
            style: {
                'background-color': nodeColors.struct || '#2196F3',
                'shape': 'rectangle'
            }
        },
        
        // Interface nodes
        {
            selector: 'node[type="interface"]',
            style: {
                'background-color': nodeColors.interface || '#FF9800',
                'shape': 'diamond'
            }
        },
        
        // Function nodes
        {
            selector: 'node[type="function"]',
            style: {
                'background-color': nodeColors.function || '#9C27B0',
                'shape': 'ellipse'
            }
        },
        
        // Method nodes
        {
            selector: 'node[type="method"]',
            style: {
                'background-color': nodeColors.method || '#E91E63',
                'shape': 'ellipse'
            }
        },
        
        // Field nodes
        {
            selector: 'node[type="field"]',
            style: {
                'background-color': nodeColors.field || '#607D8B',
                'shape': 'triangle'
            }
        },
        
        // Parameter nodes
        {
            selector: 'node[type="parameter"]',
            style: {
                'background-color': nodeColors.parameter || '#00BCD4',
                'shape': 'round-triangle',
                'width': 'mapData(complexity, 1, 10, 15, 40)',
                'height': 'mapData(complexity, 1, 10, 15, 40)',
                'font-size': '10px'
            }
        },
        
        // Constant nodes
        {
            selector: 'node[type="constant"]',
            style: {
                'background-color': nodeColors.constant || '#795548',
                'shape': 'pentagon'
            }
        },
        
        // Variable nodes
        {
            selector: 'node[type="variable"]',
            style: {
                'background-color': nodeColors.variable || '#009688',
                'shape': 'hexagon'
            }
        },
        
        // Private/unexported nodes
        {
            selector: 'node[is_exported = "false"]',
            style: {
                'opacity': 0.7,
                'border-style': 'dashed'
            }
        },
        
        // High complexity nodes
        {
            selector: 'node[complexity >= 10]',
            style: {
                'border-width': 4,
                'border-color': '#FF6B6B'
            }
        },
        
        // Selected nodes
        {
            selector: 'node:selected',
            style: {
                'border-width': 6,
                'border-color': '#FFD700',
                'overlay-opacity': 0.2,
                'overlay-color': '#FFD700'
            }
        },
        
        // Highlighted nodes (search/filter)
        {
            selector: 'node.highlighted',
            style: {
                'border-width': 4,
                'border-color': '#FF4444',
                'overlay-opacity': 0.3,
                'overlay-color': '#FF4444'
            }
        },
        
        // Focused nodes
        {
            selector: 'node.focused',
            style: {
                'border-width': 5,
                'border-color': '#4CAF50',
                'overlay-opacity': 0.2,
                'overlay-color': '#4CAF50'
            }
        },
        
        // Hidden nodes
        {
            selector: 'node.hidden',
            style: {
                'display': 'none'
            }
        },
        
        // Default edge style
        {
            selector: 'edge',
            style: {
                'width': 'mapData(weight, 1, 10, 1, 4)',
                'line-color': '#cccccc',
                'target-arrow-color': '#cccccc',
                'target-arrow-shape': 'triangle',
                'curve-style': 'bezier',
                'overlay-opacity': 0,
                'opacity': 0.7
            }
        },
        
        // Call edges
        {
            selector: 'edge[type="calls"]',
            style: {
                'line-color': edgeColors.calls || '#666666',
                'target-arrow-color': edgeColors.calls || '#666666'
            }
        },
        
        // Import edges
        {
            selector: 'edge[type="imports"]',
            style: {
                'line-color': edgeColors.imports || '#2196F3',
                'target-arrow-color': edgeColors.imports || '#2196F3',
                'width': 3,
                'line-style': 'dashed'
            }
        },
        
        // Interface implementation edges
        {
            selector: 'edge[type="implements"]',
            style: {
                'line-color': edgeColors.implements || '#4CAF50',
                'target-arrow-color': edgeColors.implements || '#4CAF50',
                'line-style': 'dotted',
                'width': 3
            }
        },
        
        // Embedding edges
        {
            selector: 'edge[type="embeds"]',
            style: {
                'line-color': edgeColors.embeds || '#FF5722',
                'target-arrow-color': edgeColors.embeds || '#FF5722',
                'width': 4
            }
        },
        
        // Field relationship edges
        {
            selector: 'edge[type="has_field"]',
            style: {
                'line-color': edgeColors.has_field || '#607D8B',
                'target-arrow-color': edgeColors.has_field || '#607D8B',
                'width': 2
            }
        },
        
        // Parameter relationship edges
        {
            selector: 'edge[type="has_parameter"]',
            style: {
                'line-color': edgeColors.has_parameter || '#00BCD4',
                'target-arrow-color': edgeColors.has_parameter || '#00BCD4',
                'width': 2,
                'line-style': 'dashed'
            }
        },
        
        // Method relationship edges
        {
            selector: 'edge[type="has_method"]',
            style: {
                'line-color': edgeColors.has_method || '#E91E63',
                'target-arrow-color': edgeColors.has_method || '#E91E63',
                'width': 3
            }
        },
        
        // Parameter type edges
        {
            selector: 'edge[type="parameter_type"]',
            style: {
                'line-color': edgeColors.parameter_type || '#9C27B0',
                'target-arrow-color': edgeColors.parameter_type || '#9C27B0',
                'width': 2
            }
        },
        
        // Return type edges
        {
            selector: 'edge[type="returns"]',
            style: {
                'line-color': edgeColors.returns || '#FF9800',
                'target-arrow-color': edgeColors.returns || '#FF9800',
                'width': 2,
                'line-style': 'solid'
            }
        },
        
        // Method ownership edges
        {
            selector: 'edge[type="method_of"]',
            style: {
                'line-color': edgeColors.method_of || '#795548',
                'target-arrow-color': edgeColors.method_of || '#795548',
                'width': 2
            }
        },
        
        // Usage edges
        {
            selector: 'edge[type="uses"]',
            style: {
                'line-color': edgeColors.uses || '#666666',
                'target-arrow-color': edgeColors.uses || '#666666',
                'width': 1,
                'opacity': 0.6
            }
        },
        
        // Construction/instantiation edges
        {
            selector: 'edge[type="constructs"]',
            style: {
                'line-color': edgeColors.constructs || '#8BC34A',
                'target-arrow-color': edgeColors.constructs || '#8BC34A',
                'width': 3,
                'line-style': 'dashed'
            }
        },
        
        // Conditional edges
        {
            selector: 'edge[?conditional]',
            style: {
                'line-style': 'dashed',
                'opacity': 0.5
            }
        },
        
        // Selected edges
        {
            selector: 'edge:selected',
            style: {
                'overlay-opacity': 0.3,
                'overlay-color': '#FFD700',
                'width': 6
            }
        }
    ];
}

function setupGraphEvents() {
    // Node click handler
    cy.on('tap', 'node', function(event) {
        const node = event.target;
        selectNode(node);
        showNodeInfo(node);
    });
    
    // Background click to clear selection
    cy.on('tap', function(event) {
        if (event.target === cy) {
            clearSelection();
        }
    });
    
    // Right-click context menu
    cy.on('cxttap', 'node', function(event) {
        event.preventDefault();
        const node = event.target;
        const position = event.renderedPosition || event.position;
        showContextMenu(node, position);
    });
    
    // Hide context menu on any click
    cy.on('tap', function() {
        hideContextMenu();
    });
    
    // Mouse hover for tooltips
    cy.on('mouseover', 'node', function(event) {
        const node = event.target;
        showTooltip(node, event.renderedPosition);
    });
    
    cy.on('mouseout', 'node', function() {
        hideTooltip();
    });
    
    // Edge hover
    cy.on('mouseover', 'edge', function(event) {
        const edge = event.target;
        const tooltip = document.getElementById('tooltip');
        tooltip.innerHTML = `${edge.data('type')}: ${edge.data('context') || 'N/A'}`;
        tooltip.style.left = event.originalEvent.clientX + 10 + 'px';
        tooltip.style.top = event.originalEvent.clientY + 10 + 'px';
        tooltip.classList.add('show');
    });
    
    cy.on('mouseout', 'edge', function() {
        hideTooltip();
    });
}

function setupEventListeners() {
    // Node type filter
    document.getElementById('node-filter').addEventListener('change', function(e) {
        applyFilters();
    });
    
    // Visibility filter
    document.getElementById('visibility-filter').addEventListener('change', function(e) {
        applyFilters();
    });
    
    // Complexity filter
    document.getElementById('complexity-filter').addEventListener('input', function(e) {
        document.getElementById('complexity-value').textContent = e.target.value + '+';
        applyFilters();
    });
    
    // Layout selector - with debugging
    const layoutSelector = document.getElementById('layout-selector');
    if (layoutSelector) {
        layoutSelector.addEventListener('change', function(e) {
            console.log(`🎯 Layout selector changed to: ${e.target.value}`);
            changeLayout(e.target.value);
        });
    } else {
        console.warn('⚠️  Layout selector element not found');
    }
    
    // Enhanced search with autocomplete
    const searchInput = document.getElementById('search-input');
    const searchResults = document.getElementById('search-results');
    
    let searchTimeout;
    searchInput.addEventListener('input', function(e) {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            handleSearch(e.target.value);
        }, 300);
    });
    
    searchInput.addEventListener('focus', function() {
        if (currentSearchResults.length > 0) {
            searchResults.style.display = 'block';
        }
    });
    
    searchInput.addEventListener('blur', function() {
        setTimeout(() => searchResults.style.display = 'none', 200);
    });
    
    // Panel toggle
    document.getElementById('panel-toggle').addEventListener('click', function() {
        toggleSidePanel();
    });
    
    // Context menu actions
    document.getElementById('focus-node').addEventListener('click', function() {
        if (selectedNode) focusOnNode(selectedNode);
        hideContextMenu();
    });
    
    document.getElementById('show-neighbors').addEventListener('click', function() {
        if (selectedNode) showNeighbors(selectedNode);
        hideContextMenu();
    });
    
    document.getElementById('hide-node').addEventListener('click', function() {
        if (selectedNode) hideNode(selectedNode);
        hideContextMenu();
    });
    
    document.getElementById('copy-name').addEventListener('click', function() {
        if (selectedNode) copyNodeName(selectedNode);
        hideContextMenu();
    });
}

function setupPanelTabs() {
    const tabs = document.querySelectorAll('.panel-tab');
    tabs.forEach(tab => {
        tab.addEventListener('click', function() {
            const tabId = this.dataset.tab;
            switchTab(tabId);
        });
    });
}

function switchTab(tabId) {
    // Update tab buttons
    document.querySelectorAll('.panel-tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.tab === tabId);
    });
    
    // Show/hide tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        content.style.display = content.id === `${tabId}-tab` ? 'block' : 'none';
    });
}

function applyFilters() {
    if (!cy) return;
    
    const nodeFilter = document.getElementById('node-filter').value;
    const visibilityFilter = document.getElementById('visibility-filter').value;
    const complexityThreshold = parseInt(document.getElementById('complexity-filter').value);
    
    let filteredNodes = cy.nodes();
    
    // Apply node type filter
    if (nodeFilter === 'packages-types') {
        // Show only packages, structs, and interfaces - the architectural view
        filteredNodes = filteredNodes.filter(node => {
            const type = node.data('type');
            return type === 'package' || type === 'struct' || type === 'interface';
        });
    } else if (nodeFilter !== 'all') {
        // Single type filtering
        filteredNodes = filteredNodes.filter(`[type="${nodeFilter}"]`);
    }
    
    // Apply visibility filter
    if (visibilityFilter !== 'all') {
        if (visibilityFilter === 'public') {
            filteredNodes = filteredNodes.filter('[is_exported = "true"]');
        } else {
            filteredNodes = filteredNodes.filter('[is_exported = "false"]');
        }
    }
    
    // Apply complexity filter
    if (complexityThreshold > 0) {
        filteredNodes = filteredNodes.filter(node => {
            return node.data('complexity') >= complexityThreshold;
        });
    }
    
    // Show/hide nodes
    cy.nodes().style('display', 'none');
    cy.edges().style('display', 'none');
    
    filteredNodes.style('display', 'element');
    
    // Show edges between visible nodes
    const visibleNodeIds = filteredNodes.map(n => n.id());
    cy.edges().forEach(edge => {
        const sourceVisible = visibleNodeIds.includes(edge.source().id());
        const targetVisible = visibleNodeIds.includes(edge.target().id());
        if (sourceVisible && targetVisible) {
            edge.style('display', 'element');
        }
    });
    
    // Fit to visible elements
    if (filteredNodes.length > 0) {
        cy.fit(filteredNodes, 50);
    }
    
    updateAnalytics();
}

function handleSearch(query) {
    const searchResults = document.getElementById('search-results');
    
    if (!query.trim()) {
        searchResults.style.display = 'none';
        cy.nodes().removeClass('highlighted');
        currentSearchResults = [];
        return;
    }
    
    const searchQuery = query.toLowerCase();
    const matchingNodes = [];
    
    cy.nodes().forEach(node => {
        const data = node.data();
        const label = data.label.toLowerCase();
        const type = data.type.toLowerCase();
        const pkg = data.package.toLowerCase();
        const signature = (data.signature || '').toLowerCase();
        
        if (label.includes(searchQuery) || 
            type.includes(searchQuery) || 
            pkg.includes(searchQuery) ||
            signature.includes(searchQuery)) {
            matchingNodes.push({
                node: node,
                score: calculateSearchScore(data, searchQuery)
            });
        }
    });
    
    // Sort by relevance score
    matchingNodes.sort((a, b) => b.score - a.score);
    currentSearchResults = matchingNodes.slice(0, 10); // Top 10 results
    
    // Update search dropdown
    updateSearchResults(currentSearchResults);
    
    // Highlight nodes
    cy.nodes().removeClass('highlighted');
    if (currentSearchResults.length > 0) {
        currentSearchResults.forEach(result => {
            result.node.addClass('highlighted');
        });
        
        // Focus on first result
        const firstResult = currentSearchResults[0].node;
        cy.center(firstResult);
    }
}

function calculateSearchScore(data, query) {
    let score = 0;
    const label = data.label.toLowerCase();
    const type = data.type.toLowerCase();
    
    // Exact match bonus
    if (label === query) score += 100;
    if (label.startsWith(query)) score += 50;
    if (type === query) score += 80;
    
    // Complexity and size bonuses
    score += (data.complexity || 1) * 2;
    score += (data.size || 1);
    
    return score;
}

function updateSearchResults(results) {
    const searchResults = document.getElementById('search-results');
    searchResults.innerHTML = '';
    
    if (results.length === 0) {
        searchResults.style.display = 'none';
        return;
    }
    
    results.forEach(result => {
        const data = result.node.data();
        const resultElement = document.createElement('div');
        resultElement.className = 'search-result';
        resultElement.innerHTML = `
            <span>${data.label}</span>
            <span class="search-result-type">${data.type}</span>
        `;
        
        resultElement.addEventListener('click', () => {
            selectNode(result.node);
            showNodeInfo(result.node);
            cy.center(result.node);
            searchResults.style.display = 'none';
        });
        
        searchResults.appendChild(resultElement);
    });
    
    searchResults.style.display = 'block';
}

function selectNode(node) {
    cy.nodes().removeClass('focused');
    node.addClass('focused');
    selectedNode = node;
}

function clearSelection() {
    cy.nodes().removeClass('focused highlighted');
    selectedNode = null;
    document.getElementById('node-info').classList.add('empty');
    document.getElementById('node-info').innerHTML = '<p>Select a node to view detailed information</p>';
}

function showNodeInfo(node) {
    const nodeInfo = document.getElementById('node-info');
    const data = node.data();
    
    nodeInfo.classList.remove('empty');
    
    // Create enhanced node info HTML
    const typeInfo = data.type_info || {};
    const positionInfo = data.position_info || {};
    
    let html = `
        <div class="node-header">
            <div class="node-type-badge ${data.type}">${data.type}</div>
            <h3 class="node-title">${data.label}</h3>
        </div>
    `;
    
    // Add signature if available
    if (data.signature) {
        html += `<div class="node-signature">${data.signature}</div>`;
    }
    
    // Basic information
    html += `
        <div class="info-section">
            <div class="info-section-title">Basic Information</div>
            <div class="info-item">
                <span class="info-label">Package</span>
                <span class="info-value">${data.package}</span>
            </div>
            <div class="info-item">
                <span class="info-label">Visibility</span>
                <span class="info-value">${data.is_exported ? 'Public' : 'Private'}</span>
            </div>
            <div class="info-item">
                <span class="info-label">Complexity</span>
                <span class="info-value">
                    ${data.complexity}
                    <div class="complexity-bar">
                        <div class="complexity-fill" style="width: ${Math.min(data.complexity * 5, 100)}%"></div>
                    </div>
                </span>
            </div>
        </div>
    `;
    
    // Type information
    if (Object.keys(typeInfo).length > 0) {
        html += `
            <div class="info-section">
                <div class="info-section-title">Type Information</div>
                <div class="info-grid">
        `;
        
        if (typeInfo.kind) {
            html += `
                <div class="info-item">
                    <span class="info-label">Kind</span>
                    <span class="info-value">${typeInfo.kind}</span>
                </div>
            `;
        }
        
        if (typeInfo.is_pointer !== undefined) {
            html += `
                <div class="info-item">
                    <span class="info-label">Pointer</span>
                    <span class="info-value">${typeInfo.is_pointer ? 'Yes' : 'No'}</span>
                </div>
            `;
        }
        
        if (typeInfo.is_slice !== undefined) {
            html += `
                <div class="info-item">
                    <span class="info-label">Slice</span>
                    <span class="info-value">${typeInfo.is_slice ? 'Yes' : 'No'}</span>
                </div>
            `;
        }
        
        if (typeInfo.is_channel !== undefined) {
            html += `
                <div class="info-item">
                    <span class="info-label">Channel</span>
                    <span class="info-value">${typeInfo.is_channel ? 'Yes' : 'No'}</span>
                </div>
            `;
        }
        
        html += '</div></div>';
    }
    
    // Source location
    if (positionInfo.filename) {
        html += `
            <div class="info-section">
                <div class="info-section-title">Source Location</div>
                <div class="info-item">
                    <span class="info-label">File</span>
                    <span class="info-value">${positionInfo.filename}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Line</span>
                    <span class="info-value">${positionInfo.line || 'N/A'}</span>
                </div>
            </div>
        `;
    }
    
    // Connection information
    const connectedEdges = node.connectedEdges();
    const incomingEdges = connectedEdges.filter(edge => edge.target().id() === data.id);
    const outgoingEdges = connectedEdges.filter(edge => edge.source().id() === data.id);
    
    html += `
        <div class="info-section">
            <div class="info-section-title">Connections</div>
            <div class="info-grid">
                <div class="info-item">
                    <span class="info-label">Incoming</span>
                    <span class="info-value">${incomingEdges.length}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Outgoing</span>
                    <span class="info-value">${outgoingEdges.length}</span>
                </div>
            </div>
        </div>
    `;
    
    // Documentation if available
    if (data.documentation) {
        html += `
            <div class="info-section">
                <div class="info-section-title">Documentation</div>
                <div style="font-style: italic; color: #6c757d; padding: 8px 0;">
                    ${data.documentation}
                </div>
            </div>
        `;
    }
    
    nodeInfo.innerHTML = html;
}

function initializeAnalytics() {
    updateAnalytics();
}

function updateAnalytics() {
    if (!cy) return;
    
    const visibleNodes = cy.nodes().filter('[^display]');
    const visibleEdges = cy.edges().filter('[^display]');
    
    // Update overview stats
    document.getElementById('total-nodes').textContent = visibleNodes.length;
    document.getElementById('total-edges').textContent = visibleEdges.length;
    
    // Count packages
    const packages = visibleNodes.filter('[type="package"]');
    document.getElementById('package-count').textContent = packages.length;
    
    // Calculate average complexity
    let totalComplexity = 0;
    let complexityCount = 0;
    visibleNodes.forEach(node => {
        const complexity = node.data('complexity') || 0;
        if (complexity > 0) {
            totalComplexity += complexity;
            complexityCount++;
        }
    });
    const avgComplexity = complexityCount > 0 ? (totalComplexity / complexityCount).toFixed(1) : '0';
    document.getElementById('avg-complexity').textContent = avgComplexity;
    
    // Update node types chart
    updateNodeTypesChart(visibleNodes);
    
    // Update complexity chart
    updateComplexityChart(visibleNodes);
    
    // Update most connected nodes
    updateMostConnectedNodes(visibleNodes);
}

function updateNodeTypesChart(nodes) {
    const typeCount = {};
    nodes.forEach(node => {
        const type = node.data('type');
        typeCount[type] = (typeCount[type] || 0) + 1;
    });
    
    const chart = document.getElementById('node-types-chart');
    chart.innerHTML = '';
    
    const maxCount = Math.max(...Object.values(typeCount));
    
    Object.entries(typeCount).forEach(([type, count]) => {
        const bar = document.createElement('div');
        bar.className = 'chart-bar';
        bar.style.height = `${(count / maxCount) * 100}%`;
        bar.title = `${type}: ${count}`;
        chart.appendChild(bar);
    });
}

function updateComplexityChart(nodes) {
    const complexityRanges = {
        '1-2': 0,
        '3-5': 0,
        '6-10': 0,
        '11-20': 0,
        '20+': 0
    };
    
    nodes.forEach(node => {
        const complexity = node.data('complexity') || 1;
        if (complexity <= 2) complexityRanges['1-2']++;
        else if (complexity <= 5) complexityRanges['3-5']++;
        else if (complexity <= 10) complexityRanges['6-10']++;
        else if (complexity <= 20) complexityRanges['11-20']++;
        else complexityRanges['20+']++;
    });
    
    const chart = document.getElementById('complexity-chart');
    chart.innerHTML = '';
    
    const maxCount = Math.max(...Object.values(complexityRanges));
    
    Object.entries(complexityRanges).forEach(([range, count]) => {
        const bar = document.createElement('div');
        bar.className = 'chart-bar';
        bar.style.height = `${(count / maxCount) * 100}%`;
        bar.title = `Complexity ${range}: ${count} nodes`;
        chart.appendChild(bar);
    });
}

function updateMostConnectedNodes(nodes) {
    const connections = [];
    nodes.forEach(node => {
        const connectedEdges = node.connectedEdges();
        connections.push({
            node: node,
            count: connectedEdges.length
        });
    });
    
    connections.sort((a, b) => b.count - a.count);
    const top5 = connections.slice(0, 5);
    
    const container = document.getElementById('connected-nodes');
    container.innerHTML = '';
    
    top5.forEach(item => {
        const data = item.node.data();
        const div = document.createElement('div');
        div.className = 'info-item';
        div.style.cursor = 'pointer';
        div.innerHTML = `
            <span class="info-label">${data.label}</span>
            <span class="info-value">${item.count}</span>
        `;
        div.addEventListener('click', () => {
            selectNode(item.node);
            showNodeInfo(item.node);
            cy.center(item.node);
        });
        container.appendChild(div);
    });
}

function changeLayout(layoutName) {
    if (!cy) return;
    
    console.log(`🔄 Changing layout to: ${layoutName}`);
    
    let layoutOptions = {
        name: layoutName,
        animate: layoutName !== 'grid',
        animationDuration: 800,
        fit: true,
        padding: 30,
        randomize: false
    };
    
    // Customize layout options based on type
    switch (layoutName) {
        case 'cose':
            layoutOptions = {
                ...layoutOptions,
                nodeRepulsion: 12000,
                idealEdgeLength: 120,
                edgeElasticity: 16,
                nestingFactor: 2,
                gravity: 50,
                numIter: performanceMode ? 200 : 500,
                initialTemp: 100,
                coolingFactor: 0.9,
                minTemp: 10,
                randomize: true
            };
            break;
            
        case 'breadthfirst':
            layoutOptions.directed = true;
            layoutOptions.roots = cy.nodes('[type="package"]');
            layoutOptions.spacingFactor = 1.5;
            break;
            
        case 'concentric':
            layoutOptions.concentric = function(node) {
                const type = node.data('type');
                if (type === 'package') return 5;
                if (type === 'struct' || type === 'interface') return 4;
                if (type === 'function' || type === 'method') return 3;
                return 1;
            };
            layoutOptions.levelWidth = function() { return 2; };
            layoutOptions.minNodeSpacing = 50;
            break;
            
        case 'circle':
            layoutOptions.radius = Math.min(400, cy.nodes().length * 3);
            break;
            
        case 'grid':
            layoutOptions.rows = Math.ceil(Math.sqrt(cy.nodes().length));
            layoutOptions.cols = Math.ceil(Math.sqrt(cy.nodes().length));
            break;
    }
    
    console.log(`📐 Layout options:`, layoutOptions);
    
    const layout = cy.layout(layoutOptions);
    layout.run();
    
    // Update the layout selector to match
    const selector = document.getElementById('layout-selector');
    if (selector && selector.value !== layoutName) {
        selector.value = layoutName;
    }
}

function toggleSidePanel() {
    const panel = document.getElementById('side-panel');
    const toggle = document.getElementById('panel-toggle');
    
    panel.classList.toggle('collapsed');
    toggle.classList.toggle('collapsed');
    
    // Update toggle button
    toggle.textContent = panel.classList.contains('collapsed') ? '📊' : '❌';
}

function showTooltip(node, position) {
    const tooltip = document.getElementById('tooltip');
    const data = node.data();
    
    tooltip.innerHTML = `
        <strong>${data.label}</strong><br>
        Type: ${data.type}<br>
        Complexity: ${data.complexity}<br>
        ${data.is_exported ? 'Public' : 'Private'}
    `;
    
    tooltip.style.left = position.x + 10 + 'px';
    tooltip.style.top = position.y - 10 + 'px';
    tooltip.classList.add('show');
}

function hideTooltip() {
    const tooltip = document.getElementById('tooltip');
    tooltip.classList.remove('show');
}

function showContextMenu(node, position) {
    const menu = document.getElementById('context-menu');
    selectedNode = node;
    
    menu.style.left = position.x + 'px';
    menu.style.top = position.y + 'px';
    menu.style.display = 'block';
}

function hideContextMenu() {
    document.getElementById('context-menu').style.display = 'none';
}

function focusOnNode(node) {
    cy.center(node);
    cy.zoom(2);
    selectNode(node);
    showNodeInfo(node);
}

function showNeighbors(node) {
    cy.nodes().style('opacity', 0.3);
    cy.edges().style('opacity', 0.3);
    
    const neighbors = node.neighborhood();
    neighbors.style('opacity', 1);
    node.style('opacity', 1);
}

function hideNode(node) {
    node.addClass('hidden');
    updateAnalytics();
}

function copyNodeName(node) {
    const name = node.data('full_name') || node.data('label');
    navigator.clipboard.writeText(name).then(() => {
        console.log('Node name copied to clipboard:', name);
    });
}

// Export enhanced API
window.GraphVisualization = {
    getCytoscape: () => cy,
    getGraphData: () => graphData,
    selectNode,
    clearSelection,
    showNodeInfo,
    applyFilters,
    handleSearch,
    changeLayout,
    toggleSidePanel,
    updateAnalytics,
    focusOnNode,
    showNeighbors
};
