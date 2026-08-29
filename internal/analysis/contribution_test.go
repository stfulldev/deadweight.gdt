package analysis

import (
	"errors"
	"math"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestSceneContributionValidationRejectsInvalidDomainValues(t *testing.T) {
	t.Parallel()

	valid := SceneContribution{
		Kind:           ContributionRoot,
		SceneCanonical: "/project/root.tscn",
		Occurrences:    1,
		Values:         ContributionValues{Nodes: 1},
		DepthCandidate: OptionalDepth{Value: 1, Known: true},
		Reliability:    ReliabilityExact,
	}
	tests := []struct {
		name   string
		mutate func(*SceneContribution)
	}{
		{name: "kind", mutate: func(item *SceneContribution) { item.Kind = "future" }},
		{name: "reliability", mutate: func(item *SceneContribution) { item.Reliability = "future" }},
		{name: "occurrences", mutate: func(item *SceneContribution) { item.Occurrences = 0 }},
		{name: "negative additive", mutate: func(item *SceneContribution) { item.Values.Nodes = -1 }},
		{name: "invalid depth", mutate: func(item *SceneContribution) { item.DepthCandidate.Value = 0 }},
		{name: "missing root identity", mutate: func(item *SceneContribution) { item.SceneCanonical = "" }},
		{name: "root has declaring scene", mutate: func(item *SceneContribution) { item.DeclaringScene = "/project/parent.tscn" }},
		{name: "unresolved target absent", mutate: func(item *SceneContribution) {
			item.Kind = ContributionUnresolved
			item.SceneCanonical = ""
			item.DeclaringScene = "/project/root.tscn"
			item.Classification = TargetImportedScene
		}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			item := valid
			test.mutate(&item)
			if err := item.Validate(); err == nil {
				t.Fatalf("Validate() accepted %#v", item)
			}
		})
	}
}

func TestContributionCompactionIsCheckedConservativeAndOwned(t *testing.T) {
	t.Parallel()

	base := SceneContribution{
		Kind:             ContributionScene,
		SceneCanonical:   "/project/child.tscn",
		SceneDisplay:     "res://child.tscn",
		DeclaringScene:   "/project/root.tscn",
		DeclaringDisplay: "res://root.tscn",
		MountPath:        "Child",
		Occurrences:      2,
		Values:           ContributionValues{Nodes: 4, SceneInstances: 2},
		DepthCandidate:   OptionalDepth{Value: 3, Known: true},
		Reliability:      ReliabilityExact,
	}
	second := base
	second.Occurrences = 3
	second.Values = ContributionValues{Nodes: 6, SceneInstances: 3}
	second.DepthCandidate.Value = 4
	second.Reliability = ReliabilityApproximate

	got, err := compactContributions([]SceneContribution{second, base})
	if err != nil {
		t.Fatalf("compactContributions() error = %v", err)
	}
	if len(got) != 1 || got[0].Occurrences != 5 ||
		got[0].Values != (ContributionValues{Nodes: 10, SceneInstances: 5}) ||
		got[0].DepthCandidate != (OptionalDepth{Value: 4, Known: true}) ||
		got[0].Reliability != ReliabilityApproximate {
		t.Fatalf("compacted contribution = %#v", got)
	}
	got[0].Values.Nodes = -1
	if base.Values.Nodes != 4 || second.Values.Nodes != 6 {
		t.Fatal("compaction result aliases caller values")
	}

	overflow := base
	overflow.Occurrences = math.MaxInt64
	if _, err := compactContributions([]SceneContribution{overflow, base}); !isArithmeticOverflow(err) {
		t.Fatalf("occurrence overflow = %v", err)
	}
	overflow = base
	overflow.Values.Nodes = math.MaxInt64
	if _, err := scaleContribution(overflow, 2); !isArithmeticOverflow(err) {
		t.Fatalf("value overflow = %v", err)
	}
}

func TestValidateContributionEvidenceChecksAllAggregationModes(t *testing.T) {
	t.Parallel()

	result := RecursiveResult{
		Summary: ExpandedSummary{
			Metrics: metrics.Values{
				Nodes: 3, TreeDepth: 4, SceneInstances: 1,
				MeshInstances: 1, ExternalResources: 1, SceneDependencies: 1,
			},
			Contributions: []SceneContribution{
				{
					Kind: ContributionRoot, SceneCanonical: "/project/root.tscn", Occurrences: 1,
					Values:         ContributionValues{Nodes: 2},
					DepthCandidate: OptionalDepth{Value: 2, Known: true}, Reliability: ReliabilityExact,
				},
				{
					Kind: ContributionScene, SceneCanonical: "/project/child.tscn",
					DeclaringScene: "/project/root.tscn", MountPath: "Child", Occurrences: 1,
					Values:         ContributionValues{Nodes: 1, SceneInstances: 1, MeshInstances: 1},
					DepthCandidate: OptionalDepth{Value: 4, Known: true}, Reliability: ReliabilityExact,
				},
			},
		},
		UniqueEvidence: []UniqueEvidence{
			{
				Metric: metrics.ExternalResources, Canonical: "/project/mesh.res", Display: "res://mesh.res",
				Referrers: []UniqueReferrer{{SceneCanonical: "/project/root.tscn", Occurrences: 1}},
			},
			{
				Metric: metrics.SceneDependencies, Canonical: "/project/child.tscn", Display: "res://child.tscn",
				Referrers: []UniqueReferrer{{SceneCanonical: "/project/root.tscn", EdgeKind: EdgeInstance, Occurrences: 1}},
			},
		},
	}
	if err := ValidateContributionEvidence(result); err != nil {
		t.Fatalf("ValidateContributionEvidence() error = %v", err)
	}

	badAdditive := result
	badAdditive.Summary.Contributions = cloneContributions(result.Summary.Contributions)
	badAdditive.Summary.Contributions[1].Values.Nodes++
	if err := ValidateContributionEvidence(badAdditive); err == nil {
		t.Fatal("additive mismatch was accepted")
	}
	badDepth := result
	badDepth.Summary.Contributions = cloneContributions(result.Summary.Contributions)
	badDepth.Summary.Contributions[1].DepthCandidate.Value--
	if err := ValidateContributionEvidence(badDepth); err == nil {
		t.Fatal("depth mismatch was accepted")
	}
	badUnique := result
	badUnique.UniqueEvidence = cloneUniqueEvidence(result.UniqueEvidence[:1])
	if err := ValidateContributionEvidence(badUnique); err == nil {
		t.Fatal("unique cardinality mismatch was accepted")
	}
}

func requireContribution(
	t *testing.T,
	items []SceneContribution,
	kind ContributionKind,
	sceneCanonical string,
	declaringScene string,
) SceneContribution {
	t.Helper()
	for _, item := range items {
		if item.Kind == kind && item.SceneCanonical == sceneCanonical && item.DeclaringScene == declaringScene {
			return item
		}
	}
	t.Fatalf("missing %q contribution for %q declared by %q: %#v", kind, sceneCanonical, declaringScene, items)
	return SceneContribution{}
}

func requireUniqueEvidence(
	t *testing.T,
	items []UniqueEvidence,
	metric metrics.Name,
	canonical string,
) UniqueEvidence {
	t.Helper()
	for _, item := range items {
		if item.Metric == metric && item.Canonical == canonical {
			return item
		}
	}
	t.Fatalf("missing unique evidence for %q %q: %#v", metric, canonical, items)
	return UniqueEvidence{}
}

func isArithmeticOverflow(err error) bool {
	var overflow *OverflowError
	return errors.As(err, &overflow) && overflow.DiagnosticCode() == diagnostic.CodeArithmeticOverflow
}
