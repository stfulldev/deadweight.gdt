package reportdiff

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

type wireDocument struct {
	SchemaVersion  *int            `json:"schema_version"`
	Kind           Kind            `json:"kind"`
	Tool           *wireTool       `json:"tool"`
	Scene          *wireScene      `json:"scene"`
	Configuration  json.RawMessage `json:"configuration"`
	Analysis       *wireAnalysis   `json:"analysis"`
	Policy         json.RawMessage `json:"policy"`
	Evaluation     *wireEvaluation `json:"evaluation"`
	DependencyTree json.RawMessage `json:"dependency_tree"`
	Error          json.RawMessage `json:"error"`
}

type wireTool struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type wireScene struct {
	Path string `json:"path"`
}

type wireAnalysis struct {
	Status         analysis.AnalysisStatus `json:"status"`
	Reliability    analysis.Reliability    `json:"reliability"`
	Metrics        *[]wireMetric           `json:"metrics"`
	Coverage       *wireCoverage           `json:"coverage"`
	Diagnostics    *[]wireDiagnostic       `json:"diagnostics"`
	Contributions  *[]json.RawMessage      `json:"contributions"`
	UniqueEvidence *[]wireUniqueEvidence   `json:"unique_evidence"`
}

type wireMetric struct {
	ID         metrics.Name    `json:"id"`
	Value      *int64          `json:"value"`
	Confidence *wireConfidence `json:"confidence"`
}

type wireConfidence struct {
	Reliability analysis.Reliability         `json:"reliability"`
	Reasons     *[]analysis.ConfidenceReason `json:"reasons"`
}

type wireCoverage struct {
	ParsedSceneFiles         int64 `json:"parsed_scene_files"`
	ResolvedSceneInstances   int64 `json:"resolved_scene_instances"`
	UnresolvedSceneInstances int64 `json:"unresolved_scene_instances"`
	InheritedScenes          int64 `json:"inherited_scenes"`
}

type wireDiagnostic struct {
	Code        diagnostic.Code     `json:"code"`
	Severity    diagnostic.Severity `json:"severity"`
	Message     string              `json:"message"`
	Occurrences *int64              `json:"occurrences"`
	Source      *wireSource         `json:"source"`
	Resource    string              `json:"resource"`
}

type wireSource struct {
	Path   string `json:"path"`
	Line   int64  `json:"line"`
	Column int64  `json:"column"`
}

type wireUniqueEvidence struct {
	Metric    metrics.Name          `json:"metric"`
	Identity  string                `json:"identity"`
	Referrers *[]wireUniqueReferrer `json:"referrers"`
}

type wireUniqueReferrer struct {
	Scene       string            `json:"scene"`
	ResourceID  string            `json:"resource_id"`
	RawTarget   string            `json:"raw_target"`
	EdgeKind    analysis.EdgeKind `json:"edge_kind"`
	Occurrences *int64            `json:"occurrences"`
}

type wireEvaluation struct {
	Comparisons *[]wireComparison `json:"comparisons"`
	Exceeded    *int64            `json:"exceeded"`
	Verdict     budget.Status     `json:"verdict"`
}

type wireComparison struct {
	Metric   metrics.Name `json:"metric"`
	Observed *int64       `json:"observed"`
	Limit    *int64       `json:"limit"`
	Delta    *int64       `json:"delta"`
	Passed   *bool        `json:"passed"`
}

// Decode reads one complete compatible schema-v1 report and owns its semantics.
func Decode(contents []byte) (Snapshot, error) {
	if len(contents) > MaxInputBytes {
		return Snapshot{}, fmt.Errorf("report exceeds %d-byte input limit", MaxInputBytes)
	}
	if !utf8.Valid(contents) {
		return Snapshot{}, errors.New("report is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	var document wireDocument
	if err := decoder.Decode(&document); err != nil {
		return Snapshot{}, fmt.Errorf("decode report JSON: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Snapshot{}, errors.New("report contains trailing JSON content")
		}
		return Snapshot{}, fmt.Errorf("decode trailing report JSON: %w", err)
	}
	return normalizeDocument(document)
}

func normalizeDocument(document wireDocument) (Snapshot, error) {
	if document.SchemaVersion == nil || *document.SchemaVersion != SchemaVersion {
		if document.SchemaVersion == nil {
			return Snapshot{}, errors.New("report is missing schema_version")
		}
		return Snapshot{}, fmt.Errorf("unsupported report schema_version %d; want %d", *document.SchemaVersion, SchemaVersion)
	}
	if !document.Kind.Valid() {
		return Snapshot{}, fmt.Errorf("unsupported report kind %q; want inspect, tree, or check", document.Kind)
	}
	if document.Tool == nil || document.Tool.Name != "deadweight.gdt" || document.Tool.Version == "" {
		return Snapshot{}, errors.New("report has invalid deadweight.gdt tool identity")
	}
	if document.Scene == nil || !portablePath(document.Scene.Path) {
		return Snapshot{}, errors.New("report scene must be a portable res:// path")
	}
	if len(document.Configuration) == 0 || string(document.Configuration) == "null" {
		return Snapshot{}, errors.New("report is missing configuration evidence")
	}
	if document.Analysis == nil {
		return Snapshot{}, errors.New("report is missing analysis evidence")
	}
	if err := validatePayloadShape(document); err != nil {
		return Snapshot{}, err
	}
	return normalizeAnalysis(document.Kind, document.Scene.Path, *document.Analysis, document.Evaluation)
}

func validatePayloadShape(document wireDocument) error {
	hasPolicy := len(document.Policy) != 0 && string(document.Policy) != "null"
	hasTree := len(document.DependencyTree) != 0 && string(document.DependencyTree) != "null"
	hasError := len(document.Error) != 0 && string(document.Error) != "null"
	switch document.Kind {
	case KindInspect:
		if hasPolicy || document.Evaluation != nil || hasTree || hasError {
			return errors.New("inspect report contains an incompatible payload")
		}
	case KindTree:
		if !hasTree || hasPolicy || document.Evaluation != nil || hasError {
			return errors.New("tree report has missing or incompatible payloads")
		}
	case KindCheck:
		if !hasPolicy || document.Evaluation == nil || hasTree || hasError {
			return errors.New("check report has missing or incompatible payloads")
		}
	}
	return nil
}

func normalizeAnalysis(kind Kind, scene string, value wireAnalysis, evaluation *wireEvaluation) (Snapshot, error) {
	if !value.Status.Valid() || !value.Reliability.Valid() {
		return Snapshot{}, errors.New("report has invalid analysis status or reliability")
	}
	if (value.Reliability == analysis.ReliabilityExact) != (value.Status == analysis.AnalysisComplete) {
		return Snapshot{}, errors.New("report analysis status and reliability are inconsistent")
	}
	if value.Metrics == nil || len(*value.Metrics) != len(metrics.OrderedNames()) {
		return Snapshot{}, fmt.Errorf("report must contain exactly %d frozen metrics", len(metrics.OrderedNames()))
	}
	metricSnapshots := make([]MetricSnapshot, 0, len(*value.Metrics))
	for index, item := range *value.Metrics {
		want := metrics.OrderedNames()[index]
		if item.ID != want || item.Value == nil || *item.Value < 0 {
			return Snapshot{}, fmt.Errorf("invalid metric at position %d; want %q with a non-negative value", index, want)
		}
		confidence, err := normalizeConfidence(item.Confidence, value.Reliability)
		if err != nil {
			return Snapshot{}, fmt.Errorf("metric %q: %w", item.ID, err)
		}
		metricSnapshots = append(metricSnapshots, MetricSnapshot{Metric: item.ID, Value: *item.Value, Confidence: confidence})
	}
	if conservativeMetricReliability(metricSnapshots) != value.Reliability {
		return Snapshot{}, errors.New("report reliability does not match its metric confidence summary")
	}
	if value.Coverage == nil {
		return Snapshot{}, errors.New("report is missing coverage")
	}
	coverage := analysis.Coverage{
		ParsedSceneFiles:         value.Coverage.ParsedSceneFiles,
		ResolvedSceneInstances:   value.Coverage.ResolvedSceneInstances,
		UnresolvedSceneInstances: value.Coverage.UnresolvedSceneInstances,
		InheritedScenes:          value.Coverage.InheritedScenes,
	}
	if err := coverage.Validate(); err != nil {
		return Snapshot{}, err
	}
	if value.Diagnostics == nil || value.Contributions == nil || len(*value.Contributions) == 0 || value.UniqueEvidence == nil {
		return Snapshot{}, errors.New("report is missing required analysis collections")
	}
	diagnostics, err := normalizeDiagnostics(*value.Diagnostics)
	if err != nil {
		return Snapshot{}, err
	}
	dependencies, err := normalizeUniqueEvidence(*value.UniqueEvidence)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{
		Kind: kind, Scene: scene, Reliability: value.Reliability,
		Metrics: metricSnapshots, Coverage: coverage,
		Diagnostics: diagnostics, Dependencies: dependencies,
	}
	if kind == KindCheck {
		normalized, evalErr := normalizeEvaluation(*evaluation, metricSnapshots, value.Reliability)
		if evalErr != nil {
			return Snapshot{}, evalErr
		}
		snapshot.Evaluation = &normalized
	}
	return snapshot, nil
}

func normalizeConfidence(value *wireConfidence, fallback analysis.Reliability) (Confidence, error) {
	if value == nil {
		return Confidence{Reliability: fallback, Source: ConfidenceReportSummary}, nil
	}
	if value.Reasons == nil {
		return Confidence{}, errors.New("confidence is missing reasons")
	}
	confidence := analysis.Confidence{Reliability: value.Reliability, Reasons: append([]analysis.ConfidenceReason(nil), (*value.Reasons)...)}
	if err := confidence.Validate(); err != nil {
		return Confidence{}, err
	}
	return Confidence{Reliability: confidence.Reliability, Reasons: confidence.Reasons, Source: ConfidenceMetric}, nil
}

func conservativeMetricReliability(items []MetricSnapshot) analysis.Reliability {
	result := analysis.ReliabilityExact
	for _, item := range items {
		if reliabilityRank(item.Confidence.Reliability) > reliabilityRank(result) {
			result = item.Confidence.Reliability
		}
	}
	return result
}

func normalizeDiagnostics(items []wireDiagnostic) ([]Diagnostic, error) {
	result := make([]Diagnostic, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Occurrences == nil || *item.Occurrences <= 0 || strings.TrimSpace(item.Message) == "" {
			return nil, errors.New("report contains an invalid diagnostic")
		}
		if !item.Code.Valid() || !item.Severity.Valid() {
			return nil, errors.New("report contains an invalid diagnostic taxonomy value")
		}
		want, _ := item.Code.Severity()
		if want != item.Severity {
			return nil, fmt.Errorf("diagnostic %q has inconsistent severity", item.Code)
		}
		normalized := Diagnostic{Code: item.Code, Severity: item.Severity, Message: item.Message, Occurrences: *item.Occurrences, Resource: item.Resource}
		if item.Source != nil {
			if !portablePath(item.Source.Path) || item.Source.Line < 0 || item.Source.Column < 0 {
				return nil, errors.New("report contains an invalid diagnostic source")
			}
			normalized.SourcePath, normalized.Line, normalized.Column = item.Source.Path, item.Source.Line, item.Source.Column
		}
		if strings.Contains(item.Resource, `\`) {
			return nil, errors.New("report contains a non-portable diagnostic resource")
		}
		key := diagnosticKey(normalized)
		if _, duplicate := seen[key]; duplicate {
			return nil, errors.New("report contains duplicate grouped diagnostics")
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}
	sort.Slice(result, func(left, right int) bool { return diagnosticKey(result[left]) < diagnosticKey(result[right]) })
	return result, nil
}

func normalizeUniqueEvidence(items []wireUniqueEvidence) ([]string, error) {
	dependencies := make([]string, 0)
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Metric != metrics.ExternalResources && item.Metric != metrics.SceneDependencies {
			return nil, fmt.Errorf("invalid unique evidence metric %q", item.Metric)
		}
		if item.Identity == "" || strings.Contains(item.Identity, `\`) || item.Referrers == nil || len(*item.Referrers) == 0 {
			return nil, errors.New("report contains invalid unique evidence")
		}
		key := string(item.Metric) + "\x00" + item.Identity
		if _, duplicate := seen[key]; duplicate {
			return nil, errors.New("report contains duplicate unique evidence")
		}
		seen[key] = struct{}{}
		for _, referrer := range *item.Referrers {
			if !portablePath(referrer.Scene) || referrer.Occurrences == nil || *referrer.Occurrences <= 0 || strings.Contains(referrer.RawTarget, `\`) {
				return nil, errors.New("report contains an invalid unique evidence referrer")
			}
			if referrer.EdgeKind != "" && referrer.EdgeKind != analysis.EdgeInstance && referrer.EdgeKind != analysis.EdgeInheritance {
				return nil, errors.New("report contains an invalid unique evidence edge kind")
			}
		}
		if item.Metric == metrics.SceneDependencies {
			if !portablePath(item.Identity) {
				return nil, errors.New("scene dependency identity must be a portable res:// path")
			}
			dependencies = append(dependencies, item.Identity)
		}
	}
	sort.Strings(dependencies)
	return dependencies, nil
}

func normalizeEvaluation(value wireEvaluation, metricSnapshots []MetricSnapshot, reliability analysis.Reliability) (Evaluation, error) {
	if !value.Verdict.Valid() || value.Comparisons == nil || value.Exceeded == nil || *value.Exceeded < 0 {
		return Evaluation{}, errors.New("report contains an invalid check evaluation")
	}
	observed := make(map[metrics.Name]int64, len(metricSnapshots))
	for _, item := range metricSnapshots {
		observed[item.Metric] = item.Value
	}
	comparisons := make([]budget.Result, 0, len(*value.Comparisons))
	previousRank := -1
	exceeded := int64(0)
	for _, item := range *value.Comparisons {
		rank := metricRank(item.Metric)
		if rank <= previousRank || item.Observed == nil || item.Limit == nil || item.Delta == nil || item.Passed == nil || *item.Observed < 0 || *item.Limit < 0 {
			return Evaluation{}, errors.New("report contains invalid or non-canonical budget comparisons")
		}
		previousRank = rank
		if want, ok := observed[item.Metric]; !ok || want != *item.Observed || *item.Delta != *item.Observed-*item.Limit || *item.Passed != (*item.Observed <= *item.Limit) {
			return Evaluation{}, fmt.Errorf("comparison %q is inconsistent", item.Metric)
		}
		if !*item.Passed {
			exceeded++
		}
		comparisons = append(comparisons, budget.Result{Metric: item.Metric, Actual: *item.Observed, Limit: *item.Limit, Delta: *item.Delta, Passed: *item.Passed})
	}
	if exceeded != *value.Exceeded {
		return Evaluation{}, errors.New("evaluation exceeded count is inconsistent")
	}
	if value.Verdict == budget.StatusPassed && exceeded != 0 || value.Verdict == budget.StatusFailed && exceeded == 0 || value.Verdict == budget.StatusIncomplete && reliability == analysis.ReliabilityExact {
		return Evaluation{}, errors.New("evaluation verdict is inconsistent")
	}
	return Evaluation{Verdict: value.Verdict, Exceeded: exceeded, Comparisons: comparisons}, nil
}

func portablePath(value string) bool {
	if !strings.HasPrefix(value, "res://") || strings.Contains(value, `\`) {
		return false
	}
	relative := strings.TrimPrefix(value, "res://")
	return relative != "" && relative != "." && relative != ".." && !strings.HasPrefix(relative, "../") && path.Clean(relative) == relative
}

func metricRank(name metrics.Name) int {
	for index, candidate := range metrics.OrderedNames() {
		if name == candidate {
			return index
		}
	}
	return -1
}

func reliabilityRank(value analysis.Reliability) int {
	switch value {
	case analysis.ReliabilityExact:
		return 0
	case analysis.ReliabilityLowerBound:
		return 1
	case analysis.ReliabilityApproximate:
		return 2
	default:
		return -1
	}
}

func diagnosticKey(item Diagnostic) string {
	return strings.Join([]string{string(item.Severity), string(item.Code), item.SourcePath, fmt.Sprint(item.Line), fmt.Sprint(item.Column), item.Resource, item.Message}, "\x00")
}
