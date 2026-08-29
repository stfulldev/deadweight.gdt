package analysis

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

// ResourceResolver is the secure path-resolution boundary required by the
// recursive analyzer.
type ResourceResolver interface {
	ResolveResource(fromScene, raw string) project.Resolution
}

// SceneOpener opens one fully resolved canonical scene identity.
type SceneOpener func(path project.ResolvedPath) (io.ReadCloser, error)

// SceneParser parses one opened scene using its stable display identity.
type SceneParser func(reader io.Reader, source string) (*tscn.Document, error)

type localSummaryBuilder func(document *tscn.Document) (LocalSummary, error)

// RecursiveAnalyzer expands nested scenes through injected resolution and
// parsing effects. Memoization state is allocated separately by every call.
type RecursiveAnalyzer struct {
	resolver  ResourceResolver
	opener    SceneOpener
	parser    SceneParser
	summarize localSummaryBuilder
}

// NewRecursiveAnalyzer validates and constructs a recursive scene analyzer.
func NewRecursiveAnalyzer(
	resolver ResourceResolver,
	opener SceneOpener,
	parser SceneParser,
) (*RecursiveAnalyzer, error) {
	if resolver == nil {
		return nil, errors.New("recursive analyzer requires a resource resolver")
	}
	if opener == nil {
		return nil, errors.New("recursive analyzer requires a scene opener")
	}
	if parser == nil {
		return nil, errors.New("recursive analyzer requires a scene parser")
	}

	return &RecursiveAnalyzer{
		resolver:  resolver,
		opener:    opener,
		parser:    parser,
		summarize: BuildLocalSummary,
	}, nil
}

// Analyze returns the deterministic recursive contribution and dependency
// graph for one canonical root scene. Fatal failures return a zero result.
func (analyzer *RecursiveAnalyzer) Analyze(root project.ResolvedPath) (RecursiveResult, error) {
	if analyzer == nil || analyzer.resolver == nil || analyzer.opener == nil || analyzer.parser == nil || analyzer.summarize == nil {
		return RecursiveResult{}, errors.New("recursive analyzer is not initialized")
	}
	if err := validateCanonicalRoot(root); err != nil {
		return RecursiveResult{}, err
	}

	state := invocationState{
		analyzer:   analyzer,
		cache:      newInvocationCache(),
		inProgress: make(map[string]project.ResolvedPath),
	}

	graph, graphResources, err := state.discoverGraph(root)
	if err != nil {
		return RecursiveResult{}, err
	}
	summary, err := state.expandScene(root)
	if err != nil {
		return RecursiveResult{}, err
	}
	summary.Dependencies = graphDependencyPaths(graph)
	summary.ExternalResources = mergeResourceIdentities(summary.ExternalResources, graphResources)
	parsedSceneFiles, err := state.cache.parsedSceneFiles()
	if err != nil {
		return RecursiveResult{}, err
	}
	result := RecursiveResult{
		Summary:          summary,
		Graph:            graph,
		ParsedSceneFiles: parsedSceneFiles,
	}

	return cloneRecursiveResult(result), nil
}

// Expand preserves the summary-only issue #9 API as a projection of Analyze.
func (analyzer *RecursiveAnalyzer) Expand(root project.ResolvedPath) (ExpandedSummary, error) {
	result, err := analyzer.Analyze(root)
	if err != nil {
		return ExpandedSummary{}, err
	}

	return cloneExpandedSummary(result.Summary), nil
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
	analyzer          *RecursiveAnalyzer
	cache             *invocationCache
	inProgress        map[string]project.ResolvedPath
	graphStates       map[string]graphVisitState
	graphStack        []project.ResolvedPath
	graphStackIndices map[string]int
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

type targetEvidence struct {
	Classification   TargetClassification
	ResolutionReason project.ResolutionReason
	ResourceID       string
	RawTarget        string
	TargetCanonical  string
	TargetDisplay    string
	TargetOriginal   string
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
	if cached, exists := state.cache.expandedSummary(path.Canonical); exists {
		return cached, nil
	}
	if active, exists := state.inProgress[path.Canonical]; exists {
		return ExpandedSummary{}, fmt.Errorf(
			"recursive expansion reached graph-unvalidated identity %q",
			active.Display,
		)
	}

	state.inProgress[path.Canonical] = path
	defer delete(state.inProgress, path.Canonical)

	local, err := state.loadLocalSummary(path)
	if err != nil {
		return ExpandedSummary{}, err
	}

	builder := newSummaryBuilder(local, path)
	resolvedResources := state.resolveSceneResources(path, local.ExternalResources)
	builder.unionResources(resourceIdentities(path, resolvedResources))
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
	state.cache.storeExpandedSummary(path.Canonical, result)

	return cloneExpandedSummary(result), nil
}

func (state *invocationState) loadLocalSummary(path project.ResolvedPath) (LocalSummary, error) {
	if local, localErr, exists := state.cache.localSummary(path.Canonical); exists {
		return local, localErr
	}

	document, err := state.cache.loadDocument(path, state.analyzer.opener, state.analyzer.parser)
	if err != nil {
		return LocalSummary{}, err
	}

	local, err := state.analyzer.summarize(document)
	if err != nil {
		state.cache.storeLocalSummaryError(path.Canonical, err)
		return LocalSummary{}, err
	}
	state.cache.storeLocalSummary(path.Canonical, local)

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
) map[string]resourceResolution {
	if cached, exists := state.cache.resources(path.Canonical); exists {
		return cached
	}

	resolved := make(map[string]resourceResolution, len(resources))
	for _, resource := range resources {
		resolution := state.analyzer.resolver.ResolveResource(path.Canonical, resource.Path)
		entry := resourceResolution{resource: resource, resolution: resolution}
		resolved[resource.ID] = entry
	}
	state.cache.storeResources(path.Canonical, resolved)

	return cloneResourceResolutions(resolved)
}

func resourceIdentities(
	path project.ResolvedPath,
	resources map[string]resourceResolution,
) []ResourceIdentity {
	identities := make(map[ResourceIdentity]struct{}, len(resources))
	for _, resolved := range resources {
		identity := ResourceIdentity{
			DeclaringScene: path.Canonical,
			ResourceID:     resolved.resource.ID,
			RawPath:        resolved.resource.Path,
		}
		if resolved.resolution.Resolved() {
			identity = ResourceIdentity{
				Resolved:  true,
				Canonical: resolved.resolution.Path.Canonical,
			}
		}
		identities[identity] = struct{}{}
	}

	return sortedResourceIdentities(identities)
}

func classifyMount(
	declaring project.ResolvedPath,
	mount InstanceMount,
	resources map[string]resourceResolution,
) (project.ResolvedPath, *UnresolvedInstance) {
	target, evidence := classifyTarget(declaring, mount.Reference, mount.Placeholder, resources)
	if evidence == nil {
		return target, nil
	}

	unresolved := unresolvedFromMount(
		declaring,
		mount,
		evidence.Classification,
		evidence.ResolutionReason,
	)
	unresolved.RawTarget = evidence.RawTarget
	unresolved.TargetCanonical = evidence.TargetCanonical
	unresolved.TargetDisplay = evidence.TargetDisplay
	unresolved.TargetOriginal = evidence.TargetOriginal

	return project.ResolvedPath{}, &unresolved
}

func classifyTarget(
	_ project.ResolvedPath,
	reference ResourceReference,
	placeholder string,
	resources map[string]resourceResolution,
) (project.ResolvedPath, *targetEvidence) {
	switch {
	case placeholder != "":
		return project.ResolvedPath{}, &targetEvidence{
			Classification: TargetPlaceholder,
			RawTarget:      placeholder,
		}
	case reference.Kind == tscn.ResourceRefInternal:
		return project.ResolvedPath{}, &targetEvidence{
			Classification: TargetSubResource,
			ResourceID:     reference.ID,
			RawTarget:      reference.ID,
		}
	case reference.Kind != tscn.ResourceRefExternal:
		return project.ResolvedPath{}, &targetEvidence{
			Classification: TargetMissingExternalResource,
			ResourceID:     reference.ID,
		}
	}

	resource, exists := resources[reference.ID]
	if !exists {
		return project.ResolvedPath{}, &targetEvidence{
			Classification: TargetMissingExternalResource,
			ResourceID:     reference.ID,
		}
	}

	if !resource.resolution.Resolved() {
		return project.ResolvedPath{}, &targetEvidence{
			Classification:   TargetUnresolvedPath,
			ResolutionReason: resource.resolution.Reason,
			ResourceID:       reference.ID,
			RawTarget:        resource.resource.Path,
			TargetOriginal:   resource.resolution.Path.Original,
		}
	}

	target := resource.resolution.Path
	if target.Original == "" {
		target.Original = resource.resource.Path
	}
	extension := filepath.Ext(target.Canonical)
	switch extension {
	case ".tscn":
		return target, nil
	case ".glb", ".gltf", ".blend", ".scn":
		return project.ResolvedPath{}, resolvedTargetEvidence(
			TargetImportedScene,
			reference.ID,
			resource.resource.Path,
			target,
		)
	default:
		return project.ResolvedPath{}, resolvedTargetEvidence(
			TargetUnsupportedScene,
			reference.ID,
			resource.resource.Path,
			target,
		)
	}
}

func resolvedTargetEvidence(
	classification TargetClassification,
	resourceID string,
	raw string,
	target project.ResolvedPath,
) *targetEvidence {
	return &targetEvidence{
		Classification:   classification,
		ResolutionReason: project.ResolutionResolved,
		ResourceID:       resourceID,
		RawTarget:        raw,
		TargetCanonical:  target.Canonical,
		TargetDisplay:    target.Display,
		TargetOriginal:   target.Original,
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
	nodes, err := checkedAdd(builder.summary.Metrics.Nodes, evidence.Occurrences)
	if err != nil {
		return err
	}
	unresolved, err := checkedAdd(
		builder.summary.Coverage.Unresolved,
		evidence.Occurrences,
	)
	if err != nil {
		return err
	}

	next := cloneExpandedSummary(builder.summary)
	next.Metrics.Nodes = nodes
	next.Coverage.Unresolved = unresolved
	if !evidence.MountDepth.Known {
		next.DepthPartial = true
	}
	next.Unresolved = append(next.Unresolved, evidence)
	builder.summary = next

	return nil
}

func (builder *summaryBuilder) addInherited(
	declaring project.ResolvedPath,
	mount InstanceMount,
	inherited *inheritedSceneError,
) error {
	nodes, err := checkedAdd(builder.summary.Metrics.Nodes, 1)
	if err != nil {
		return err
	}
	unresolved, err := checkedAdd(builder.summary.Coverage.Unresolved, 1)
	if err != nil {
		return err
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

	next := cloneExpandedSummary(builder.summary)
	next.Metrics.Nodes = nodes
	next.Coverage.Unresolved = unresolved
	if !mount.Depth.Known {
		next.DepthPartial = true
	}
	next.InheritedTargets = append(next.InheritedTargets, evidence)
	builder.summary = next

	return nil
}

func (builder *summaryBuilder) applyResolved(
	childPath project.ResolvedPath,
	key resolvedApplicationKey,
	multiplicity int64,
	child ExpandedSummary,
) error {
	next := cloneExpandedSummary(builder.summary)
	fields := []struct {
		target *int64
		value  int64
	}{
		{target: &next.Metrics.Nodes, value: child.Metrics.Nodes},
		{target: &next.Metrics.SceneInstances, value: child.Metrics.SceneInstances},
		{target: &next.Metrics.MeshInstances, value: child.Metrics.MeshInstances},
		{target: &next.Metrics.Lights, value: child.Metrics.Lights},
		{target: &next.Metrics.ShadowLights, value: child.Metrics.ShadowLights},
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
	next.Coverage.Resolved, err = checkedAdd(
		next.Coverage.Resolved,
		resolvedContribution,
	)
	if err != nil {
		return err
	}
	unresolvedContribution, err := checkedMultiply(multiplicity, child.Coverage.Unresolved)
	if err != nil {
		return err
	}
	next.Coverage.Unresolved, err = checkedAdd(
		next.Coverage.Unresolved,
		unresolvedContribution,
	)
	if err != nil {
		return err
	}

	if !key.known || child.Metrics.TreeDepth <= 0 {
		next.DepthPartial = true
	} else {
		candidate, depthErr := checkedDepth(key.depth, child.Metrics.TreeDepth)
		if depthErr != nil {
			return depthErr
		}
		if candidate > next.Metrics.TreeDepth {
			next.Metrics.TreeDepth = candidate
		}
	}
	if child.DepthPartial {
		next.DepthPartial = true
	}

	for _, unresolved := range child.Unresolved {
		unresolved.Occurrences, err = checkedMultiply(unresolved.Occurrences, multiplicity)
		if err != nil {
			return err
		}
		next.Unresolved = append(next.Unresolved, unresolved)
	}
	for _, inherited := range child.InheritedTargets {
		inherited.Occurrences, err = checkedMultiply(inherited.Occurrences, multiplicity)
		if err != nil {
			return err
		}
		next.InheritedTargets = append(next.InheritedTargets, inherited)
	}
	for _, finding := range child.ParentFindings {
		finding.Occurrences, err = checkedMultiply(finding.Occurrences, multiplicity)
		if err != nil {
			return err
		}
		next.ParentFindings = append(next.ParentFindings, finding)
	}

	builder.summary = next
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
	return sortedResourceIdentities(builder.resources)
}

func sortedResourceIdentities(resourcesSet map[ResourceIdentity]struct{}) []ResourceIdentity {
	resources := make([]ResourceIdentity, 0, len(resourcesSet))
	for identity := range resourcesSet {
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

func mergeResourceIdentities(groups ...[]ResourceIdentity) []ResourceIdentity {
	resources := make(map[ResourceIdentity]struct{})
	for _, group := range groups {
		for _, identity := range group {
			resources[identity] = struct{}{}
		}
	}

	return sortedResourceIdentities(resources)
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
