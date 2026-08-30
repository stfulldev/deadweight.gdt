package analysis

import (
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

func confidenceReasonForUnresolved(evidence UnresolvedInstance) ConfidenceReason {
	return confidenceReasonForSceneTarget(evidence.Classification, evidence.ResolutionReason)
}

func confidenceReasonForSceneTarget(
	classification TargetClassification,
	reason project.ResolutionReason,
) ConfidenceReason {
	switch classification {
	case TargetImportedScene:
		return ConfidenceImportedScene
	case TargetUnsupportedScene:
		return ConfidenceUnsupportedScene
	case TargetSubResource:
		return ConfidenceSubresourceScene
	case TargetPlaceholder:
		return ConfidencePlaceholderInstance
	case TargetUnavailableScene:
		return ConfidenceUnavailableScene
	case TargetUnresolvedPath:
		if reason == project.ResolutionUIDOnly ||
			reason == project.ResolutionUserData ||
			reason == project.ResolutionUnsupportedTarget {
			return ConfidenceUnsupportedResourcePath
		}
		return ConfidenceUnavailableScene
	default:
		return ConfidenceUnresolvedSceneInstance
	}
}

func confidenceReasonForResource(reason project.ResolutionReason) ConfidenceReason {
	if reason == project.ResolutionUIDOnly ||
		reason == project.ResolutionUserData ||
		reason == project.ResolutionUnsupportedTarget {
		return ConfidenceUnsupportedResourcePath
	}

	return ConfidenceUnavailableResource
}

func qualifyContributionConfidence(summary *ExpandedSummary) error {
	coveredResources := make(map[resourceDeclarationKey]struct{})
	for _, evidence := range summary.Unresolved {
		if evidence.ResourceID != "" {
			coveredResources[resourceDeclarationKey{
				declaringScene: evidence.DeclaringScene,
				resourceID:     evidence.ResourceID,
			}] = struct{}{}
		}
	}

	for index := range summary.Contributions {
		item := &summary.Contributions[index]
		if item.Kind == ContributionInherited {
			if err := item.MetricConfidence.merge(
				allMetricNames(),
				ReliabilityApproximate,
				ConfidenceInheritedScene,
			); err != nil {
				return err
			}
		}

		if item.Kind != ContributionRoot && !item.MountDepth.Known {
			if err := item.MetricConfidence.merge(
				[]metrics.Name{metrics.TreeDepth},
				ReliabilityLowerBound,
				ConfidenceUnsupportedParent,
			); err != nil {
				return err
			}
		}

		for _, finding := range summary.ParentFindings {
			if finding.DeclaringScene != item.SceneCanonical {
				continue
			}
			if err := item.MetricConfidence.merge(
				[]metrics.Name{metrics.TreeDepth},
				ReliabilityLowerBound,
				ConfidenceUnsupportedParent,
			); err != nil {
				return err
			}
		}

		for _, identity := range summary.ExternalResources {
			if identity.Resolved || identity.DeclaringScene != item.SceneCanonical {
				continue
			}
			declaration := resourceDeclarationKey{
				declaringScene: identity.DeclaringScene,
				resourceID:     identity.ResourceID,
			}
			if _, covered := coveredResources[declaration]; covered {
				continue
			}
			if err := item.MetricConfidence.merge(
				[]metrics.Name{metrics.ExternalResources},
				ReliabilityLowerBound,
				confidenceReasonForResource(identity.ResolutionReason),
			); err != nil {
				return err
			}
		}

		for _, evidence := range summary.InheritedTargets {
			if evidence.DeclaringScene != item.SceneCanonical {
				continue
			}
			if err := item.MetricConfidence.merge(
				allMetricNames(),
				ReliabilityApproximate,
				ConfidenceInheritedScene,
			); err != nil {
				return err
			}
		}

		item.Reliability = item.MetricConfidence.Reliability()
	}

	return nil
}
