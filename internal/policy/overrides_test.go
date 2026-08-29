package policy

import (
	"math"
	"strconv"
	"testing"

	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

func TestParseBudgetOverridesAllMetricsAndBoundaries(t *testing.T) {
	t.Parallel()

	values := make([]string, 0, len(metrics.OrderedNames())+1)
	for index, name := range metrics.OrderedNames() {
		values = append(values, string(name)+"="+strconv.Itoa(index))
	}
	values = append(values, "nodes="+strconv.FormatInt(math.MaxInt64, 10))

	limits, err := ParseBudgetOverrides("cli", values)
	if err != nil {
		t.Fatalf("ParseBudgetOverrides() error = %v", err)
	}
	for index, name := range metrics.OrderedNames() {
		want := int64(index)
		if name == metrics.Nodes {
			want = math.MaxInt64
		}
		got, configured := limits.Get(name)
		if !configured || got != want {
			t.Errorf("limit %q = %d, %v; want %d, true", name, got, configured, want)
		}
	}

	*limits.Nodes = 0
	again, err := ParseBudgetOverrides("cli", values)
	if err != nil {
		t.Fatalf("second ParseBudgetOverrides() error = %v", err)
	}
	if got, _ := again.Get(metrics.Nodes); got != math.MaxInt64 {
		t.Fatalf("second nodes = %d, want %d", got, int64(math.MaxInt64))
	}
}

func TestParseBudgetOverridesRejectsMalformedValues(t *testing.T) {
	t.Parallel()

	tests := []string{
		"",
		"nodes",
		"nodes=1=2",
		"=1",
		"unknown=1",
		"nodes=",
		"nodes=-1",
		"nodes=+1",
		"nodes=1.0",
		"nodes=true",
		"nodes= 1",
		" nodes=1",
		"nodes=1 ",
		"nodes\t=1",
		"nodes=9223372036854775808",
	}

	for index, value := range tests {
		index, value := index, value
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			t.Parallel()

			got, err := ParseBudgetOverrides("config.json", []string{"lights=1", value})
			if got != (budget.Limits{}) {
				t.Fatalf("limits = %#v, want zero", got)
			}
			configErr := requirePolicyError(t, err, "config.json", "cli.budgets[1]", strconv.Quote(value))
			if configErr.Detail == "" {
				t.Fatal("error detail is empty")
			}
		})
	}
}
