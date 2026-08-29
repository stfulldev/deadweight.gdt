package budget

import (
	"fmt"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

// Status is the frozen non-fatal check verdict.
type Status string

const (
	StatusPassed     Status = "PASSED"
	StatusFailed     Status = "FAILED"
	StatusIncomplete Status = "INCOMPLETE"
)

// Valid reports whether status is part of the MVP check taxonomy.
func (status Status) Valid() bool {
	return status == StatusPassed || status == StatusFailed || status == StatusIncomplete
}

// ReliabilityError reports an unsupported analysis reliability value.
type ReliabilityError struct {
	Reliability analysis.Reliability
}

func (err *ReliabilityError) Error() string {
	return fmt.Sprintf("invalid budget evaluation reliability %q", err.Reliability)
}

// Evaluation is one owned reliability-aware set of ordered comparisons.
type Evaluation struct {
	Status        Status
	Reliability   analysis.Reliability
	FailOnPartial bool
	Exceeded      int
	Results       []Result
}

// Clone returns an independent evaluation result slice.
func (evaluation Evaluation) Clone() Evaluation {
	evaluation.Results = append([]Result(nil), evaluation.Results...)
	return evaluation
}

// Evaluate validates domain inputs, performs inclusive comparisons, and
// applies the frozen non-fatal verdict priority.
func Evaluate(
	values metrics.Values,
	limits Limits,
	reliability analysis.Reliability,
	failOnPartial bool,
) (Evaluation, error) {
	if err := values.Validate(); err != nil {
		return Evaluation{}, err
	}
	if err := limits.Validate(); err != nil {
		return Evaluation{}, err
	}
	if !reliability.Valid() {
		return Evaluation{}, &ReliabilityError{Reliability: reliability}
	}

	results := Check(values, limits)
	exceeded := 0
	for _, result := range results {
		if !result.Passed {
			exceeded++
		}
	}

	status := StatusPassed
	if reliability != analysis.ReliabilityExact && failOnPartial {
		status = StatusIncomplete
	} else if exceeded > 0 {
		status = StatusFailed
	}

	evaluation := Evaluation{
		Status:        status,
		Reliability:   reliability,
		FailOnPartial: failOnPartial,
		Exceeded:      exceeded,
		Results:       results,
	}
	return evaluation.Clone(), nil
}
