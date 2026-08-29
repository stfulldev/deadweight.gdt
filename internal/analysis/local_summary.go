package analysis

import (
	"errors"
	"sort"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

var errMissingSceneRoot = errors.New("local scene summary requires a parsed document with one root node")

type localNodeClass uint8

const (
	classOrdinary localNodeClass = iota
	classMount
	classInheritedRoot
	classOverrideStub
)

type localNodeState struct {
	node      tscn.Node
	class     localNodeClass
	path      string
	depth     OptionalDepth
	pathValid bool
	resolved  bool
	resolving bool
}

// BuildLocalSummary converts one parsed TSCN document into a deterministic,
// non-recursive local scene summary.
func BuildLocalSummary(document *tscn.Document) (LocalSummary, error) {
	if document == nil || len(document.Nodes) == 0 {
		return LocalSummary{}, errMissingSceneRoot
	}

	resources, resourcesByID := extractExternalResources(document.ExtResources)
	states, pathIndex, findings := buildLocalStates(document.Nodes)

	for index := range states {
		resolveLocalDepth(index, states, pathIndex, &findings)
	}

	summary := LocalSummary{ExternalResources: resources}
	for index := range states {
		addLocalState(&summary, states[index], index, resourcesByID)
	}

	sortParentFindings(findings)
	summary.Findings = findings
	summary.DepthPartial = len(findings) > 0

	return summary, nil
}

func extractExternalResources(input map[string]tscn.ExtResource) ([]ExternalResource, map[string]ExternalResource) {
	ids := make([]string, 0, len(input))
	for id := range input {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	resources := make([]ExternalResource, 0, len(ids))
	resourcesByID := make(map[string]ExternalResource, len(ids))
	for _, id := range ids {
		parsed := input[id]
		resource := ExternalResource{
			ID:       parsed.ID,
			Type:     parsed.Type,
			UID:      parsed.UID,
			Path:     parsed.Path,
			Position: parsed.Position,
		}
		resources = append(resources, resource)
		resourcesByID[id] = resource
	}

	return resources, resourcesByID
}

func buildLocalStates(nodes []tscn.Node) ([]localNodeState, map[string][]int, []ParentFinding) {
	states := make([]localNodeState, len(nodes))
	pathIndex := make(map[string][]int, len(nodes))
	findings := make([]ParentFinding, 0)

	for index, node := range nodes {
		state := localNodeState{
			node:  cloneParsedNode(node),
			class: classifyNode(index, node),
		}

		if index == 0 {
			state.path = "."
			state.pathValid = true
			state.depth = OptionalDepth{Value: 1, Known: true}
			state.resolved = true
		} else if validSerializedParent(node.Parent) {
			state.path = localNodePath(node.Parent, node.Name)
			state.pathValid = true
		} else {
			findings = append(findings, parentFinding(ParentInvalid, node, ""))
		}

		states[index] = state
		if state.pathValid {
			pathIndex[state.path] = append(pathIndex[state.path], index)
		}
	}

	return states, pathIndex, findings
}

func classifyNode(index int, node tscn.Node) localNodeClass {
	if index == 0 && node.Instance != nil {
		return classInheritedRoot
	}
	if node.InstancePlaceholder != "" || (index > 0 && node.Instance != nil) {
		return classMount
	}
	if node.Type != "" {
		return classOrdinary
	}

	return classOverrideStub
}

func cloneParsedNode(node tscn.Node) tscn.Node {
	cloned := node
	if node.Instance != nil {
		reference := *node.Instance
		cloned.Instance = &reference
	}
	if node.Index != nil {
		index := *node.Index
		cloned.Index = &index
	}
	if node.ShadowEnabled != nil {
		shadowEnabled := *node.ShadowEnabled
		cloned.ShadowEnabled = &shadowEnabled
	}

	return cloned
}

func validSerializedParent(parent string) bool {
	if parent == "." {
		return true
	}
	if parent == "" || strings.HasPrefix(parent, "/") || strings.Contains(parent, "\\") {
		return false
	}
	for _, segment := range strings.Split(parent, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}

	return true
}

func localNodePath(parent, name string) string {
	if parent == "." {
		return name
	}

	return parent + "/" + name
}

func resolveLocalDepth(index int, states []localNodeState, pathIndex map[string][]int, findings *[]ParentFinding) OptionalDepth {
	state := &states[index]
	if state.resolved {
		return state.depth
	}
	if !state.pathValid {
		state.resolved = true
		return state.depth
	}
	if state.resolving {
		*findings = append(*findings, parentFinding(ParentInvalid, state.node, state.path))
		state.resolved = true
		return state.depth
	}

	state.resolving = true
	parents := pathIndex[state.node.Parent]
	switch len(parents) {
	case 0:
		*findings = append(*findings, parentFinding(ParentMissing, state.node, state.path))
	case 1:
		parentDepth := resolveLocalDepth(parents[0], states, pathIndex, findings)
		if parentDepth.Known {
			state.depth = OptionalDepth{Value: parentDepth.Value + 1, Known: true}
		}
	default:
		*findings = append(*findings, parentFinding(ParentAmbiguous, state.node, state.path))
	}

	state.resolving = false
	state.resolved = true
	return state.depth
}

func parentFinding(kind ParentFindingKind, node tscn.Node, path string) ParentFinding {
	return ParentFinding{
		Kind:     kind,
		NodeName: node.Name,
		NodePath: path,
		Parent:   node.Parent,
		Position: node.Position,
	}
}

func addLocalState(summary *LocalSummary, state localNodeState, index int, resources map[string]ExternalResource) {
	record := NodeRecord{
		Name:     state.node.Name,
		Path:     state.path,
		Parent:   state.node.Parent,
		Depth:    state.depth,
		Position: state.node.Position,
	}

	switch state.class {
	case classOrdinary:
		summary.Nodes = append(summary.Nodes, OrdinaryNode{
			NodeRecord:    record,
			Type:          state.node.Type,
			ShadowEnabled: cloneBool(state.node.ShadowEnabled),
		})
		summary.Metrics.Nodes++
		addLiteralMetrics(&summary.Metrics.MeshInstances, &summary.Metrics.Lights, &summary.Metrics.ShadowLights, state.node)
		includeKnownDepth(&summary.Metrics.TreeDepth, state.depth)
	case classMount:
		mount := buildMount(record, state.node, resources)
		summary.Mounts = append(summary.Mounts, mount)
		if index > 0 {
			summary.Metrics.SceneInstances++
		}
		includeKnownDepth(&summary.Metrics.TreeDepth, state.depth)
	case classInheritedRoot:
		root := InheritedRoot{
			NodeRecord: record,
			Reference:  copyReference(state.node.Instance),
			Candidate:  packedSceneCandidate(state.node.Instance, resources),
		}
		summary.InheritedRoot = &root
	case classOverrideStub:
		summary.OverrideStubs = append(summary.OverrideStubs, OverrideStub{NodeRecord: record})
	}
}

func buildMount(record NodeRecord, node tscn.Node, resources map[string]ExternalResource) InstanceMount {
	mount := InstanceMount{
		NodeRecord:  record,
		Reference:   copyReference(node.Instance),
		Placeholder: node.InstancePlaceholder,
	}

	switch {
	case node.InstancePlaceholder != "":
		mount.Kind = MountPlaceholder
	case node.Instance != nil && node.Instance.Kind == tscn.ResourceRefExternal:
		mount.Candidate = packedSceneCandidate(node.Instance, resources)
		if mount.Candidate != nil {
			mount.Kind = MountPackedSceneCandidate
		} else {
			mount.Kind = MountExternalResource
		}
	default:
		mount.Kind = MountSubResource
	}

	return mount
}

func packedSceneCandidate(reference *tscn.ResourceRef, resources map[string]ExternalResource) *ExternalResource {
	if reference == nil || reference.Kind != tscn.ResourceRefExternal {
		return nil
	}
	resource, exists := resources[reference.ID]
	if !exists || resource.Type != "PackedScene" {
		return nil
	}

	candidate := resource
	return &candidate
}

func copyReference(reference *tscn.ResourceRef) ResourceReference {
	if reference == nil {
		return ResourceReference{}
	}

	return ResourceReference{Kind: reference.Kind, ID: reference.ID}
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value

	return &cloned
}

func addLiteralMetrics(meshes, lights, shadowLights *int64, node tscn.Node) {
	if node.Type == "MeshInstance3D" {
		*meshes++
	}

	switch node.Type {
	case "DirectionalLight3D", "OmniLight3D", "SpotLight3D":
		*lights++
		if node.ShadowEnabled != nil && *node.ShadowEnabled {
			*shadowLights++
		}
	}
}

func includeKnownDepth(maximum *int64, depth OptionalDepth) {
	if depth.Known && depth.Value > *maximum {
		*maximum = depth.Value
	}
}

func sortParentFindings(findings []ParentFinding) {
	sort.SliceStable(findings, func(left, right int) bool {
		first := findings[left]
		second := findings[right]
		if first.Position.Line != second.Position.Line {
			return first.Position.Line < second.Position.Line
		}
		if first.Position.Column != second.Position.Column {
			return first.Position.Column < second.Position.Column
		}
		if first.Kind != second.Kind {
			return first.Kind < second.Kind
		}
		if first.NodeName != second.NodeName {
			return first.NodeName < second.NodeName
		}
		if first.NodePath != second.NodePath {
			return first.NodePath < second.NodePath
		}

		return first.Parent < second.Parent
	})
}
