package report

import (
	"fmt"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

type contributionV1 struct {
	Kind           analysis.ContributionKind     `json:"kind"`
	Scene          string                        `json:"scene"`
	DeclaringScene string                        `json:"declaring_scene,omitempty"`
	MountPath      string                        `json:"mount_path,omitempty"`
	MountDepth     *int64                        `json:"mount_depth,omitempty"`
	ResourceID     string                        `json:"resource_id,omitempty"`
	RawTarget      string                        `json:"raw_target,omitempty"`
	Classification analysis.TargetClassification `json:"classification,omitempty"`
	Occurrences    int64                         `json:"occurrences"`
	Reliability    analysis.Reliability          `json:"reliability"`
	Metrics        []contributionMetricV1        `json:"metrics"`
}

type contributionMetricV1 struct {
	ID          metrics.Name `json:"id"`
	Aggregation string       `json:"aggregation"`
	Value       *int64       `json:"value,omitempty"`
	Available   *bool        `json:"available,omitempty"`
}

type uniqueEvidenceV1 struct {
	Metric    metrics.Name       `json:"metric"`
	Identity  string             `json:"identity"`
	Referrers []uniqueReferrerV1 `json:"referrers"`
}

type uniqueReferrerV1 struct {
	Scene       string            `json:"scene"`
	ResourceID  string            `json:"resource_id,omitempty"`
	RawTarget   string            `json:"raw_target,omitempty"`
	EdgeKind    analysis.EdgeKind `json:"edge_kind,omitempty"`
	Occurrences int64             `json:"occurrences"`
}

type topContributorsV1 struct {
	Metric metrics.Name     `json:"metric"`
	Limit  int64            `json:"limit"`
	Rows   []contributionV1 `json:"rows"`
}

func contributionDocumentsV1(result application.InspectResult) ([]contributionV1, error) {
	documents := make([]contributionV1, 0, len(result.Analysis.Summary.Contributions))
	for _, item := range result.Analysis.Summary.Contributions {
		document, err := contributionDocumentV1(result.Project.Directory, item)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	sort.Slice(documents, func(left, right int) bool {
		return contributionDocumentLess(documents[left], documents[right])
	})

	return documents, nil
}

func contributionDocumentV1(projectRoot string, item analysis.SceneContribution) (contributionV1, error) {
	if err := item.Validate(); err != nil {
		return contributionV1{}, err
	}
	document := contributionV1{
		Kind:           item.Kind,
		Scene:          portableContributionIdentity(projectRoot, item),
		DeclaringScene: portableContributionDeclaring(projectRoot, item),
		MountPath:      portableMountPath(item.MountPath),
		ResourceID:     item.ResourceID,
		RawTarget:      portableRawTarget(projectRoot, item.RawTarget),
		Classification: item.Classification,
		Occurrences:    item.Occurrences,
		Reliability:    item.Reliability,
		Metrics:        contributionMetricsV1(item),
	}
	if item.MountDepth.Known {
		depth := item.MountDepth.Value
		document.MountDepth = &depth
	}

	return document, nil
}

func contributionMetricsV1(item analysis.SceneContribution) []contributionMetricV1 {
	result := make([]contributionMetricV1, 0, len(metrics.OrderedNames()))
	for _, name := range metrics.OrderedNames() {
		entry := contributionMetricV1{ID: name}
		switch name {
		case metrics.TreeDepth:
			entry.Aggregation = "maximum"
			available := item.DepthCandidate.Known
			entry.Available = &available
			if available {
				value := item.DepthCandidate.Value
				entry.Value = &value
			}
		case metrics.ExternalResources, metrics.SceneDependencies:
			entry.Aggregation = "unique_union"
		default:
			entry.Aggregation = "additive"
			value, _ := item.Values.Get(name)
			entry.Value = &value
		}
		result = append(result, entry)
	}

	return result
}

func contributionDocumentLess(left, right contributionV1) bool {
	if left.Scene != right.Scene {
		return left.Scene < right.Scene
	}
	if left.DeclaringScene != right.DeclaringScene {
		return left.DeclaringScene < right.DeclaringScene
	}
	if left.MountPath != right.MountPath {
		return left.MountPath < right.MountPath
	}
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	if left.ResourceID != right.ResourceID {
		return left.ResourceID < right.ResourceID
	}
	if left.RawTarget != right.RawTarget {
		return left.RawTarget < right.RawTarget
	}
	if left.Classification != right.Classification {
		return left.Classification < right.Classification
	}
	if (left.MountDepth == nil) != (right.MountDepth == nil) {
		return left.MountDepth == nil
	}
	if left.MountDepth != nil && *left.MountDepth != *right.MountDepth {
		return *left.MountDepth < *right.MountDepth
	}
	if left.Occurrences != right.Occurrences {
		return left.Occurrences < right.Occurrences
	}

	return left.Reliability < right.Reliability
}

func uniqueEvidenceDocumentsV1(result application.InspectResult) ([]uniqueEvidenceV1, error) {
	documents := make([]uniqueEvidenceV1, 0, len(result.Analysis.UniqueEvidence))
	for _, item := range result.Analysis.UniqueEvidence {
		if err := item.Validate(); err != nil {
			return nil, err
		}
		identity, err := portableUniqueIdentity(result.Project.Directory, item)
		if err != nil {
			return nil, err
		}
		document := uniqueEvidenceV1{Metric: item.Metric, Identity: identity}
		for _, referrer := range item.Referrers {
			scene, ok := firstPortablePath(
				result.Project.Directory,
				referrer.SceneDisplay,
				referrer.SceneCanonical,
			)
			if !ok {
				return nil, fmt.Errorf("unique referrer has no portable scene identity")
			}
			document.Referrers = append(document.Referrers, uniqueReferrerV1{
				Scene:       scene,
				ResourceID:  referrer.ResourceID,
				RawTarget:   portableRawTarget(result.Project.Directory, referrer.RawTarget),
				EdgeKind:    referrer.EdgeKind,
				Occurrences: referrer.Occurrences,
			})
		}
		sort.Slice(document.Referrers, func(left, right int) bool {
			first := document.Referrers[left]
			second := document.Referrers[right]
			if first.Scene != second.Scene {
				return first.Scene < second.Scene
			}
			if first.ResourceID != second.ResourceID {
				return first.ResourceID < second.ResourceID
			}
			if first.RawTarget != second.RawTarget {
				return first.RawTarget < second.RawTarget
			}
			if first.EdgeKind != second.EdgeKind {
				return first.EdgeKind < second.EdgeKind
			}

			return first.Occurrences < second.Occurrences
		})
		documents = append(documents, document)
	}
	sort.Slice(documents, func(left, right int) bool {
		if documents[left].Metric != documents[right].Metric {
			return uniqueMetricOrder(documents[left].Metric) < uniqueMetricOrder(documents[right].Metric)
		}

		return documents[left].Identity < documents[right].Identity
	})

	return documents, nil
}

func portableUniqueIdentity(projectRoot string, item analysis.UniqueEvidence) (string, error) {
	if portable, ok := firstPortablePath(projectRoot, item.Display, item.Canonical); ok {
		return portable, nil
	}
	raw := portableRawTarget(projectRoot, item.RawTarget)
	declaring, _ := firstPortablePath(projectRoot, item.DeclaringScene)
	if declaring != "" && item.ResourceID != "" {
		return declaring + "#" + item.ResourceID + ":" + raw, nil
	}
	if raw != "" {
		return raw, nil
	}

	return "", fmt.Errorf("unique evidence %q has no portable identity", item.Metric)
}

func firstPortablePath(projectRoot string, candidates ...string) (string, bool) {
	for _, candidate := range candidates {
		if portable, ok := portableOptionalPath(projectRoot, candidate); ok {
			return portable, true
		}
	}

	return "", false
}

func uniqueMetricOrder(name metrics.Name) int {
	if name == metrics.ExternalResources {
		return 0
	}
	if name == metrics.SceneDependencies {
		return 1
	}

	return 2
}

func topContributorsDocumentV1(
	result application.InspectResult,
	selection ContributionSelection,
) (*topContributorsV1, error) {
	if !selection.Present() {
		return nil, nil
	}
	projected, err := topContributions(result, selection)
	if err != nil {
		return nil, err
	}
	document := &topContributorsV1{
		Metric: selection.Metric,
		Limit:  selection.Limit,
		Rows:   make([]contributionV1, 0, len(projected)),
	}
	for _, item := range projected {
		row, err := contributionDocumentV1(
			result.Project.Directory,
			result.Analysis.Summary.Contributions[item.SourceIndex],
		)
		if err != nil {
			return nil, err
		}
		document.Rows = append(document.Rows, row)
	}

	return document, nil
}
