package analysis

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

// ResourceResolver is the secure path-resolution boundary required by the
// recursive analyzer.
type ResourceResolver interface {
	ResolveResource(fromScene, raw string) project.Resolution
}

// SceneLoader loads and parses one fully resolved canonical scene identity.
type SceneLoader func(path project.ResolvedPath) (*tscn.Document, error)

type localSummaryBuilder func(document *tscn.Document) (LocalSummary, error)

// RecursiveAnalyzer expands nested scenes through injected resolution and
// parsing effects. Memoization state is allocated separately by every call.
type RecursiveAnalyzer struct {
	resolver  ResourceResolver
	loader    SceneLoader
	summarize localSummaryBuilder
}

// NewRecursiveAnalyzer validates and constructs a recursive scene analyzer.
func NewRecursiveAnalyzer(resolver ResourceResolver, loader SceneLoader) (*RecursiveAnalyzer, error) {
	if resolver == nil {
		return nil, errors.New("recursive analyzer requires a resource resolver")
	}
	if loader == nil {
		return nil, errors.New("recursive analyzer requires a scene loader")
	}

	return &RecursiveAnalyzer{
		resolver:  resolver,
		loader:    loader,
		summarize: BuildLocalSummary,
	}, nil
}

// Expand returns the deterministic recursive contribution of one canonical
// root scene. Fatal failures return a zero summary.
func (analyzer *RecursiveAnalyzer) Expand(root project.ResolvedPath) (ExpandedSummary, error) {
	if analyzer == nil || analyzer.resolver == nil || analyzer.loader == nil || analyzer.summarize == nil {
		return ExpandedSummary{}, errors.New("recursive analyzer is not initialized")
	}
	if err := validateCanonicalRoot(root); err != nil {
		return ExpandedSummary{}, err
	}

	state := invocationState{
		analyzer:       analyzer,
		documents:      make(map[string]*tscn.Document),
		documentErrors: make(map[string]error),
		localSummaries: make(map[string]LocalSummary),
		summaries:      make(map[string]ExpandedSummary),
		inProgress:     make(map[string]project.ResolvedPath),
	}

	summary, err := state.expandScene(root)
	if err != nil {
		return ExpandedSummary{}, err
	}

	return cloneExpandedSummary(summary), nil
}

func validateCanonicalRoot(root project.ResolvedPath) error {
	if root.Canonical == "" || !filepath.IsAbs(root.Canonical) || filepath.Clean(root.Canonical) != root.Canonical {
		return fmt.Errorf("recursive root must have a clean absolute canonical path, got %q", root.Canonical)
	}
	if filepath.Ext(root.Canonical) != ".tscn" {
		return fmt.Errorf("recursive root must be an exact .tscn path, got %q", root.Canonical)
	}
	if root.Display == "" {
		return errors.New("recursive root must retain a display path")
	}

	return nil
}

var _ ResourceResolver = project.Resolver{}

type invocationState struct {
	analyzer       *RecursiveAnalyzer
	documents      map[string]*tscn.Document
	documentErrors map[string]error
	localSummaries map[string]LocalSummary
	summaries      map[string]ExpandedSummary
	inProgress     map[string]project.ResolvedPath
}

type sceneLoadError struct {
	path  project.ResolvedPath
	cause error
}

func (err *sceneLoadError) Error() string {
	return fmt.Sprintf("load scene %q: %v", err.path.Display, err.cause)
}

func (err *sceneLoadError) Unwrap() error {
	return err.cause
}

type inheritedSceneError struct {
	path      project.ResolvedPath
	root      InheritedRoot
	resources []ResourceIdentity
	base      resourceResolution
}

func (err *inheritedSceneError) Error() string {
	return fmt.Sprintf("nested scene %q has an inherited root", err.path.Display)
}

type resourceResolution struct {
	resource   ExternalResource
	resolution project.Resolution
}

type resolvedApplicationKey struct {
	canonical string
	depth     int64
	known     bool
}

type resolvedApplication struct {
	path   project.ResolvedPath
	mounts []InstanceMount
	count  int64
}

type summaryBuilder struct {
	summary      ExpandedSummary
	resources    map[ResourceIdentity]struct{}
	dependencies map[string]struct{}
}

func (state *invocationState) expandScene(path project.ResolvedPath) (ExpandedSummary, error) {
	if cached, exists := state.summaries[path.Canonical]; exists {
		return cloneExpandedSummary(cached), nil
	}
	if active, exists := state.inProgress[path.Canonical]; exists {
		return ExpandedSummary{}, &RecursiveReferenceError{
			Canonical: active.Canonical,
			Display:   active.Display,
		}
	}

	state.inProgress[path.Canonical] = path
	defer delete(state.inProgress, path.Canonical)

	local, err := state.loadLocalSummary(path)
	if err != nil {
		return ExpandedSummary{}, err
	}

	builder := newSummaryBuilder(local, path)
	resolvedResources := state.resolveSceneResources(path, local.ExternalResources, builder)
	if local.InheritedRoot != nil {
		base := resolvedResources[local.InheritedRoot.Reference.ID]
		return ExpandedSummary{}, &inheritedSceneError{
			path:      path,
			root:      cloneInheritedRoot(*local.InheritedRoot),
			resources: builder.sortedResources(),
			base:      base,
		}
	}

	applications := make(map[resolvedApplicationKey]*resolvedApplication)
	for _, mount := range local.Mounts {
		resolved, unresolved := classifyMount(path, mount, resolvedResources)
		if unresolved != nil {
			if err := builder.addUnresolved(*unresolved); err != nil {
				return ExpandedSummary{}, err
			}
			continue
		}

		key := resolvedApplicationKey{
			canonical: resolved.Canonical,
			depth:     mount.Depth.Value,
			known:     mount.Depth.Known,
		}
		application := applications[key]
		if application == nil {
			application = &resolvedApplication{path: resolved}
			applications[key] = application
		}
		application.mounts = append(application.mounts, cloneInstanceMount(mount))
		application.count, err = checkedAdd(application.count, 1)
		if err != nil {
			return ExpandedSummary{}, err
		}
	}

	for _, key := range sortedApplicationKeys(applications) {
		application := applications[key]
		child, expandErr := state.expandScene(application.path)
		if expandErr != nil {
			var inherited *inheritedSceneError
			var unavailable *sceneLoadError
			switch {
			case errors.As(expandErr, &inherited):
				builder.unionResources(inherited.resources)
				builder.dependencies[inherited.path.Canonical] = struct{}{}
				for _, mount := range application.mounts {
					if err := builder.addInherited(path, mount, inherited); err != nil {
						return ExpandedSummary{}, err
					}
				}
				continue
			case errors.As(expandErr, &unavailable):
				for _, mount := range application.mounts {
					evidence := unresolvedFromMount(path, mount, TargetUnavailableScene, project.ResolutionResolved)
					evidence.TargetCanonical = application.path.Canonical
					evidence.TargetDisplay = application.path.Display
					evidence.TargetOriginal = application.path.Original
					evidence.RawTarget = application.path.Original
					if err := builder.addUnresolved(evidence); err != nil {
						return ExpandedSummary{}, err
					}
				}
				continue
			default:
				return ExpandedSummary{}, expandErr
			}
		}

		if err := builder.applyResolved(application.path, key, application.count, child); err != nil {
			return ExpandedSummary{}, err
		}
	}

	result := builder.finish()
	state.summaries[path.Canonical] = cloneExpandedSummary(result)

	return cloneExpandedSummary(result), nil
}

func (state *invocationState) loadLocalSummary(path project.ResolvedPath) (LocalSummary, error) {
	if local, exists := state.localSummaries[path.Canonical]; exists {
		return cloneLocalSummary(local), nil
	}
	if loadErr, exists := state.documentErrors[path.Canonical]; exists {
		return LocalSummary{}, loadErr
	}

	document := state.documents[path.Canonical]
	if document == nil {
		loaded, err := state.analyzer.loader(path)
		if err != nil {
			if code, coded := diagnostic.CodeOf(err); coded && code == diagnostic.CodeInvalidTSCNRoot {
				state.documentErrors[path.Canonical] = err
				return LocalSummary{}, err
			}

			loadErr := &sceneLoadError{path: path, cause: err}
			state.documentErrors[path.Canonical] = loadErr
			return LocalSummary{}, loadErr
		}
		document = loaded
		state.documents[path.Canonical] = loaded
	}

	local, err := state.analyzer.summarize(document)
	if err != nil {
		return LocalSummary{}, err
	}
	state.localSummaries[path.Canonical] = cloneLocalSummary(local)

	return cloneLocalSummary(local), nil
}

func newSummaryBuilder(local LocalSummary, path project.ResolvedPath) *summaryBuilder {
	summary := ExpandedSummary{
		Metrics:      local.Metrics,
		DepthPartial: local.DepthPartial,
	}
	summary.Metrics.ExternalResources = 0
	summary.Metrics.SceneDependencies = 0
	for _, finding := range local.Findings {
		summary.ParentFindings = append(summary.ParentFindings, SceneParentFinding{
			DeclaringScene:   path.Canonical,
			DeclaringDisplay: path.Display,
			Finding:          finding,
			Occurrences:      1,
		})
	}

	return &summaryBuilder{
		summary:      summary,
		resources:    make(map[ResourceIdentity]struct{}),
		dependencies: make(map[string]struct{}),
	}
}

func (state *invocationState) resolveSceneResources(
	path project.ResolvedPath,
	resources []ExternalResource,
	builder *summaryBuilder,
) map[string]resourceResolution {
	resolved := make(map[string]resourceResolution, len(resources))
	for _, resource := range resources {
		resolution := state.analyzer.resolver.ResolveResource(path.Canonical, resource.Path)
		entry := resourceResolution{resource: resource, resolution: resolution}
		resolved[resource.ID] = entry

		identity := ResourceIdentity{
			DeclaringScene: path.Canonical,
			ResourceID:     resource.ID,
			RawPath:        resource.Path,
		}
		if resolution.Resolved() {
			identity = ResourceIdentity{Resolved: true, Canonical: resolution.Path.Canonical}
		}
		builder.resources[identity] = struct{}{}
	}

	return resolved
}

func classifyMount(
	declaring project.ResolvedPath,
	mount InstanceMount,
	resources map[string]resourceResolution,
) (project.ResolvedPath, *UnresolvedInstance) {
	switch {
	case mount.Kind == MountPlaceholder:
		evidence := unresolvedFromMount(declaring, mount, TargetPlaceholder, "")
		evidence.RawTarget = mount.Placeholder
		return project.ResolvedPath{}, &evidence
	case mount.Reference.Kind == tscn.ResourceRefInternal || mount.Kind == MountSubResource:
		evidence := unresolvedFromMount(declaring, mount, TargetSubResource, "")
		return project.ResolvedPath{}, &evidence
	case mount.Reference.Kind != tscn.ResourceRefExternal:
		evidence := unresolvedFromMount(declaring, mount, TargetMissingExternalResource, "")
		return project.ResolvedPath{}, &evidence
	}

	resource, exists := resources[mount.Reference.ID]
	if !exists {
		evidence := unresolvedFromMount(declaring, mount, TargetMissingExternalResource, "")
		return project.ResolvedPath{}, &evidence
	}

	if !resource.resolution.Resolved() {
		evidence := unresolvedFromMount(declaring, mount, TargetUnresolvedPath, resource.resolution.Reason)
		evidence.RawTarget = resource.resource.Path
		evidence.TargetOriginal = resource.resolution.Path.Original
		return project.ResolvedPath{}, &evidence
	}

	target := resource.resolution.Path
	extension := filepath.Ext(target.Canonical)
	switch extension {
	case ".tscn":
		return target, nil
	case ".glb", ".gltf", ".blend", ".scn":
		evidence := unresolvedFromMount(declaring, mount, TargetImportedScene, project.ResolutionResolved)
		populateResolvedTarget(&evidence, resource.resource.Path, target)
		return project.ResolvedPath{}, &evidence
	default:
		evidence := unresolvedFromMount(declaring, mount, TargetUnsupportedScene, project.ResolutionResolved)
		populateResolvedTarget(&evidence, resource.resource.Path, target)
		return project.ResolvedPath{}, &evidence
	}
}

func unresolvedFromMount(
	declaring project.ResolvedPath,
	mount InstanceMount,
	classification TargetClassification,
	reason project.ResolutionReason,
) UnresolvedInstance {
	return UnresolvedInstance{
		Classification:   classification,
		ResolutionReason: reason,
		DeclaringScene:   declaring.Canonical,
		DeclaringDisplay: declaring.Display,
		ResourceID:       mount.Reference.ID,
		MountName:        mount.Name,
		MountPath:        mount.Path,
		MountDepth:       mount.Depth,
		Position:         mount.Position,
		Occurrences:      1,
	}
}

func populateResolvedTarget(evidence *UnresolvedInstance, raw string, target project.ResolvedPath) {
	evidence.RawTarget = raw
	evidence.TargetCanonical = target.Canonical
	evidence.TargetDisplay = target.Display
	evidence.TargetOriginal = target.Original
}

func sortedApplicationKeys(applications map[resolvedApplicationKey]*resolvedApplication) []resolvedApplicationKey {
	keys := make([]resolvedApplicationKey, 0, len(applications))
	for key := range applications {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(left, right int) bool {
		if keys[left].canonical != keys[right].canonical {
			return keys[left].canonical < keys[right].canonical
		}
		if keys[left].known != keys[right].known {
			return !keys[left].known
		}

		return keys[left].depth < keys[right].depth
	})

	return keys
}

func (builder *summaryBuilder) addUnresolved(evidence UnresolvedInstance) error {
	var err error
	builder.summary.Metrics.Nodes, err = checkedAdd(builder.summary.Metrics.Nodes, evidence.Occurrences)
	if err != nil {
		return err
	}
	builder.summary.Coverage.Unresolved, err = checkedAdd(
		builder.summary.Coverage.Unresolved,
		evidence.Occurrences,
	)
	if err != nil {
		return err
	}
	if !evidence.MountDepth.Known {
		builder.summary.DepthPartial = true
	}
	builder.summary.Unresolved = append(builder.summary.Unresolved, evidence)

	return nil
}

func (builder *summaryBuilder) addInherited(
	declaring project.ResolvedPath,
	mount InstanceMount,
	inherited *inheritedSceneError,
) error {
	var err error
	builder.summary.Metrics.Nodes, err = checkedAdd(builder.summary.Metrics.Nodes, 1)
	if err != nil {
		return err
	}
	builder.summary.Coverage.Unresolved, err = checkedAdd(builder.summary.Coverage.Unresolved, 1)
	if err != nil {
		return err
	}
	if !mount.Depth.Known {
		builder.summary.DepthPartial = true
	}

	evidence := InheritedTarget{
		Classification:        TargetInheritedScene,
		DeclaringScene:        declaring.Canonical,
		DeclaringDisplay:      declaring.Display,
		TargetCanonical:       inherited.path.Canonical,
		TargetDisplay:         inherited.path.Display,
		TargetOriginal:        inherited.path.Original,
		BaseResourceID:        inherited.root.Reference.ID,
		MountName:             mount.Name,
		MountPath:             mount.Path,
		MountDepth:            mount.Depth,
		MountPosition:         mount.Position,
		InheritedRootPosition: inherited.root.Position,
		Occurrences:           1,
	}
	if inherited.base.resource.ID != "" {
		evidence.BaseRawTarget = inherited.base.resource.Path
	}
	if inherited.base.resolution.Resolved() {
		evidence.BaseCanonical = inherited.base.resolution.Path.Canonical
		evidence.BaseDisplay = inherited.base.resolution.Path.Display
	}
	builder.summary.InheritedTargets = append(builder.summary.InheritedTargets, evidence)

	return nil
}

func (builder *summaryBuilder) applyResolved(
	childPath project.ResolvedPath,
	key resolvedApplicationKey,
	multiplicity int64,
	child ExpandedSummary,
) error {
	fields := []struct {
		target *int64
		value  int64
	}{
		{target: &builder.summary.Metrics.Nodes, value: child.Metrics.Nodes},
		{target: &builder.summary.Metrics.SceneInstances, value: child.Metrics.SceneInstances},
		{target: &builder.summary.Metrics.MeshInstances, value: child.Metrics.MeshInstances},
		{target: &builder.summary.Metrics.Lights, value: child.Metrics.Lights},
		{target: &builder.summary.Metrics.ShadowLights, value: child.Metrics.ShadowLights},
	}
	for _, field := range fields {
		contribution, err := checkedMultiply(multiplicity, field.value)
		if err != nil {
			return err
		}
		*field.target, err = checkedAdd(*field.target, contribution)
		if err != nil {
			return err
		}
	}

	resolvedPerOccurrence, err := checkedAdd(1, child.Coverage.Resolved)
	if err != nil {
		return err
	}
	resolvedContribution, err := checkedMultiply(multiplicity, resolvedPerOccurrence)
	if err != nil {
		return err
	}
	builder.summary.Coverage.Resolved, err = checkedAdd(
		builder.summary.Coverage.Resolved,
		resolvedContribution,
	)
	if err != nil {
		return err
	}
	unresolvedContribution, err := checkedMultiply(multiplicity, child.Coverage.Unresolved)
	if err != nil {
		return err
	}
	builder.summary.Coverage.Unresolved, err = checkedAdd(
		builder.summary.Coverage.Unresolved,
		unresolvedContribution,
	)
	if err != nil {
		return err
	}

	if !key.known || child.Metrics.TreeDepth <= 0 {
		builder.summary.DepthPartial = true
	} else {
		candidate, depthErr := checkedDepth(key.depth, child.Metrics.TreeDepth)
		if depthErr != nil {
			return depthErr
		}
		if candidate > builder.summary.Metrics.TreeDepth {
			builder.summary.Metrics.TreeDepth = candidate
		}
	}
	if child.DepthPartial {
		builder.summary.DepthPartial = true
	}

	for _, unresolved := range child.Unresolved {
		unresolved.Occurrences, err = checkedMultiply(unresolved.Occurrences, multiplicity)
		if err != nil {
			return err
		}
		builder.summary.Unresolved = append(builder.summary.Unresolved, unresolved)
	}
	for _, inherited := range child.InheritedTargets {
		inherited.Occurrences, err = checkedMultiply(inherited.Occurrences, multiplicity)
		if err != nil {
			return err
		}
		builder.summary.InheritedTargets = append(builder.summary.InheritedTargets, inherited)
	}
	for _, finding := range child.ParentFindings {
		finding.Occurrences, err = checkedMultiply(finding.Occurrences, multiplicity)
		if err != nil {
			return err
		}
		builder.summary.ParentFindings = append(builder.summary.ParentFindings, finding)
	}

	builder.dependencies[childPath.Canonical] = struct{}{}
	for _, dependency := range child.Dependencies {
		builder.dependencies[dependency] = struct{}{}
	}
	builder.unionResources(child.ExternalResources)

	return nil
}

func (builder *summaryBuilder) unionResources(resources []ResourceIdentity) {
	for _, identity := range resources {
		builder.resources[identity] = struct{}{}
	}
}

func (builder *summaryBuilder) sortedResources() []ResourceIdentity {
	resources := make([]ResourceIdentity, 0, len(builder.resources))
	for identity := range builder.resources {
		resources = append(resources, identity)
	}
	sort.Slice(resources, func(left, right int) bool {
		first := resources[left]
		second := resources[right]
		if first.Resolved != second.Resolved {
			return first.Resolved
		}
		if first.Canonical != second.Canonical {
			return first.Canonical < second.Canonical
		}
		if first.DeclaringScene != second.DeclaringScene {
			return first.DeclaringScene < second.DeclaringScene
		}
		if first.ResourceID != second.ResourceID {
			return first.ResourceID < second.ResourceID
		}

		return first.RawPath < second.RawPath
	})

	return resources
}

func (builder *summaryBuilder) finish() ExpandedSummary {
	builder.summary.ExternalResources = builder.sortedResources()
	builder.summary.Dependencies = make([]string, 0, len(builder.dependencies))
	for dependency := range builder.dependencies {
		builder.summary.Dependencies = append(builder.summary.Dependencies, dependency)
	}
	sort.Strings(builder.summary.Dependencies)
	sortUnresolved(builder.summary.Unresolved)
	sortInheritedTargets(builder.summary.InheritedTargets)
	sortSceneParentFindings(builder.summary.ParentFindings)

	return cloneExpandedSummary(builder.summary)
}

func sortUnresolved(evidence []UnresolvedInstance) {
	sort.Slice(evidence, func(left, right int) bool {
		first := evidence[left]
		second := evidence[right]
		if first.DeclaringScene != second.DeclaringScene {
			return first.DeclaringScene < second.DeclaringScene
		}
		if first.Classification != second.Classification {
			return first.Classification < second.Classification
		}
		if first.ResolutionReason != second.ResolutionReason {
			return first.ResolutionReason < second.ResolutionReason
		}
		if first.ResourceID != second.ResourceID {
			return first.ResourceID < second.ResourceID
		}
		if first.RawTarget != second.RawTarget {
			return first.RawTarget < second.RawTarget
		}
		if first.MountPath != second.MountPath {
			return first.MountPath < second.MountPath
		}
		if first.Position.Line != second.Position.Line {
			return first.Position.Line < second.Position.Line
		}
		if first.Position.Column != second.Position.Column {
			return first.Position.Column < second.Position.Column
		}

		return first.Occurrences < second.Occurrences
	})
}

func sortInheritedTargets(evidence []InheritedTarget) {
	sort.Slice(evidence, func(left, right int) bool {
		first := evidence[left]
		second := evidence[right]
		if first.DeclaringScene != second.DeclaringScene {
			return first.DeclaringScene < second.DeclaringScene
		}
		if first.TargetCanonical != second.TargetCanonical {
			return first.TargetCanonical < second.TargetCanonical
		}
		if first.MountPath != second.MountPath {
			return first.MountPath < second.MountPath
		}
		if first.MountPosition.Line != second.MountPosition.Line {
			return first.MountPosition.Line < second.MountPosition.Line
		}
		if first.MountPosition.Column != second.MountPosition.Column {
			return first.MountPosition.Column < second.MountPosition.Column
		}

		return first.Occurrences < second.Occurrences
	})
}

func sortSceneParentFindings(findings []SceneParentFinding) {
	sort.Slice(findings, func(left, right int) bool {
		first := findings[left]
		second := findings[right]
		if first.DeclaringScene != second.DeclaringScene {
			return first.DeclaringScene < second.DeclaringScene
		}
		if first.Finding.Position.Line != second.Finding.Position.Line {
			return first.Finding.Position.Line < second.Finding.Position.Line
		}
		if first.Finding.Position.Column != second.Finding.Position.Column {
			return first.Finding.Position.Column < second.Finding.Position.Column
		}
		if first.Finding.Kind != second.Finding.Kind {
			return first.Finding.Kind < second.Finding.Kind
		}

		return first.Occurrences < second.Occurrences
	})
}

func cloneExpandedSummary(summary ExpandedSummary) ExpandedSummary {
	cloned := summary
	cloned.ExternalResources = append([]ResourceIdentity(nil), summary.ExternalResources...)
	cloned.Dependencies = append([]string(nil), summary.Dependencies...)
	cloned.Unresolved = append([]UnresolvedInstance(nil), summary.Unresolved...)
	cloned.InheritedTargets = append([]InheritedTarget(nil), summary.InheritedTargets...)
	cloned.ParentFindings = append([]SceneParentFinding(nil), summary.ParentFindings...)

	return cloned
}

func cloneLocalSummary(summary LocalSummary) LocalSummary {
	cloned := summary
	cloned.Nodes = append([]OrdinaryNode(nil), summary.Nodes...)
	for index := range cloned.Nodes {
		cloned.Nodes[index].ShadowEnabled = cloneBool(cloned.Nodes[index].ShadowEnabled)
	}
	cloned.Mounts = make([]InstanceMount, len(summary.Mounts))
	for index, mount := range summary.Mounts {
		cloned.Mounts[index] = cloneInstanceMount(mount)
	}
	if summary.InheritedRoot != nil {
		root := cloneInheritedRoot(*summary.InheritedRoot)
		cloned.InheritedRoot = &root
	}
	cloned.OverrideStubs = append([]OverrideStub(nil), summary.OverrideStubs...)
	cloned.ExternalResources = append([]ExternalResource(nil), summary.ExternalResources...)
	cloned.Findings = append([]ParentFinding(nil), summary.Findings...)

	return cloned
}

func cloneInstanceMount(mount InstanceMount) InstanceMount {
	cloned := mount
	if mount.Candidate != nil {
		candidate := *mount.Candidate
		cloned.Candidate = &candidate
	}

	return cloned
}

func cloneInheritedRoot(root InheritedRoot) InheritedRoot {
	cloned := root
	if root.Candidate != nil {
		candidate := *root.Candidate
		cloned.Candidate = &candidate
	}

	return cloned
}
