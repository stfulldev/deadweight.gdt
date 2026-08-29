package analysis

import (
	"errors"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/project"
)

type graphVisitState uint8

const (
	graphUnvisited graphVisitState = iota
	graphVisiting
	graphVisited
)

type graphEdgeKey struct {
	fromCanonical    string
	fromDisplay      string
	toCanonical      string
	toDisplay        string
	rawTarget        string
	resourceID       string
	kind             EdgeKind
	resolved         bool
	classification   TargetClassification
	resolutionReason project.ResolutionReason
}

type graphBuilder struct {
	root            project.ResolvedPath
	nodes           map[string]GraphNode
	edges           map[graphEdgeKey]GraphEdge
	resources       map[ResourceIdentity]struct{}
	dependencyCount int64
}

type graphCandidate struct {
	edge   GraphEdge
	target project.ResolvedPath
}

func newGraphBuilder(root project.ResolvedPath) *graphBuilder {
	return &graphBuilder{
		root:      root,
		nodes:     make(map[string]GraphNode),
		edges:     make(map[graphEdgeKey]GraphEdge),
		resources: make(map[ResourceIdentity]struct{}),
	}
}

func (builder *graphBuilder) addNode(path project.ResolvedPath) error {
	if _, exists := builder.nodes[path.Canonical]; exists {
		return nil
	}
	if path.Canonical != builder.root.Canonical {
		count, err := checkedAdd(builder.dependencyCount, 1)
		if err != nil {
			return err
		}
		builder.dependencyCount = count
	}
	builder.nodes[path.Canonical] = GraphNode{Canonical: path.Canonical, Display: path.Display}

	return nil
}

func (builder *graphBuilder) addEdge(edge GraphEdge) error {
	key := graphEdgeKey{
		fromCanonical:    edge.FromCanonical,
		fromDisplay:      edge.FromDisplay,
		toCanonical:      edge.ToCanonical,
		toDisplay:        edge.ToDisplay,
		rawTarget:        edge.RawTarget,
		resourceID:       edge.ResourceID,
		kind:             edge.Kind,
		resolved:         edge.Resolved,
		classification:   edge.Classification,
		resolutionReason: edge.ResolutionReason,
	}
	current := builder.edges[key]
	occurrences, err := checkedAdd(current.Occurrences, edge.Occurrences)
	if err != nil {
		return err
	}
	edge.Occurrences = occurrences
	builder.edges[key] = edge

	return nil
}

func (builder *graphBuilder) addResources(resources []ResourceIdentity) {
	for _, resource := range resources {
		builder.resources[resource] = struct{}{}
	}
}

func (builder *graphBuilder) finish() (DependencyGraph, []ResourceIdentity) {
	graph := DependencyGraph{
		RootCanonical:     builder.root.Canonical,
		RootDisplay:       builder.root.Display,
		SceneDependencies: builder.dependencyCount,
		Nodes:             make([]GraphNode, 0, len(builder.nodes)),
		Edges:             make([]GraphEdge, 0, len(builder.edges)),
	}
	for _, node := range builder.nodes {
		graph.Nodes = append(graph.Nodes, node)
	}
	sort.Slice(graph.Nodes, func(left, right int) bool {
		if graph.Nodes[left].Canonical != graph.Nodes[right].Canonical {
			return graph.Nodes[left].Canonical < graph.Nodes[right].Canonical
		}

		return graph.Nodes[left].Display < graph.Nodes[right].Display
	})
	for _, edge := range builder.edges {
		graph.Edges = append(graph.Edges, edge)
	}
	sortGraphEdges(graph.Edges)

	return cloneDependencyGraph(graph), sortedResourceIdentities(builder.resources)
}

func sortGraphEdges(edges []GraphEdge) {
	sort.Slice(edges, func(left, right int) bool {
		first := edges[left]
		second := edges[right]
		if first.FromCanonical != second.FromCanonical {
			return first.FromCanonical < second.FromCanonical
		}
		if first.ToCanonical != second.ToCanonical {
			return first.ToCanonical < second.ToCanonical
		}
		if first.Kind != second.Kind {
			return first.Kind < second.Kind
		}
		if first.RawTarget != second.RawTarget {
			return first.RawTarget < second.RawTarget
		}
		if first.ResourceID != second.ResourceID {
			return first.ResourceID < second.ResourceID
		}
		if first.Classification != second.Classification {
			return first.Classification < second.Classification
		}
		if first.ResolutionReason != second.ResolutionReason {
			return first.ResolutionReason < second.ResolutionReason
		}
		if first.FromDisplay != second.FromDisplay {
			return first.FromDisplay < second.FromDisplay
		}
		if first.ToDisplay != second.ToDisplay {
			return first.ToDisplay < second.ToDisplay
		}

		return first.Occurrences < second.Occurrences
	})
}

func (state *invocationState) discoverGraph(root project.ResolvedPath) (DependencyGraph, []ResourceIdentity, error) {
	builder := newGraphBuilder(root)
	state.graphStates = make(map[string]graphVisitState)
	state.graphStack = nil
	state.graphStackIndices = make(map[string]int)
	if err := state.visitGraphScene(root, builder); err != nil {
		return DependencyGraph{}, nil, err
	}

	graph, resources := builder.finish()
	return graph, resources, nil
}

func (state *invocationState) visitGraphScene(path project.ResolvedPath, builder *graphBuilder) error {
	local, err := state.loadLocalSummary(path)
	if err != nil {
		return err
	}
	if err := builder.addNode(path); err != nil {
		return err
	}

	state.graphStates[path.Canonical] = graphVisiting
	state.graphStackIndices[path.Canonical] = len(state.graphStack)
	state.graphStack = append(state.graphStack, path)

	resources := state.resolveSceneResources(path, local.ExternalResources)
	builder.addResources(resourceIdentities(path, resources))
	candidates := graphCandidates(path, local, resources)
	for _, candidate := range candidates {
		if !candidate.edge.Resolved {
			if err := builder.addEdge(candidate.edge); err != nil {
				return err
			}
			continue
		}

		visitState := state.graphStates[candidate.target.Canonical]
		if visitState == graphUnvisited {
			_, loadErr := state.loadLocalSummary(candidate.target)
			if loadErr != nil {
				var unavailable *sceneLoadError
				if errors.As(loadErr, &unavailable) {
					edge := candidate.edge
					edge.Resolved = false
					edge.ToCanonical = ""
					edge.ToDisplay = ""
					edge.Classification = TargetUnavailableScene
					if err := builder.addEdge(edge); err != nil {
						return err
					}
					continue
				}
				return loadErr
			}
		}

		if err := builder.addEdge(candidate.edge); err != nil {
			return err
		}
		switch visitState {
		case graphVisiting:
			return state.cycleError(candidate.target)
		case graphUnvisited:
			if err := state.visitGraphScene(candidate.target, builder); err != nil {
				return err
			}
		case graphVisited:
			// The canonical child was already fully traversed in this invocation.
		}
	}

	state.graphStack = state.graphStack[:len(state.graphStack)-1]
	delete(state.graphStackIndices, path.Canonical)
	state.graphStates[path.Canonical] = graphVisited

	return nil
}

func (state *invocationState) cycleError(target project.ResolvedPath) *CycleError {
	start := state.graphStackIndices[target.Canonical]
	canonical := make([]string, 0, len(state.graphStack)-start+1)
	display := make([]string, 0, len(state.graphStack)-start+1)
	for _, path := range state.graphStack[start:] {
		canonical = append(canonical, path.Canonical)
		identity := path.Display
		if identity == "" {
			identity = path.Canonical
		}
		display = append(display, identity)
	}
	canonical = append(canonical, target.Canonical)
	identity := target.Display
	if identity == "" {
		identity = target.Canonical
	}
	display = append(display, identity)

	return &CycleError{Canonical: canonical, Display: display}
}

func graphCandidates(
	declaring project.ResolvedPath,
	local LocalSummary,
	resources map[string]resourceResolution,
) []graphCandidate {
	candidates := make([]graphCandidate, 0, len(local.Mounts)+1)
	for _, mount := range local.Mounts {
		target, unresolved := classifyTarget(declaring, mount.Reference, mount.Placeholder, resources)
		candidates = append(candidates, graphCandidateForTarget(
			declaring,
			EdgeInstance,
			mount.Reference.ID,
			target,
			unresolved,
		))
	}
	if local.InheritedRoot != nil {
		target, unresolved := classifyTarget(declaring, local.InheritedRoot.Reference, "", resources)
		candidates = append(candidates, graphCandidateForTarget(
			declaring,
			EdgeInheritance,
			local.InheritedRoot.Reference.ID,
			target,
			unresolved,
		))
	}
	sort.Slice(candidates, func(left, right int) bool {
		first := candidates[left]
		second := candidates[right]
		if first.target.Canonical != second.target.Canonical {
			return first.target.Canonical < second.target.Canonical
		}
		if first.edge.Kind != second.edge.Kind {
			return first.edge.Kind < second.edge.Kind
		}
		if first.edge.RawTarget != second.edge.RawTarget {
			return first.edge.RawTarget < second.edge.RawTarget
		}
		if first.edge.ResourceID != second.edge.ResourceID {
			return first.edge.ResourceID < second.edge.ResourceID
		}
		if first.edge.Classification != second.edge.Classification {
			return first.edge.Classification < second.edge.Classification
		}

		return first.edge.ResolutionReason < second.edge.ResolutionReason
	})

	return candidates
}

func graphCandidateForTarget(
	declaring project.ResolvedPath,
	kind EdgeKind,
	resourceID string,
	target project.ResolvedPath,
	unresolved *targetEvidence,
) graphCandidate {
	edge := GraphEdge{
		FromCanonical: declaring.Canonical,
		FromDisplay:   declaring.Display,
		ResourceID:    resourceID,
		Kind:          kind,
		Occurrences:   1,
	}
	if unresolved != nil {
		edge.RawTarget = unresolved.RawTarget
		edge.Classification = unresolved.Classification
		edge.ResolutionReason = unresolved.ResolutionReason
		return graphCandidate{edge: edge}
	}

	edge.ToCanonical = target.Canonical
	edge.ToDisplay = target.Display
	edge.RawTarget = target.Original
	edge.Resolved = true
	edge.ResolutionReason = project.ResolutionResolved

	return graphCandidate{edge: edge, target: target}
}

func graphDependencyPaths(graph DependencyGraph) []string {
	dependencies := make([]string, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if node.Canonical != graph.RootCanonical {
			dependencies = append(dependencies, node.Canonical)
		}
	}

	return dependencies
}
