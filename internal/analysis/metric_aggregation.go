package analysis

import "github.com/stfulldev/deadweight.gdt/internal/metrics"

func finalizeMetrics(summary ExpandedSummary, graph DependencyGraph) (metrics.Values, error) {
	return finalizeMetricValues(
		summary.Metrics,
		uint64(len(summary.ExternalResources)),
		graph.SceneDependencies,
	)
}

func finalizeMetricValues(
	values metrics.Values,
	externalResourceCount uint64,
	sceneDependencies int64,
) (metrics.Values, error) {
	externalResources, err := checkedCardinality(externalResourceCount)
	if err != nil {
		return metrics.Values{}, err
	}

	values.ExternalResources = externalResources
	values.SceneDependencies = sceneDependencies
	if err := values.Validate(); err != nil {
		return metrics.Values{}, err
	}

	return values, nil
}
