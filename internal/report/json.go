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
	ID    metrics.Name `json:"id"`
	Value int64        `json:"value"`
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
		metricsDocument = append(metricsDocument, metricV1{ID: name, Value: value})
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
