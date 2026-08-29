package budget_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestStatusCatalogAndEvaluationClone(t *testing.T) {
	t.Parallel()

	for _, status := range []budget.Status{
		budget.StatusPassed,
		budget.StatusFailed,
		budget.StatusIncomplete,
	} {
		if !status.Valid() {
			t.Errorf("%q.Valid() = false", status)
		}
	}
	for _, status := range []budget.Status{"", "passed", "UNKNOWN"} {
		if status.Valid() {
			t.Errorf("%q.Valid() = true", status)
		}
	}

	original := budget.Evaluation{
		Status:      budget.StatusFailed,
		Reliability: analysis.ReliabilityExact,
		Exceeded:    1,
		Results: []budget.Result{{
			Metric: metrics.Nodes, Actual: 2, Limit: 1, Delta: 1,
		}},
	}
	cloned := original.Clone()
	cloned.Results[0].Actual = 99
	if original.Results[0].Actual != 2 {
		t.Fatalf("original result mutated to %#v", original.Results[0])
	}
}

func TestLimitsValidateUsesCanonicalOrder(t *testing.T) {
	t.Parallel()

	if err := (budget.Limits{}).Validate(); err != nil {
		t.Fatalf("empty Validate() error = %v", err)
	}
	valid := limitsFor(map[metrics.Name]int64{
		metrics.Nodes: 0, metrics.TreeDepth: 1, metrics.SceneInstances: 2,
		metrics.MeshInstances: 3, metrics.Lights: 4, metrics.ShadowLights: 5,
		metrics.ExternalResources: 6, metrics.SceneDependencies: 7,
	})
	if err := valid.Validate(); err != nil {
		t.Fatalf("full Validate() error = %v", err)
	}

	invalid := limitsFor(map[metrics.Name]int64{
		metrics.SceneDependencies: -3,
		metrics.Nodes:             -1,
		metrics.Lights:            -2,
	})
	err := invalid.Validate()
	var limitErr *budget.LimitError
	if !errors.As(err, &limitErr) {
		t.Fatalf("error = %T %v, want *budget.LimitError", err, err)
	}
	if limitErr.Metric != metrics.Nodes || limitErr.Limit != -1 {
		t.Fatalf("limit error = %#v, want canonical first nodes=-1", limitErr)
	}
}

func TestEvaluateAllMetricsInCanonicalOrder(t *testing.T) {
	t.Parallel()

	values := metricValues(map[metrics.Name]int64{
		metrics.Nodes: 0, metrics.TreeDepth: 2, metrics.SceneInstances: 3,
		metrics.MeshInstances: 4, metrics.Lights: 6, metrics.ShadowLights: 6,
		metrics.ExternalResources: 7, metrics.SceneDependencies: 9,
	})
	limits := limitsFor(map[metrics.Name]int64{
		metrics.SceneDependencies: 8,
		metrics.ExternalResources: 7,
		metrics.ShadowLights:      6,
		metrics.Lights:            5,
		metrics.MeshInstances:     4,
		metrics.SceneInstances:    3,
		metrics.TreeDepth:         2,
		metrics.Nodes:             0,
	})

	evaluation, err := budget.Evaluate(values, limits, analysis.ReliabilityExact, false)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if evaluation.Status != budget.StatusFailed || evaluation.Exceeded != 2 || evaluation.FailOnPartial {
		t.Fatalf("evaluation summary = %#v", evaluation)
	}
	if evaluation.Reliability != analysis.ReliabilityExact || len(evaluation.Results) != 8 {
		t.Fatalf("evaluation evidence = %#v", evaluation)
	}

	for index, name := range metrics.OrderedNames() {
		result := evaluation.Results[index]
		if result.Metric != name {
			t.Errorf("result[%d].Metric = %q, want %q", index, result.Metric, name)
		}
		wantDelta := result.Actual - result.Limit
		if result.Delta != wantDelta || result.Passed != (result.Actual <= result.Limit) {
			t.Errorf("result[%d] = %#v", index, result)
		}
	}
	if result := evaluation.Results[0]; !result.Passed || result.Limit != 0 || result.Delta != 0 {
		t.Fatalf("zero boundary result = %#v", result)
	}
	if result := evaluation.Results[4]; result.Passed || result.Delta != 1 {
		t.Fatalf("limit+1 result = %#v", result)
	}
}

func TestEvaluateSkipsAbsentLimitsAndAllowsEmpty(t *testing.T) {
	t.Parallel()

	zero := int64(0)
	evaluation, err := budget.Evaluate(
		metrics.Values{},
		budget.Limits{ShadowLights: &zero},
		analysis.ReliabilityExact,
		false,
	)
	if err != nil {
		t.Fatalf("Evaluate(zero) error = %v", err)
	}
	want := []budget.Result{{
		Metric: metrics.ShadowLights, Actual: 0, Limit: 0, Delta: 0, Passed: true,
	}}
	if !reflect.DeepEqual(evaluation.Results, want) || evaluation.Status != budget.StatusPassed {
		t.Fatalf("zero evaluation = %#v, want results %#v", evaluation, want)
	}

	empty, err := budget.Evaluate(
		metrics.Values{}, budget.Limits{}, analysis.ReliabilityExact, false,
	)
	if err != nil {
		t.Fatalf("Evaluate(empty) error = %v", err)
	}
	if empty.Status != budget.StatusPassed || empty.Exceeded != 0 || len(empty.Results) != 0 {
		t.Fatalf("empty evaluation = %#v", empty)
	}
}

func TestEvaluateVerdictAndReliabilityMatrix(t *testing.T) {
	t.Parallel()

	valuesPass := metrics.Values{Nodes: 10, Lights: 1}
	valuesFail := metrics.Values{Nodes: 11, Lights: 3}
	limits := limitsFor(map[metrics.Name]int64{metrics.Nodes: 10, metrics.Lights: 2})

	tests := []struct {
		name          string
		values        metrics.Values
		reliability   analysis.Reliability
		failOnPartial bool
		status        budget.Status
		exceeded      int
	}{
		{name: "exact pass", values: valuesPass, reliability: analysis.ReliabilityExact, status: budget.StatusPassed},
		{name: "exact fail", values: valuesFail, reliability: analysis.ReliabilityExact, status: budget.StatusFailed, exceeded: 2},
		{name: "exact ignores partial policy pass", values: valuesPass, reliability: analysis.ReliabilityExact, failOnPartial: true, status: budget.StatusPassed},
		{name: "exact ignores partial policy fail", values: valuesFail, reliability: analysis.ReliabilityExact, failOnPartial: true, status: budget.StatusFailed, exceeded: 2},
		{name: "lower bound allowed pass", values: valuesPass, reliability: analysis.ReliabilityLowerBound, status: budget.StatusPassed},
		{name: "lower bound allowed fail", values: valuesFail, reliability: analysis.ReliabilityLowerBound, status: budget.StatusFailed, exceeded: 2},
		{name: "lower bound rejected pass", values: valuesPass, reliability: analysis.ReliabilityLowerBound, failOnPartial: true, status: budget.StatusIncomplete},
		{name: "lower bound rejected fail", values: valuesFail, reliability: analysis.ReliabilityLowerBound, failOnPartial: true, status: budget.StatusIncomplete, exceeded: 2},
		{name: "approximate allowed fail", values: valuesFail, reliability: analysis.ReliabilityApproximate, status: budget.StatusFailed, exceeded: 2},
		{name: "approximate rejected fail", values: valuesFail, reliability: analysis.ReliabilityApproximate, failOnPartial: true, status: budget.StatusIncomplete, exceeded: 2},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			evaluation, err := budget.Evaluate(test.values, limits, test.reliability, test.failOnPartial)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}
			if evaluation.Status != test.status || evaluation.Exceeded != test.exceeded ||
				evaluation.Reliability != test.reliability || evaluation.FailOnPartial != test.failOnPartial {
				t.Fatalf("evaluation = %#v", evaluation)
			}
			if len(evaluation.Results) != 2 {
				t.Fatalf("len(results) = %d, want 2", len(evaluation.Results))
			}
			failed := 0
			for _, result := range evaluation.Results {
				if !result.Passed {
					failed++
				}
			}
			if failed != test.exceeded {
				t.Fatalf("retained failed results = %d, want %d", failed, test.exceeded)
			}
		})
	}
}

func TestResolveFailOnPartialMatrix(t *testing.T) {
	t.Parallel()

	for _, override := range []budget.PartialOverride{
		budget.PartialInherit,
		budget.PartialFail,
		budget.PartialAllow,
	} {
		if !override.Valid() {
			t.Errorf("%q.Valid() = false", override)
		}
	}
	if budget.PartialOverride("unknown").Valid() {
		t.Fatal("unknown override is valid")
	}

	tests := []struct {
		configured bool
		override   budget.PartialOverride
		want       bool
	}{
		{configured: false, override: budget.PartialInherit, want: false},
		{configured: true, override: budget.PartialInherit, want: true},
		{configured: false, override: budget.PartialFail, want: true},
		{configured: true, override: budget.PartialFail, want: true},
		{configured: false, override: budget.PartialAllow, want: false},
		{configured: true, override: budget.PartialAllow, want: false},
	}
	for _, test := range tests {
		got, err := budget.ResolveFailOnPartial(test.configured, test.override)
		if err != nil || got != test.want {
			t.Errorf("ResolveFailOnPartial(%v, %q) = %v, %v; want %v, nil", test.configured, test.override, got, err, test.want)
		}
	}

	got, err := budget.ResolveFailOnPartial(true, "unknown")
	if got {
		t.Fatal("invalid override returned true")
	}
	var overrideErr *budget.PartialOverrideError
	if !errors.As(err, &overrideErr) || overrideErr.Override != "unknown" {
		t.Fatalf("error = %T %#v, want PartialOverrideError", err, err)
	}
}

func TestEvaluateRejectsEveryNegativeMetric(t *testing.T) {
	t.Parallel()

	for _, name := range metrics.OrderedNames() {
		name := name
		t.Run(string(name), func(t *testing.T) {
			t.Parallel()

			values := metricValues(map[metrics.Name]int64{name: -1})
			got, err := budget.Evaluate(values, budget.Limits{}, analysis.ReliabilityExact, false)
			if !reflect.DeepEqual(got, budget.Evaluation{}) {
				t.Fatalf("evaluation = %#v, want zero", got)
			}
			var valueErr *metrics.ValueError
			if !errors.As(err, &valueErr) || valueErr.Name != name || valueErr.Value != -1 {
				t.Fatalf("error = %T %#v, want ValueError for %q", err, err, name)
			}
		})
	}
}

func TestEvaluateRejectsEveryNegativeLimit(t *testing.T) {
	t.Parallel()

	for _, name := range metrics.OrderedNames() {
		name := name
		t.Run(string(name), func(t *testing.T) {
			t.Parallel()

			limits := limitsFor(map[metrics.Name]int64{name: -1})
			got, err := budget.Evaluate(metrics.Values{}, limits, analysis.ReliabilityExact, false)
			if !reflect.DeepEqual(got, budget.Evaluation{}) {
				t.Fatalf("evaluation = %#v, want zero", got)
			}
			var limitErr *budget.LimitError
			if !errors.As(err, &limitErr) || limitErr.Metric != name || limitErr.Limit != -1 {
				t.Fatalf("error = %T %#v, want LimitError for %q", err, err, name)
			}
		})
	}
}

func TestEvaluateRejectsUnknownReliabilityAndOwnsResults(t *testing.T) {
	t.Parallel()

	limits := limitsFor(map[metrics.Name]int64{metrics.Nodes: 1})
	got, err := budget.Evaluate(metrics.Values{}, limits, "unknown", false)
	if !reflect.DeepEqual(got, budget.Evaluation{}) {
		t.Fatalf("invalid reliability evaluation = %#v, want zero", got)
	}
	var reliabilityErr *budget.ReliabilityError
	if !errors.As(err, &reliabilityErr) || reliabilityErr.Reliability != "unknown" {
		t.Fatalf("error = %T %#v, want ReliabilityError", err, err)
	}

	first, err := budget.Evaluate(metrics.Values{Nodes: 1}, limits, analysis.ReliabilityExact, false)
	if err != nil {
		t.Fatalf("first Evaluate() error = %v", err)
	}
	first.Results[0].Actual = 99
	second, err := budget.Evaluate(metrics.Values{Nodes: 1}, limits, analysis.ReliabilityExact, false)
	if err != nil {
		t.Fatalf("second Evaluate() error = %v", err)
	}
	if second.Results[0].Actual != 1 {
		t.Fatalf("second result mutated to %#v", second.Results[0])
	}
}

func limitsFor(values map[metrics.Name]int64) budget.Limits {
	var limits budget.Limits
	for name, value := range values {
		owned := value
		switch name {
		case metrics.Nodes:
			limits.Nodes = &owned
		case metrics.TreeDepth:
			limits.TreeDepth = &owned
		case metrics.SceneInstances:
			limits.SceneInstances = &owned
		case metrics.MeshInstances:
			limits.MeshInstances = &owned
		case metrics.Lights:
			limits.Lights = &owned
		case metrics.ShadowLights:
			limits.ShadowLights = &owned
		case metrics.ExternalResources:
			limits.ExternalResources = &owned
		case metrics.SceneDependencies:
			limits.SceneDependencies = &owned
		}
	}
	return limits
}

func metricValues(values map[metrics.Name]int64) metrics.Values {
	var result metrics.Values
	for name, value := range values {
		switch name {
		case metrics.Nodes:
			result.Nodes = value
		case metrics.TreeDepth:
			result.TreeDepth = value
		case metrics.SceneInstances:
			result.SceneInstances = value
		case metrics.MeshInstances:
			result.MeshInstances = value
		case metrics.Lights:
			result.Lights = value
		case metrics.ShadowLights:
			result.ShadowLights = value
		case metrics.ExternalResources:
			result.ExternalResources = value
		case metrics.SceneDependencies:
			result.SceneDependencies = value
		}
	}
	return result
}
