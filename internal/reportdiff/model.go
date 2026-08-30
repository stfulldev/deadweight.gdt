// Package reportdiff decodes and compares portable deadweight.gdt reports.
package reportdiff

import (
	"fmt"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

const (
	// SchemaVersion is the report schema version understood by the diff reader.
	SchemaVersion = 1
	// MaxInputBytes bounds each report before decoding.
	MaxInputBytes = 16 << 20
)

// Kind is a comparable producer report kind.
type Kind string

const (
	KindInspect Kind = "inspect"
	KindTree    Kind = "tree"
	KindCheck   Kind = "check"
)

// Valid reports whether kind can be used as a baseline.
func (kind Kind) Valid() bool {
	return kind == KindInspect || kind == KindTree || kind == KindCheck
}

// ConfidenceSource identifies where effective per-metric confidence came from.
type ConfidenceSource string

const (
	ConfidenceMetric        ConfidenceSource = "metric"
	ConfidenceReportSummary ConfidenceSource = "report_summary"
)

// Confidence is owned confidence evidence for one reported metric.
type Confidence struct {
	Reliability analysis.Reliability
	Reasons     []analysis.ConfidenceReason
	Source      ConfidenceSource
}

// MetricSnapshot is one normalized frozen metric.
type MetricSnapshot struct {
	Metric     metrics.Name
	Value      int64
	Confidence Confidence
}

// Diagnostic is one portable grouped diagnostic.
type Diagnostic struct {
	Code        diagnostic.Code
	Severity    diagnostic.Severity
	Message     string
	Occurrences int64
	SourcePath  string
	Line        int64
	Column      int64
	Resource    string
}

// Evaluation is one normalized check outcome.
type Evaluation struct {
	Verdict     budget.Status
	Exceeded    int64
	Comparisons []budget.Result
}

// Snapshot is the owned semantic evidence read from one report.
type Snapshot struct {
	Kind         Kind
	Scene        string
	Reliability  analysis.Reliability
	Metrics      []MetricSnapshot
	Coverage     analysis.Coverage
	Diagnostics  []Diagnostic
	Dependencies []string
	Evaluation   *Evaluation
}

// Assessment qualifies a numerical metric change.
type Assessment string

const (
	AssessmentRegression  Assessment = "regression"
	AssessmentImprovement Assessment = "improvement"
	AssessmentUncertain   Assessment = "uncertain"
)

// MetricChange is one semantic metric value or confidence change.
type MetricChange struct {
	Metric           metrics.Name
	Before           int64
	After            int64
	Delta            int64
	BeforeConfidence Confidence
	AfterConfidence  Confidence
	Assessment       Assessment
}

// ReliabilityChange records a report-wide confidence change.
type ReliabilityChange struct {
	Before analysis.Reliability
	After  analysis.Reliability
}

// CoverageChange records one signed coverage delta.
type CoverageChange struct {
	Field  string
	Before int64
	After  int64
	Delta  int64
}

// EvidenceChange identifies whether portable evidence was added, removed, or changed.
type EvidenceChange string

const (
	EvidenceAdded              EvidenceChange = "added"
	EvidenceRemoved            EvidenceChange = "removed"
	EvidenceOccurrencesChanged EvidenceChange = "occurrences_changed"
)

// DiagnosticChange is one grouped diagnostic semantic change.
type DiagnosticChange struct {
	Change            EvidenceChange
	Diagnostic        Diagnostic
	BeforeOccurrences int64
	AfterOccurrences  int64
	Delta             int64
}

// DependencyChange is one portable scene dependency identity change.
type DependencyChange struct {
	Change   EvidenceChange
	Identity string
}

// EvaluationChange retains complete before/after check evidence when it changed.
type EvaluationChange struct {
	Before Evaluation
	After  Evaluation
}

// TriggerKind is one opt-in enforcement reason.
type TriggerKind string

const (
	TriggerMetricIncrease         TriggerKind = "metric_increase"
	TriggerReliabilityDegradation TriggerKind = "reliability_degradation"
)

// Trigger is one deterministic enforcement reason.
type Trigger struct {
	Kind              TriggerKind
	Metric            metrics.Name
	Assessment        Assessment
	BeforeReliability analysis.Reliability
	AfterReliability  analysis.Reliability
}

// Enforcement is the policy-independent diff's opt-in outcome.
type Enforcement struct {
	Enabled  bool
	Status   budget.Status
	Triggers []Trigger
}

// Policy selects regression conditions that affect the process outcome.
type Policy struct {
	MetricIncreases   []metrics.Name
	FailOnReliability bool
}

// NormalizePolicy validates, owns, and orders a policy in frozen metric order.
func NormalizePolicy(policy Policy) (Policy, error) {
	selected := make(map[metrics.Name]struct{}, len(policy.MetricIncreases))
	for _, name := range policy.MetricIncreases {
		if !name.Valid() {
			return Policy{}, fmt.Errorf("unknown metric %q", name)
		}
		if _, duplicate := selected[name]; duplicate {
			return Policy{}, fmt.Errorf("duplicate metric %q", name)
		}
		selected[name] = struct{}{}
	}
	normalized := Policy{FailOnReliability: policy.FailOnReliability}
	for _, name := range metrics.OrderedNames() {
		if _, ok := selected[name]; ok {
			normalized.MetricIncreases = append(normalized.MetricIncreases, name)
		}
	}
	return normalized, nil
}

// Result is an owned deterministic semantic comparison.
type Result struct {
	Kind              Kind
	Scene             string
	BeforeReliability analysis.Reliability
	AfterReliability  analysis.Reliability
	MetricChanges     []MetricChange
	ReliabilityChange *ReliabilityChange
	CoverageChanges   []CoverageChange
	DiagnosticChanges []DiagnosticChange
	DependencyChanges []DependencyChange
	EvaluationChange  *EvaluationChange
	Changed           bool
	Enforcement       Enforcement
}
