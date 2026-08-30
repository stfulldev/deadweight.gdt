package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path"
	"path/filepath"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	application "github.com/stfulldev/deadweight.gdt/internal/app"
	"github.com/stfulldev/deadweight.gdt/internal/budget"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/policy"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/reportdiff"
)

const reportSchemaVersion = 1

type documentV1 struct {
	SchemaVersion  int               `json:"schema_version"`
	Kind           string            `json:"kind"`
	Tool           toolV1            `json:"tool"`
	Scene          *sceneV1          `json:"scene,omitempty"`
	Configuration  *configurationV1  `json:"configuration,omitempty"`
	Analysis       *analysisV1       `json:"analysis,omitempty"`
	Policy         *policyV1         `json:"policy,omitempty"`
	Evaluation     *evaluationV1     `json:"evaluation,omitempty"`
	DependencyTree *dependencyTreeV1 `json:"dependency_tree,omitempty"`
	Error          *fatalErrorV1     `json:"error,omitempty"`
	Diff           *diffV1           `json:"diff,omitempty"`
}

type toolV1 struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type sceneV1 struct {
	Path string `json:"path"`
}

type configurationV1 struct {
	Present   bool   `json:"present"`
	Selection string `json:"selection"`
	Path      string `json:"path,omitempty"`
}

type analysisV1 struct {
	Status          analysis.AnalysisStatus `json:"status"`
	Reliability     analysis.Reliability    `json:"reliability"`
	Metrics         []metricV1              `json:"metrics"`
	Coverage        coverageV1              `json:"coverage"`
	Diagnostics     []diagnosticV1          `json:"diagnostics"`
	Contributions   []contributionV1        `json:"contributions"`
	UniqueEvidence  []uniqueEvidenceV1      `json:"unique_evidence"`
	TopContributors *topContributorsV1      `json:"top_contributors,omitempty"`
}

type metricV1 struct {
	ID         metrics.Name `json:"id"`
	Value      int64        `json:"value"`
	Confidence confidenceV1 `json:"confidence"`
}

type confidenceV1 struct {
	Reliability analysis.Reliability        `json:"reliability"`
	Reasons     []analysis.ConfidenceReason `json:"reasons"`
}

type coverageV1 struct {
	ParsedSceneFiles         int64 `json:"parsed_scene_files"`
	ResolvedSceneInstances   int64 `json:"resolved_scene_instances"`
	UnresolvedSceneInstances int64 `json:"unresolved_scene_instances"`
	InheritedScenes          int64 `json:"inherited_scenes"`
}

type diagnosticV1 struct {
	Code        diagnostic.Code     `json:"code"`
	Severity    diagnostic.Severity `json:"severity"`
	Message     string              `json:"message"`
	Occurrences int64               `json:"occurrences"`
	Source      *sourceV1           `json:"source,omitempty"`
	Resource    string              `json:"resource,omitempty"`
}

type sourceV1 struct {
	Path   string `json:"path"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
}

type policyV1 struct {
	Kind          string           `json:"kind"`
	ID            string           `json:"id,omitempty"`
	Metadata      policyMetadataV1 `json:"metadata"`
	FailOnPartial bool             `json:"fail_on_partial"`
}

type policyMetadataV1 struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Platform    string `json:"platform,omitempty"`
	Renderer    string `json:"renderer,omitempty"`
	TargetFPS   int64  `json:"target_fps,omitempty"`
	Quality     string `json:"quality,omitempty"`
	Status      string `json:"status,omitempty"`
	Stability   string `json:"stability,omitempty"`
}

type evaluationV1 struct {
	Comparisons []comparisonV1 `json:"comparisons"`
	Exceeded    int64          `json:"exceeded"`
	Verdict     budget.Status  `json:"verdict"`
}

type comparisonV1 struct {
	Metric   metrics.Name `json:"metric"`
	Observed int64        `json:"observed"`
	Limit    int64        `json:"limit"`
	Delta    int64        `json:"delta"`
	Passed   bool         `json:"passed"`
}

type fatalErrorV1 struct {
	Code     diagnostic.Code     `json:"code,omitempty"`
	Severity diagnostic.Severity `json:"severity"`
	Message  string              `json:"message"`
	Details  []string            `json:"details,omitempty"`
}

type dependencyTreeV1 struct {
	Root    string                  `json:"root"`
	Entries []dependencyTreeEntryV1 `json:"entries"`
}

type dependencyTreeEntryV1 struct {
	Depth            int64                         `json:"depth"`
	Source           string                        `json:"source"`
	Target           string                        `json:"target"`
	Kind             analysis.EdgeKind             `json:"kind"`
	Resolved         bool                          `json:"resolved"`
	Occurrences      int64                         `json:"occurrences"`
	Reliability      analysis.Reliability          `json:"reliability"`
	BackReference    bool                          `json:"back_reference"`
	Classification   analysis.TargetClassification `json:"classification,omitempty"`
	ResourceID       string                        `json:"resource_id,omitempty"`
	RawTarget        string                        `json:"raw_target,omitempty"`
	ResolutionReason project.ResolutionReason      `json:"resolution_reason,omitempty"`
}

type diffV1 struct {
	ReportKind        reportdiff.Kind         `json:"report_kind"`
	Scene             string                  `json:"scene"`
	BeforeReliability analysis.Reliability    `json:"before_reliability"`
	AfterReliability  analysis.Reliability    `json:"after_reliability"`
	Changed           bool                    `json:"changed"`
	Metrics           []diffMetricV1          `json:"metrics"`
	Reliability       *diffReliabilityV1      `json:"reliability,omitempty"`
	Coverage          []diffCoverageV1        `json:"coverage"`
	Diagnostics       []diffDiagnosticV1      `json:"diagnostics"`
	Dependencies      []diffDependencyV1      `json:"dependencies"`
	Evaluation        *diffEvaluationChangeV1 `json:"evaluation,omitempty"`
	Enforcement       diffEnforcementV1       `json:"enforcement"`
}

type diffMetricV1 struct {
	Metric           metrics.Name          `json:"metric"`
	Before           int64                 `json:"before"`
	After            int64                 `json:"after"`
	Delta            int64                 `json:"delta"`
	BeforeConfidence diffConfidenceV1      `json:"before_confidence"`
	AfterConfidence  diffConfidenceV1      `json:"after_confidence"`
	Assessment       reportdiff.Assessment `json:"assessment"`
}

type diffConfidenceV1 struct {
	Reliability analysis.Reliability        `json:"reliability"`
	Reasons     []analysis.ConfidenceReason `json:"reasons"`
	Source      reportdiff.ConfidenceSource `json:"source"`
}

type diffReliabilityV1 struct {
	Before analysis.Reliability `json:"before"`
	After  analysis.Reliability `json:"after"`
}

type diffCoverageV1 struct {
	Field  string `json:"field"`
	Before int64  `json:"before"`
	After  int64  `json:"after"`
	Delta  int64  `json:"delta"`
}

type diffDiagnosticV1 struct {
	Change            reportdiff.EvidenceChange `json:"change"`
	Code              diagnostic.Code           `json:"code"`
	Severity          diagnostic.Severity       `json:"severity"`
	Message           string                    `json:"message"`
	Source            *sourceV1                 `json:"source,omitempty"`
	Resource          string                    `json:"resource,omitempty"`
	BeforeOccurrences int64                     `json:"before_occurrences"`
	AfterOccurrences  int64                     `json:"after_occurrences"`
	Delta             int64                     `json:"delta"`
}

type diffDependencyV1 struct {
	Change   reportdiff.EvidenceChange `json:"change"`
	Identity string                    `json:"identity"`
}

type diffEvaluationChangeV1 struct {
	Before diffEvaluationV1 `json:"before"`
	After  diffEvaluationV1 `json:"after"`
}

type diffEvaluationV1 struct {
	Verdict     budget.Status  `json:"verdict"`
	Exceeded    int64          `json:"exceeded"`
	Comparisons []comparisonV1 `json:"comparisons"`
}

type diffEnforcementV1 struct {
	Enabled  bool            `json:"enabled"`
	Status   budget.Status   `json:"status"`
	Triggers []diffTriggerV1 `json:"triggers"`
}

type diffTriggerV1 struct {
	Kind              reportdiff.TriggerKind `json:"kind"`
	Metric            metrics.Name           `json:"metric,omitempty"`
	Assessment        reportdiff.Assessment  `json:"assessment,omitempty"`
	BeforeReliability analysis.Reliability   `json:"before_reliability"`
	AfterReliability  analysis.Reliability   `json:"after_reliability"`
}

// InspectJSON renders one portable schema-version-one inspect document.
func InspectJSON(result application.InspectResult, options Options) (string, error) {
	document, err := inspectDocumentV1(result, options)
	if err != nil {
		return "", err
	}

	return encodeDocumentV1(document)
}

// TreeJSON renders one portable schema-version-one dependency-tree document.
func TreeJSON(result application.TreeResult, options Options) (string, error) {
	tree, err := projectDependencyTree(result)
	if err != nil {
		return "", err
	}
	document, err := inspectDocumentV1(result.Inspect, options)
	if err != nil {
		return "", err
	}
	entries := make([]dependencyTreeEntryV1, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		entries = append(entries, dependencyTreeEntryV1{
			Depth:            entry.Depth,
			Source:           entry.Source,
			Target:           entry.Target,
			Kind:             entry.Kind,
			Resolved:         entry.Resolved,
			Occurrences:      entry.Occurrences,
			Reliability:      entry.Reliability,
			BackReference:    entry.BackReference,
			Classification:   entry.Classification,
			ResourceID:       entry.ResourceID,
			RawTarget:        entry.RawTarget,
			ResolutionReason: entry.ResolutionReason,
		})
	}
	document.Kind = "tree"
	document.DependencyTree = &dependencyTreeV1{Root: tree.Root, Entries: entries}

	return encodeDocumentV1(document)
}

// CheckJSON renders one portable schema-version-one check document.
func CheckJSON(result application.CheckResult, options Options) (string, error) {
	if !result.Policy.Kind.Valid() {
		return "", fmt.Errorf("invalid policy kind %q", result.Policy.Kind)
	}
	if result.Policy.Kind == policy.KindNone && result.Policy.ID != "" {
		return "", fmt.Errorf("override-only policy must not have ID %q", result.Policy.ID)
	}
	if result.Policy.Kind != policy.KindNone && result.Policy.ID == "" {
		return "", fmt.Errorf("policy kind %q requires an ID", result.Policy.Kind)
	}
	if result.Policy.Metadata.TargetFPS < 0 {
		return "", fmt.Errorf("policy target_fps must be non-negative, got %d", result.Policy.Metadata.TargetFPS)
	}
	if !result.Evaluation.Status.Valid() {
		return "", fmt.Errorf("invalid check status %q", result.Evaluation.Status)
	}
	if result.Evaluation.Reliability != result.Inspect.Analysis.Reliability {
		return "", fmt.Errorf(
			"evaluation reliability %q does not match analysis reliability %q",
			result.Evaluation.Reliability,
			result.Inspect.Analysis.Reliability,
		)
	}

	inspect, err := inspectDocumentV1(result.Inspect, options)
	if err != nil {
		return "", err
	}
	comparisons, exceeded, err := comparisonsV1(result.Evaluation)
	if err != nil {
		return "", err
	}

	inspect.Kind = "check"
	inspect.Policy = &policyV1{
		Kind:          portablePolicyKind(result.Policy.Kind),
		ID:            result.Policy.ID,
		Metadata:      policyMetadataDocumentV1(result.Policy.Metadata),
		FailOnPartial: result.Evaluation.FailOnPartial,
	}
	inspect.Evaluation = &evaluationV1{
		Comparisons: comparisons,
		Exceeded:    exceeded,
		Verdict:     result.Evaluation.Status,
	}

	return encodeDocumentV1(inspect)
}

// DiffJSON renders one portable schema-version-one semantic diff document.
func DiffJSON(result application.DiffResult, options Options) (string, error) {
	comparison := result.Comparison
	if !comparison.Kind.Valid() || comparison.Scene == "" || !comparison.Enforcement.Status.Valid() {
		return "", fmt.Errorf("invalid diff result")
	}
	metricDocuments := make([]diffMetricV1, 0, len(comparison.MetricChanges))
	for _, change := range comparison.MetricChanges {
		metricDocuments = append(metricDocuments, diffMetricV1{
			Metric: change.Metric, Before: change.Before, After: change.After, Delta: change.Delta,
			BeforeConfidence: diffConfidenceDocumentV1(change.BeforeConfidence),
			AfterConfidence:  diffConfidenceDocumentV1(change.AfterConfidence), Assessment: change.Assessment,
		})
	}
	var reliability *diffReliabilityV1
	if comparison.ReliabilityChange != nil {
		reliability = &diffReliabilityV1{Before: comparison.ReliabilityChange.Before, After: comparison.ReliabilityChange.After}
	}
	options = normalizedOptions(options)
	coverageDocuments := make([]diffCoverageV1, 0, len(comparison.CoverageChanges))
	for _, change := range comparison.CoverageChanges {
		coverageDocuments = append(coverageDocuments, diffCoverageV1{Field: change.Field, Before: change.Before, After: change.After, Delta: change.Delta})
	}
	diagnosticDocuments := make([]diffDiagnosticV1, 0, len(comparison.DiagnosticChanges))
	for _, change := range comparison.DiagnosticChanges {
		document := diffDiagnosticV1{
			Change: change.Change, Code: change.Diagnostic.Code, Severity: change.Diagnostic.Severity,
			Message: change.Diagnostic.Message, Resource: change.Diagnostic.Resource,
			BeforeOccurrences: change.BeforeOccurrences, AfterOccurrences: change.AfterOccurrences, Delta: change.Delta,
		}
		if change.Diagnostic.SourcePath != "" {
			document.Source = &sourceV1{Path: change.Diagnostic.SourcePath, Line: int(change.Diagnostic.Line), Column: int(change.Diagnostic.Column)}
		}
		diagnosticDocuments = append(diagnosticDocuments, document)
	}
	dependencyDocuments := make([]diffDependencyV1, 0, len(comparison.DependencyChanges))
	for _, change := range comparison.DependencyChanges {
		dependencyDocuments = append(dependencyDocuments, diffDependencyV1{Change: change.Change, Identity: change.Identity})
	}
	triggerDocuments := make([]diffTriggerV1, 0, len(comparison.Enforcement.Triggers))
	for _, trigger := range comparison.Enforcement.Triggers {
		triggerDocuments = append(triggerDocuments, diffTriggerV1{
			Kind: trigger.Kind, Metric: trigger.Metric, Assessment: trigger.Assessment,
			BeforeReliability: trigger.BeforeReliability, AfterReliability: trigger.AfterReliability,
		})
	}
	var evaluationDocument *diffEvaluationChangeV1
	if comparison.EvaluationChange != nil {
		evaluationDocument = &diffEvaluationChangeV1{
			Before: diffEvaluationDocumentV1(comparison.EvaluationChange.Before),
			After:  diffEvaluationDocumentV1(comparison.EvaluationChange.After),
		}
	}
	return encodeDocumentV1(documentV1{
		SchemaVersion: reportSchemaVersion, Kind: "diff", Tool: toolDocumentV1(options),
		Diff: &diffV1{
			ReportKind: comparison.Kind, Scene: comparison.Scene,
			BeforeReliability: comparison.BeforeReliability, AfterReliability: comparison.AfterReliability,
			Changed: comparison.Changed, Metrics: metricDocuments, Reliability: reliability,
			Coverage: coverageDocuments, Diagnostics: diagnosticDocuments, Dependencies: dependencyDocuments,
			Evaluation: evaluationDocument,
			Enforcement: diffEnforcementV1{
				Enabled: comparison.Enforcement.Enabled, Status: comparison.Enforcement.Status, Triggers: triggerDocuments,
			},
		},
	})
}

func diffEvaluationDocumentV1(value reportdiff.Evaluation) diffEvaluationV1 {
	comparisons := make([]comparisonV1, 0, len(value.Comparisons))
	for _, item := range value.Comparisons {
		comparisons = append(comparisons, comparisonV1{
			Metric: item.Metric, Observed: item.Actual, Limit: item.Limit, Delta: item.Delta, Passed: item.Passed,
		})
	}
	return diffEvaluationV1{Verdict: value.Verdict, Exceeded: value.Exceeded, Comparisons: comparisons}
}

func diffConfidenceDocumentV1(value reportdiff.Confidence) diffConfidenceV1 {
	return diffConfidenceV1{
		Reliability: value.Reliability,
		Reasons:     append([]analysis.ConfidenceReason{}, value.Reasons...),
		Source:      value.Source,
	}
}

// ErrorJSON renders one schema-version-one fatal diagnostic document.
func ErrorJSON(err error, options Options) (string, error) {
	if err == nil {
		return "", errors.New("fatal error is required")
	}
	options = normalizedOptions(options)
	message := strings.TrimRight(diagnostic.MessageOf(err), "\n")
	if message == "" {
		message = strings.TrimRight(err.Error(), "\n")
	}
	if message == "" {
		return "", errors.New("fatal error message is required")
	}
	lines := strings.Split(message, "\n")
	heading := strings.TrimSpace(lines[0])
	details := make([]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line != "" {
			details = append(details, line)
		}
	}
	fatal := fatalErrorV1{
		Severity: diagnostic.SeverityError,
		Message:  heading,
		Details:  details,
	}
	if code, ok := diagnostic.CodeOf(err); ok {
		fatal.Code = code
	}

	return encodeDocumentV1(documentV1{
		SchemaVersion: reportSchemaVersion,
		Kind:          "error",
		Tool:          toolDocumentV1(options),
		Error:         &fatal,
	})
}

func inspectDocumentV1(result application.InspectResult, options Options) (documentV1, error) {
	if err := validateInspect(result); err != nil {
		return documentV1{}, err
	}
	scenePath, err := portableSceneIdentity(result)
	if err != nil {
		return documentV1{}, err
	}
	metricsDocument := make([]metricV1, 0, len(metrics.OrderedNames()))
	for _, name := range metrics.OrderedNames() {
		value, _ := result.Analysis.Summary.Metrics.Get(name)
		confidence, _ := result.Analysis.MetricConfidence.Get(name)
		metricsDocument = append(metricsDocument, metricV1{
			ID:         name,
			Value:      value,
			Confidence: confidenceDocumentV1(confidence),
		})
	}
	diagnostics := sortedDiagnostics(result.Analysis.Diagnostics)
	diagnosticDocuments := make([]diagnosticV1, 0, len(diagnostics))
	for _, item := range diagnostics {
		diagnosticDocuments = append(diagnosticDocuments, diagnosticDocumentV1(result.Project.Directory, item))
	}
	coverage := result.Analysis.Coverage
	contributions, err := contributionDocumentsV1(result)
	if err != nil {
		return documentV1{}, err
	}
	uniqueEvidence, err := uniqueEvidenceDocumentsV1(result)
	if err != nil {
		return documentV1{}, err
	}
	top, err := topContributorsDocumentV1(result, options.Contributions)
	if err != nil {
		return documentV1{}, err
	}
	options = normalizedOptions(options)

	return documentV1{
		SchemaVersion: reportSchemaVersion,
		Kind:          "inspect",
		Tool:          toolDocumentV1(options),
		Scene:         &sceneV1{Path: scenePath},
		Configuration: configurationDocumentV1(result),
		Analysis: &analysisV1{
			Status:      result.Analysis.Status,
			Reliability: result.Analysis.Reliability,
			Metrics:     metricsDocument,
			Coverage: coverageV1{
				ParsedSceneFiles:         coverage.ParsedSceneFiles,
				ResolvedSceneInstances:   coverage.ResolvedSceneInstances,
				UnresolvedSceneInstances: coverage.UnresolvedSceneInstances,
				InheritedScenes:          coverage.InheritedScenes,
			},
			Diagnostics:     diagnosticDocuments,
			Contributions:   contributions,
			UniqueEvidence:  uniqueEvidence,
			TopContributors: top,
		},
	}, nil
}

func confidenceDocumentV1(confidence analysis.Confidence) confidenceV1 {
	return confidenceV1{
		Reliability: confidence.Reliability,
		Reasons:     append([]analysis.ConfidenceReason{}, confidence.Reasons...),
	}
}

func toolDocumentV1(options Options) toolV1 {
	return toolV1{Name: "deadweight.gdt", Version: options.Version}
}

func configurationDocumentV1(result application.InspectResult) *configurationV1 {
	configuration := &configurationV1{Selection: "absent"}
	if !result.ConfigPresent {
		return configuration
	}
	configuration.Present = true
	configuration.Selection = "implicit"
	if result.ConfigSource.Explicit {
		configuration.Selection = "explicit"
	}
	configuration.Path, _ = portableProjectPath(result.Project.Directory, result.ConfigSource.Path)
	return configuration
}

func diagnosticDocumentV1(projectRoot string, item diagnostic.Diagnostic) diagnosticV1 {
	occurrences := item.Occurrences
	if occurrences == 0 {
		occurrences = 1
	}
	result := diagnosticV1{
		Code:        item.Code,
		Severity:    item.Severity,
		Message:     item.Message,
		Occurrences: occurrences,
	}
	if source, ok := portableOptionalPath(projectRoot, item.File); ok {
		result.Source = &sourceV1{Path: source, Line: item.Line, Column: item.Column}
	}
	if resource, ok := portableOptionalResource(projectRoot, item.Resource); ok {
		result.Resource = resource
	}
	return result
}

func policyMetadataDocumentV1(metadata policy.Metadata) policyMetadataV1 {
	return policyMetadataV1{
		Name:        metadata.Name,
		Description: metadata.Description,
		Platform:    metadata.Platform,
		Renderer:    metadata.Renderer,
		TargetFPS:   metadata.TargetFPS,
		Quality:     metadata.Quality,
		Status:      metadata.Status,
		Stability:   metadata.Stability,
	}
}

func comparisonsV1(evaluation budget.Evaluation) ([]comparisonV1, int64, error) {
	comparisons, err := sortedComparisons(evaluation.Results)
	if err != nil {
		return nil, 0, err
	}
	if evaluation.Exceeded < 0 {
		return nil, 0, fmt.Errorf("negative exceeded comparison count %d", evaluation.Exceeded)
	}
	result := make([]comparisonV1, 0, len(comparisons))
	actualExceeded := 0
	seen := make(map[metrics.Name]struct{}, len(comparisons))
	for _, comparison := range comparisons {
		if _, duplicate := seen[comparison.Metric]; duplicate {
			return nil, 0, fmt.Errorf("duplicate comparison metric %q", comparison.Metric)
		}
		seen[comparison.Metric] = struct{}{}
		delta, ok := checkedSubtract(comparison.Actual, comparison.Limit)
		if !ok {
			return nil, 0, fmt.Errorf("comparison delta overflows for %q", comparison.Metric)
		}
		if comparison.Delta != delta {
			return nil, 0, fmt.Errorf(
				"comparison delta for %q is %d, want %d",
				comparison.Metric,
				comparison.Delta,
				delta,
			)
		}
		passed := comparison.Actual <= comparison.Limit
		if comparison.Passed != passed {
			return nil, 0, fmt.Errorf("comparison pass state for %q is inconsistent", comparison.Metric)
		}
		if !passed {
			actualExceeded++
		}
		result = append(result, comparisonV1{
			Metric:   comparison.Metric,
			Observed: comparison.Actual,
			Limit:    comparison.Limit,
			Delta:    comparison.Delta,
			Passed:   comparison.Passed,
		})
	}
	if evaluation.Exceeded != actualExceeded {
		return nil, 0, fmt.Errorf(
			"exceeded comparison count is %d, want %d",
			evaluation.Exceeded,
			actualExceeded,
		)
	}
	switch evaluation.Status {
	case budget.StatusPassed:
		if actualExceeded != 0 {
			return nil, 0, errors.New("passed evaluation contains exceeded comparisons")
		}
	case budget.StatusFailed:
		if actualExceeded == 0 {
			return nil, 0, errors.New("failed evaluation contains no exceeded comparison")
		}
	case budget.StatusIncomplete:
		if evaluation.Reliability == analysis.ReliabilityExact || !evaluation.FailOnPartial {
			return nil, 0, errors.New("incomplete evaluation requires rejected non-exact evidence")
		}
	}

	return result, int64(actualExceeded), nil
}

func checkedSubtract(left, right int64) (int64, bool) {
	if right > 0 && left < math.MinInt64+right {
		return 0, false
	}
	if right < 0 && left > math.MaxInt64+right {
		return 0, false
	}
	return left - right, true
}

func portablePolicyKind(kind policy.Kind) string {
	if kind == policy.KindNone {
		return "overrides"
	}
	return string(kind)
}

func portableSceneIdentity(result application.InspectResult) (string, error) {
	for _, candidate := range []string{result.Scene.Display, result.Scene.Original, result.Scene.Canonical} {
		if portable, ok := portableOptionalPath(result.Project.Directory, candidate); ok {
			return portable, nil
		}
	}

	return "", errors.New("scene has no portable in-project identity")
}

func portableOptionalPath(projectRoot, candidate string) (string, bool) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return "", false
	}
	normalized := strings.ReplaceAll(candidate, `\`, "/")
	if strings.HasPrefix(strings.ToLower(normalized), "res:/") {
		relative := strings.TrimPrefix(normalized[5:], "/")
		return resourcePath(relative), relative != ""
	}
	if portable, ok := portableProjectPath(projectRoot, candidate); ok {
		return portable, true
	}
	if absolutePath(normalized) || strings.Contains(normalized, "://") {
		return "", false
	}
	relative := strings.TrimPrefix(path.Clean(normalized), "./")
	if relative == "" || relative == "." || relative == ".." || strings.HasPrefix(relative, "../") {
		return "", false
	}
	return resourcePath(relative), true
}

func portableOptionalResource(projectRoot, candidate string) (string, bool) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return "", false
	}
	normalized := strings.ReplaceAll(candidate, `\`, "/")
	if strings.HasPrefix(strings.ToLower(normalized), "res:/") {
		relative := strings.TrimPrefix(normalized[5:], "/")
		return resourcePath(relative), relative != ""
	}
	if portable, ok := portableProjectPath(projectRoot, candidate); ok {
		return portable, true
	}
	if absolutePath(normalized) {
		return "", false
	}
	return normalized, true
}

func portableProjectPath(projectRoot, candidate string) (string, bool) {
	root := strings.TrimSuffix(strings.ReplaceAll(filepath.Clean(projectRoot), `\`, "/"), "/")
	value := strings.ReplaceAll(filepath.Clean(candidate), `\`, "/")
	if root == "" || root == "." || value == "" || value == "." {
		return "", false
	}
	compareRoot := root
	compareValue := value
	if windowsAbsolute(root) || windowsAbsolute(value) {
		compareRoot = strings.ToLower(root)
		compareValue = strings.ToLower(value)
	}
	if compareValue == compareRoot {
		return "", false
	}
	prefix := compareRoot + "/"
	if !strings.HasPrefix(compareValue, prefix) {
		return "", false
	}
	relative := strings.TrimPrefix(value[len(root):], "/")
	if relative == "" || relative == "." || relative == ".." || strings.HasPrefix(relative, "../") {
		return "", false
	}
	return resourcePath(relative), true
}

func resourcePath(relative string) string {
	return "res://" + strings.TrimPrefix(path.Clean(strings.ReplaceAll(relative, `\`, "/")), "/")
}

func absolutePath(value string) bool {
	return strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || windowsAbsolute(value)
}

func windowsAbsolute(value string) bool {
	return len(value) >= 3 && value[1] == ':' && value[2] == '/'
}

func encodeDocumentV1(document documentV1) (string, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(document); err != nil {
		return "", fmt.Errorf("encode report JSON: %w", err)
	}

	return output.String(), nil
}
