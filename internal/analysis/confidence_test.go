package analysis

import (
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func TestMetricConfidenceValidationMergeAndOwnership(t *testing.T) {
	t.Parallel()

	exact := ExactMetricConfidence()
	if err := exact.Validate(); err != nil || exact.Reliability() != ReliabilityExact {
		t.Fatalf("exact confidence = %#v, %v", exact, err)
	}

	mixed, err := exact.With(
		metrics.ExternalResources,
		ReliabilityLowerBound,
		ConfidenceUnavailableResource,
	)
	if err != nil {
		t.Fatalf("With(lower_bound) error = %v", err)
	}
	mixed, err = mixed.With(
		metrics.ExternalResources,
		ReliabilityApproximate,
		ConfidenceInheritedScene,
		ConfidenceUnavailableResource,
	)
	if err != nil {
		t.Fatalf("With(approximate) error = %v", err)
	}
	resource, _ := mixed.Get(metrics.ExternalResources)
	if resource.Reliability != ReliabilityApproximate ||
		!reflect.DeepEqual(resource.Reasons, []ConfidenceReason{
			ConfidenceUnavailableResource,
			ConfidenceInheritedScene,
		}) || mixed.Reliability() != ReliabilityApproximate {
		t.Fatalf("mixed confidence = %#v / %#v", mixed, resource)
	}
	resource.Reasons[0] = ConfidenceUnsupportedParent
	again, _ := mixed.Get(metrics.ExternalResources)
	if again.Reasons[0] != ConfidenceUnavailableResource {
		t.Fatal("Get() returned aliased reasons")
	}
	if value, _ := exact.Get(metrics.ExternalResources); value.Reliability != ReliabilityExact {
		t.Fatal("With() mutated its source")
	}
}

func TestConfidenceReasonMappingCoversStaticEvidenceClasses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		classification TargetClassification
		resolution     project.ResolutionReason
		want           ConfidenceReason
	}{
		{TargetMissingExternalResource, project.ResolutionEmpty, ConfidenceUnresolvedSceneInstance},
		{TargetUnresolvedPath, project.ResolutionMissing, ConfidenceUnavailableScene},
		{TargetUnresolvedPath, project.ResolutionUIDOnly, ConfidenceUnsupportedResourcePath},
		{TargetImportedScene, project.ResolutionResolved, ConfidenceImportedScene},
		{TargetUnsupportedScene, project.ResolutionResolved, ConfidenceUnsupportedScene},
		{TargetSubResource, project.ResolutionEmpty, ConfidenceSubresourceScene},
		{TargetPlaceholder, project.ResolutionEmpty, ConfidencePlaceholderInstance},
		{TargetUnavailableScene, project.ResolutionFilesystem, ConfidenceUnavailableScene},
	}
	for _, test := range tests {
		if got := confidenceReasonForSceneTarget(test.classification, test.resolution); got != test.want {
			t.Errorf("reason for %q/%q = %q, want %q", test.classification, test.resolution, got, test.want)
		}
	}
	if got := confidenceReasonForResource(project.ResolutionMissing); got != ConfidenceUnavailableResource {
		t.Errorf("ordinary missing resource reason = %q", got)
	}
	if got := confidenceReasonForResource(project.ResolutionUserData); got != ConfidenceUnsupportedResourcePath {
		t.Errorf("ordinary unsupported resource reason = %q", got)
	}
}

func TestFinalizeCompletenessScopesEvidenceToAffectedMetrics(t *testing.T) {
	t.Parallel()

	ordinary := ExpandedSummary{ExternalResources: []ResourceIdentity{{
		DeclaringScene:   "/project/root.tscn",
		ResourceID:       "texture",
		RawPath:          "missing.png",
		ResolutionReason: project.ResolutionMissing,
	}}}
	resourceResult, err := finalizeCompleteness(ordinary, completionTestGraph(), 1)
	if err != nil {
		t.Fatalf("ordinary resource finalization error = %v", err)
	}
	assertOnlyMetricConfidence(
		t,
		resourceResult.MetricConfidence,
		metrics.ExternalResources,
		ReliabilityLowerBound,
		ConfidenceUnavailableResource,
	)
	if resourceResult.Status != AnalysisPartial || resourceResult.Reliability != ReliabilityLowerBound {
		t.Fatalf("ordinary resource summary = %q/%q", resourceResult.Status, resourceResult.Reliability)
	}

	parentResult, err := finalizeCompleteness(
		ExpandedSummary{DepthPartial: true},
		completionTestGraph(),
		1,
	)
	if err != nil {
		t.Fatalf("parent finalization error = %v", err)
	}
	assertOnlyMetricConfidence(
		t,
		parentResult.MetricConfidence,
		metrics.TreeDepth,
		ReliabilityLowerBound,
		ConfidenceUnsupportedParent,
	)
}

func TestUnresolvedSceneAndInheritanceQualifyEveryFrozenMetric(t *testing.T) {
	t.Parallel()

	unresolvedSummary := ExpandedSummary{
		Metrics:  metrics.Values{SceneInstances: 1},
		Coverage: SceneInstanceCoverage{Unresolved: 1},
		Unresolved: []UnresolvedInstance{{
			Classification:   TargetImportedScene,
			ResolutionReason: project.ResolutionResolved,
			DeclaringScene:   "/project/root.tscn",
			RawTarget:        "model.glb",
			Occurrences:      1,
		}},
	}
	unresolved, err := finalizeCompleteness(unresolvedSummary, completionTestGraph(), 1)
	if err != nil {
		t.Fatalf("unresolved finalization error = %v", err)
	}
	for _, entry := range unresolved.MetricConfidence.Entries() {
		if entry.Confidence.Reliability != ReliabilityLowerBound ||
			!reflect.DeepEqual(entry.Confidence.Reasons, []ConfidenceReason{ConfidenceImportedScene}) {
			t.Errorf("unresolved %q confidence = %#v", entry.Metric, entry.Confidence)
		}
	}

	inheritedSummary := ExpandedSummary{InheritedTargets: []InheritedTarget{{
		Classification: TargetInheritedScene,
		DeclaringScene: "/project/root.tscn",
		Occurrences:    1,
	}}}
	inherited, err := finalizeCompleteness(inheritedSummary, completionTestGraph(), 1)
	if err != nil {
		t.Fatalf("inherited finalization error = %v", err)
	}
	for _, entry := range inherited.MetricConfidence.Entries() {
		if entry.Confidence.Reliability != ReliabilityApproximate ||
			!reflect.DeepEqual(entry.Confidence.Reasons, []ConfidenceReason{ConfidenceInheritedScene}) {
			t.Errorf("inherited %q confidence = %#v", entry.Metric, entry.Confidence)
		}
	}
}

func TestContributionConfidenceScopesParentAndOwnsReasons(t *testing.T) {
	t.Parallel()

	summary := ExpandedSummary{
		Contributions: []SceneContribution{{
			Kind:             ContributionRoot,
			SceneCanonical:   "/project/root.tscn",
			Occurrences:      1,
			Reliability:      ReliabilityExact,
			MetricConfidence: ExactMetricConfidence(),
		}},
		ParentFindings: []SceneParentFinding{{DeclaringScene: "/project/root.tscn"}},
	}
	if err := qualifyContributionConfidence(&summary); err != nil {
		t.Fatalf("qualifyContributionConfidence() error = %v", err)
	}
	item := summary.Contributions[0]
	if item.Reliability != ReliabilityLowerBound {
		t.Fatalf("row reliability = %q", item.Reliability)
	}
	assertOnlyMetricConfidence(
		t,
		item.MetricConfidence,
		metrics.TreeDepth,
		ReliabilityLowerBound,
		ConfidenceUnsupportedParent,
	)
	cloned := cloneContributions(summary.Contributions)
	cloned[0].MetricConfidence.TreeDepth.Reasons[0] = ConfidenceUnavailableResource
	original, _ := summary.Contributions[0].MetricConfidence.Get(metrics.TreeDepth)
	if original.Reasons[0] != ConfidenceUnsupportedParent {
		t.Fatal("cloned contribution aliases confidence reasons")
	}
}

func assertOnlyMetricConfidence(
	t *testing.T,
	confidence MetricConfidence,
	selected metrics.Name,
	reliability Reliability,
	reason ConfidenceReason,
) {
	t.Helper()
	for _, entry := range confidence.Entries() {
		if entry.Metric == selected {
			if entry.Confidence.Reliability != reliability ||
				!reflect.DeepEqual(entry.Confidence.Reasons, []ConfidenceReason{reason}) {
				t.Errorf("selected %q confidence = %#v", entry.Metric, entry.Confidence)
			}
			continue
		}
		if entry.Confidence.Reliability != ReliabilityExact || len(entry.Confidence.Reasons) != 0 {
			t.Errorf("unrelated %q confidence = %#v", entry.Metric, entry.Confidence)
		}
	}
}

func TestConfidenceRejectsInvalidOrNonCanonicalMetadata(t *testing.T) {
	t.Parallel()

	tests := []Confidence{
		{},
		{Reliability: ReliabilityExact, Reasons: []ConfidenceReason{ConfidenceUnavailableResource}},
		{Reliability: ReliabilityLowerBound},
		{Reliability: ReliabilityLowerBound, Reasons: []ConfidenceReason{"future"}},
		{
			Reliability: ReliabilityLowerBound,
			Reasons: []ConfidenceReason{
				ConfidenceInheritedScene,
				ConfidenceUnavailableResource,
			},
		},
		{
			Reliability: ReliabilityLowerBound,
			Reasons: []ConfidenceReason{
				ConfidenceUnavailableResource,
				ConfidenceUnavailableResource,
			},
		},
	}
	for _, confidence := range tests {
		if err := confidence.Validate(); err == nil {
			t.Errorf("Validate() accepted %#v", confidence)
		}
	}

	if _, err := ExactMetricConfidence().With("future", ReliabilityLowerBound, ConfidenceUnavailableResource); err == nil {
		t.Fatal("With() accepted an unknown metric")
	}
}

func testMetricConfidence(reliability Reliability) MetricConfidence {
	reasons := []ConfidenceReason(nil)
	switch reliability {
	case ReliabilityLowerBound:
		reasons = []ConfidenceReason{ConfidenceUnresolvedSceneInstance}
	case ReliabilityApproximate:
		reasons = []ConfidenceReason{ConfidenceInheritedScene}
	}
	confidence, err := UniformMetricConfidence(reliability, reasons...)
	if err != nil {
		panic(err)
	}

	return confidence
}
