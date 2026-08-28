package budget

import "github.com/stfulldev/deadweight.gdt/internal/metrics"

// Result is one inclusive upper-bound comparison.
type Result struct {
	Metric metrics.Name
	Actual int64
	Limit  int64
	Delta  int64
	Passed bool
}

// Check evaluates configured limits in canonical metric order.
func Check(values metrics.Values, limits Limits) []Result {
	results := make([]Result, 0, limits.Count())

	for _, name := range metrics.OrderedNames() {
		limit, configured := limits.Get(name)
		if !configured {
			continue
		}

		actual, _ := values.Get(name)
		results = append(results, Result{
			Metric: name,
			Actual: actual,
			Limit:  limit,
			Delta:  actual - limit,
			Passed: actual <= limit,
		})
	}

	return results
}
