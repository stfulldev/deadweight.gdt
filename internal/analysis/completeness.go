package analysis

import (
	"fmt"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

type completionResult struct {
	Status      AnalysisStatus
	Reliability Reliability
	Coverage    Coverage
	Diagnostics []diagnostic.Diagnostic
}

type diagnosticGroupKey struct {
	code           diagnostic.Code
	file           string
	resource       string
	classification TargetClassification
	reason         project.ResolutionReason
	message        string
}

type diagnosticGroup struct {
	item diagnostic.Diagnostic
}

type resourceDeclarationKey struct {
	declaringScene string
	resourceID     string
}

func finalizeCompleteness(
	summary ExpandedSummary,
	graph DependencyGraph,
	parsedSceneFiles int64,
) (completionResult, error) {
	coverage := Coverage{
		ResolvedSceneInstances:   summary.Coverage.Resolved,
		UnresolvedSceneInstances: summary.Coverage.Unresolved,
		ParsedSceneFiles:         parsedSceneFiles,
	}
	if err := coverage.Validate(); err != nil {
		return completionResult{}, err
	}
	instanceCoverage, err := checkedAdd(
		coverage.ResolvedSceneInstances,
		coverage.UnresolvedSceneInstances,
	)
	if err != nil {
		return completionResult{}, err
	}
	if instanceCoverage != summary.Metrics.SceneInstances {
		return completionResult{}, fmt.Errorf(
			"scene-instance coverage %d does not match metric %d",
			instanceCoverage,
			summary.Metrics.SceneInstances,
		)
	}

	displays := graphDisplayPaths(graph)
	groups := make(map[diagnosticGroupKey]diagnosticGroup)
	coveredResources := make(map[resourceDeclarationKey]struct{})
	lowerBound := false
	approximate := false

	for _, evidence := range summary.Unresolved {
		if err := validateUnresolvedEvidence(evidence); err != nil {
			return completionResult{}, err
		}
		item, key := unresolvedDiagnostic(evidence)
		if err := addDiagnosticGroup(groups, key, item, evidence.Occurrences); err != nil {
			return completionResult{}, err
		}
		lowerBound = true
		if evidence.ResourceID != "" {
			coveredResources[resourceDeclarationKey{
				declaringScene: evidence.DeclaringScene,
				resourceID:     evidence.ResourceID,
			}] = struct{}{}
		}
	}

	for _, evidence := range summary.InheritedTargets {
		if !evidence.Classification.Valid() ||
			evidence.Classification != TargetInheritedScene ||
			evidence.Occurrences <= 0 ||
			(evidence.BaseClassification != "" && !evidence.BaseClassification.Valid()) ||
			(evidence.BaseResolutionReason != "" && !evidence.BaseResolutionReason.Valid()) {
			return completionResult{}, fmt.Errorf("invalid inherited-scene evidence: %#v", evidence)
		}
		coverage.InheritedScenes, err = checkedAdd(coverage.InheritedScenes, evidence.Occurrences)
		if err != nil {
			return completionResult{}, err
		}
		item, key := inheritedDiagnostic(evidence)
		if err := addDiagnosticGroup(groups, key, item, evidence.Occurrences); err != nil {
			return completionResult{}, err
		}
		if evidence.BaseResourceID != "" {
			coveredResources[resourceDeclarationKey{
				declaringScene: evidence.DeclaringScene,
				resourceID:     evidence.BaseResourceID,
			}] = struct{}{}
		}
		approximate = true
	}

	for _, finding := range summary.ParentFindings {
		if !validParentFinding(finding) {
			return completionResult{}, fmt.Errorf("invalid parent finding: %#v", finding)
		}
		item, key := parentDiagnostic(finding)
		if err := addDiagnosticGroup(groups, key, item, finding.Occurrences); err != nil {
			return completionResult{}, err
		}
		lowerBound = true
	}
	if summary.DepthPartial && len(summary.ParentFindings) == 0 {
		file := displayPath(displays, graph.RootCanonical)
		item := diagnostic.Diagnostic{
			Code:     diagnostic.CodeUnsupportedParent,
			Severity: diagnostic.SeverityWarning,
			Message:  "tree depth is partial because parent semantics are unsupported",
			File:     file,
		}
		key := diagnosticKey(
			item,
			"depth_partial",
			"",
			TargetClassification("unsupported_parent"),
		)
		if err := addDiagnosticGroup(groups, key, item, 1); err != nil {
			return completionResult{}, err
		}
		lowerBound = true
	}

	for _, identity := range summary.ExternalResources {
		if identity.Resolved {
			if identity.Canonical == "" {
				return completionResult{}, fmt.Errorf("resolved resource identity has no canonical path")
			}
			continue
		}
		if identity.DeclaringScene == "" || identity.ResourceID == "" ||
			!identity.ResolutionReason.Valid() ||
			identity.ResolutionReason == project.ResolutionResolved {
			return completionResult{}, fmt.Errorf("invalid unresolved resource identity: %#v", identity)
		}
		declaration := resourceDeclarationKey{
			declaringScene: identity.DeclaringScene,
			resourceID:     identity.ResourceID,
		}
		if _, covered := coveredResources[declaration]; covered {
			lowerBound = true
			continue
		}
		item, key := resourceDiagnostic(identity, displayPath(displays, identity.DeclaringScene))
		if err := addDiagnosticGroup(groups, key, item, 1); err != nil {
			return completionResult{}, err
		}
		lowerBound = true
	}

	if err := coverage.Validate(); err != nil {
		return completionResult{}, err
	}
	diagnostics, err := finishDiagnosticGroups(groups)
	if err != nil {
		return completionResult{}, err
	}

	result := completionResult{
		Status:      AnalysisComplete,
		Reliability: ReliabilityExact,
		Coverage:    coverage,
		Diagnostics: diagnostics,
	}
	if lowerBound || approximate {
		result.Status = AnalysisPartial
		result.Reliability = ReliabilityLowerBound
	}
	if approximate {
		result.Reliability = ReliabilityApproximate
	}
	if err := validateCompletion(result); err != nil {
		return completionResult{}, err
	}

	return result, nil
}

func validateCompletion(result completionResult) error {
	if !result.Status.Valid() {
		return fmt.Errorf("invalid analysis status %q", result.Status)
	}
	if !result.Reliability.Valid() {
		return fmt.Errorf("invalid analysis reliability %q", result.Reliability)
	}
	if (result.Status == AnalysisComplete) != (result.Reliability == ReliabilityExact) {
		return fmt.Errorf(
			"inconsistent analysis status %q and reliability %q",
			result.Status,
			result.Reliability,
		)
	}

	return result.Coverage.Validate()
}

func validateUnresolvedEvidence(evidence UnresolvedInstance) error {
	if !evidence.Classification.Valid() ||
		evidence.Classification == TargetInheritedScene ||
		evidence.DeclaringScene == "" ||
		evidence.Occurrences <= 0 {
		return fmt.Errorf("invalid unresolved scene evidence: %#v", evidence)
	}
	if evidence.Classification == TargetUnresolvedPath &&
		(!evidence.ResolutionReason.Valid() || evidence.ResolutionReason == project.ResolutionResolved) {
		return fmt.Errorf("invalid unresolved path reason %q", evidence.ResolutionReason)
	}

	return nil
}

func validParentFinding(finding SceneParentFinding) bool {
	if finding.DeclaringScene == "" || finding.Occurrences <= 0 {
		return false
	}
	switch finding.Finding.Kind {
	case ParentInvalid, ParentMissing, ParentAmbiguous:
		return true
	default:
		return false
	}
}

func unresolvedDiagnostic(evidence UnresolvedInstance) (diagnostic.Diagnostic, diagnosticGroupKey) {
	code := diagnostic.CodeUnresolvedSceneInstance
	switch evidence.Classification {
	case TargetImportedScene:
		code = diagnostic.CodeImportedScene
	case TargetPlaceholder:
		code = diagnostic.CodeInstancePlaceholder
	case TargetUnavailableScene:
		code = diagnostic.CodeUnavailableResource
	case TargetUnresolvedPath:
		switch evidence.ResolutionReason {
		case project.ResolutionUIDOnly, project.ResolutionUserData:
			code = diagnostic.CodeUnsupportedResourcePath
		case project.ResolutionEmpty,
			project.ResolutionMissing,
			project.ResolutionOutsideProject,
			project.ResolutionFilesystem,
			project.ResolutionInvalidDeclaringScene:
			code = diagnostic.CodeUnavailableResource
		}
	}
	resource := unresolvedTarget(evidence)
	message := unresolvedMessage(code, evidence.Classification, evidence.ResolutionReason)
	item := diagnostic.Diagnostic{
		Code:     code,
		Severity: diagnostic.SeverityWarning,
		Message:  message,
		File:     evidence.DeclaringDisplay,
		Line:     evidence.Position.Line,
		Column:   evidence.Position.Column,
		Resource: resource,
	}
	if item.File == "" {
		item.File = evidence.DeclaringScene
	}

	return item, diagnosticKey(item, resource, evidence.ResolutionReason, evidence.Classification)
}

func inheritedDiagnostic(evidence InheritedTarget) (diagnostic.Diagnostic, diagnosticGroupKey) {
	resource := firstNonEmpty(
		evidence.BaseDisplay,
		evidence.BaseCanonical,
		evidence.BaseRawTarget,
		evidence.TargetDisplay,
		evidence.TargetCanonical,
		evidence.TargetOriginal,
	)
	item := diagnostic.Diagnostic{
		Code:     diagnostic.CodeInheritedScene,
		Severity: diagnostic.SeverityWarning,
		Message:  "inherited scene detected; override merge is unsupported",
		File:     evidence.DeclaringDisplay,
		Line:     evidence.MountPosition.Line,
		Column:   evidence.MountPosition.Column,
		Resource: resource,
	}
	if item.File == "" {
		item.File = evidence.DeclaringScene
	}

	classification := evidence.BaseClassification
	if classification == "" {
		classification = evidence.Classification
	}

	return item, diagnosticKey(item, resource, evidence.BaseResolutionReason, classification)
}

func parentDiagnostic(finding SceneParentFinding) (diagnostic.Diagnostic, diagnosticGroupKey) {
	resource := firstNonEmpty(finding.Finding.NodePath, finding.Finding.NodeName)
	item := diagnostic.Diagnostic{
		Code:     diagnostic.CodeUnsupportedParent,
		Severity: diagnostic.SeverityWarning,
		Message: fmt.Sprintf(
			"unsupported parent semantics: %s parent %q",
			finding.Finding.Kind,
			finding.Finding.Parent,
		),
		File:     finding.DeclaringDisplay,
		Line:     finding.Finding.Position.Line,
		Column:   finding.Finding.Position.Column,
		Resource: resource,
	}
	if item.File == "" {
		item.File = finding.DeclaringScene
	}

	return item, diagnosticKey(
		item,
		resource,
		project.ResolutionReason(finding.Finding.Kind),
		TargetClassification("unsupported_parent"),
	)
}

func resourceDiagnostic(
	identity ResourceIdentity,
	file string,
) (diagnostic.Diagnostic, diagnosticGroupKey) {
	code := diagnostic.CodeUnavailableResource
	message := fmt.Sprintf("external resource path is unavailable: %s", identity.ResolutionReason)
	if identity.ResolutionReason == project.ResolutionUIDOnly ||
		identity.ResolutionReason == project.ResolutionUserData {
		code = diagnostic.CodeUnsupportedResourcePath
		message = fmt.Sprintf("external resource path is unsupported: %s", identity.ResolutionReason)
	}
	item := diagnostic.Diagnostic{
		Code:     code,
		Severity: diagnostic.SeverityWarning,
		Message:  message,
		File:     file,
		Resource: firstNonEmpty(identity.RawPath, identity.ResourceID),
	}
	resource := identity.ResourceID + "\x00" + identity.RawPath

	return item, diagnosticKey(item, resource, identity.ResolutionReason, TargetUnresolvedPath)
}

func diagnosticKey(
	item diagnostic.Diagnostic,
	resource string,
	reason project.ResolutionReason,
	classification TargetClassification,
) diagnosticGroupKey {
	return diagnosticGroupKey{
		code:           item.Code,
		file:           item.File,
		resource:       resource,
		classification: classification,
		reason:         reason,
		message:        item.Message,
	}
}

func addDiagnosticGroup(
	groups map[diagnosticGroupKey]diagnosticGroup,
	key diagnosticGroupKey,
	item diagnostic.Diagnostic,
	occurrences int64,
) error {
	if occurrences <= 0 {
		return fmt.Errorf("diagnostic occurrences must be positive, got %d", occurrences)
	}
	group := groups[key]
	nextOccurrences, err := checkedAdd(group.item.Occurrences, occurrences)
	if err != nil {
		return err
	}
	if group.item.Code == "" {
		group.item = item
	} else if earlierPosition(item, group.item) {
		group.item.Line = item.Line
		group.item.Column = item.Column
	}
	group.item.Occurrences = nextOccurrences
	groups[key] = group

	return nil
}

func earlierPosition(left, right diagnostic.Diagnostic) bool {
	if left.Line == 0 {
		return false
	}
	if right.Line == 0 || left.Line != right.Line {
		return right.Line == 0 || left.Line < right.Line
	}

	return left.Column < right.Column
}

func finishDiagnosticGroups(groups map[diagnosticGroupKey]diagnosticGroup) ([]diagnostic.Diagnostic, error) {
	items := make([]diagnostic.Diagnostic, 0, len(groups))
	for _, group := range groups {
		if err := group.item.Validate(); err != nil {
			return nil, err
		}
		items = append(items, group.item)
	}
	sort.Slice(items, func(left, right int) bool {
		first := items[left]
		second := items[right]
		if first.Severity != second.Severity {
			return first.Severity < second.Severity
		}
		if first.Code != second.Code {
			return first.Code < second.Code
		}
		if first.File != second.File {
			return first.File < second.File
		}
		if first.Line != second.Line {
			return first.Line < second.Line
		}
		if first.Column != second.Column {
			return first.Column < second.Column
		}
		if first.Resource != second.Resource {
			return first.Resource < second.Resource
		}

		return first.Message < second.Message
	})

	return items, nil
}

func graphDisplayPaths(graph DependencyGraph) map[string]string {
	displays := make(map[string]string, len(graph.Nodes))
	for _, node := range graph.Nodes {
		displays[node.Canonical] = node.Display
	}
	if _, exists := displays[graph.RootCanonical]; !exists {
		displays[graph.RootCanonical] = graph.RootDisplay
	}

	return displays
}

func displayPath(displays map[string]string, canonical string) string {
	if display := displays[canonical]; display != "" {
		return display
	}

	return canonical
}

func unresolvedTarget(evidence UnresolvedInstance) string {
	return firstNonEmpty(
		evidence.TargetDisplay,
		evidence.RawTarget,
		evidence.TargetCanonical,
		evidence.TargetOriginal,
		evidence.ResourceID,
		evidence.MountPath,
	)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func unresolvedMessage(
	code diagnostic.Code,
	classification TargetClassification,
	reason project.ResolutionReason,
) string {
	switch code {
	case diagnostic.CodeImportedScene:
		return "imported or binary scene cannot be expanded"
	case diagnostic.CodeInstancePlaceholder:
		return "instance placeholder cannot be expanded"
	case diagnostic.CodeUnavailableResource:
		return fmt.Sprintf("resource path is unavailable: %s", reason)
	case diagnostic.CodeUnsupportedResourcePath:
		return fmt.Sprintf("resource path is unsupported: %s", reason)
	default:
		return fmt.Sprintf("scene instance cannot be expanded: %s", classification)
	}
}
