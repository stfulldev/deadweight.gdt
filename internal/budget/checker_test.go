package budget_test

import (
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestCheckUsesInclusiveLimitsAndCanonicalOrder(t *testing.T) {
	t.Parallel()

	nodes := int64(10)
	lights := int64(1)
	results := budget.Check(metrics.Values{
		Nodes:  10,
		Lights: 2,
	}, budget.Limits{
		Lights: &lights,
		Nodes:  &nodes,
	})

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}

	if results[0].Metric != metrics.Nodes || !results[0].Passed || results[0].Delta != 0 {
		t.Errorf("nodes result = %#v, want boundary pass", results[0])
	}

	if results[1].Metric != metrics.Lights || results[1].Passed || results[1].Delta != 1 {
		t.Errorf("lights result = %#v, want failure with delta 1", results[1])
	}
}

func TestZeroLimitIsConfigured(t *testing.T) {
	t.Parallel()

	zero := int64(0)
	limits := budget.Limits{ShadowLights: &zero}

	if limits.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", limits.Count())
	}

	value, ok := limits.Get(metrics.ShadowLights)
	if !ok || value != 0 {
		t.Fatalf("Get(shadow_lights) = (%d, %t), want (0, true)", value, ok)
	}
}
