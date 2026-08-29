package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stfulldev/deadweight.gdt/internal/analysis"
	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
)

type unresolvedGroup struct {
	path           string
	reason         string
	classification analysis.TargetClassification
	occurrences    int64
}

type inheritedGroup struct {
	path        string
	reason      string
	occurrences int64
}

func projectUnresolved(items []analysis.UnresolvedInstance) []unresolvedGroup {
	groups := make(map[string]unresolvedGroup, len(items))
	for _, item := range items {
		path := unresolvedDisplayPath(item)
		reason := displayClassification(item.Classification, string(item.ResolutionReason))
		key := path + "\x00" + reason + "\x00" + string(item.Classification)
		group := groups[key]
		group.path = path
		group.reason = reason
		group.classification = item.Classification
		group.occurrences += item.Occurrences
		groups[key] = group
	}

	result := make([]unresolvedGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].path != result[right].path {
			return result[left].path < result[right].path
		}
		if result[left].reason != result[right].reason {
			return result[left].reason < result[right].reason
		}

		return result[left].classification < result[right].classification
	})

	return result
}

func projectInherited(items []analysis.InheritedTarget) []inheritedGroup {
	groups := make(map[string]inheritedGroup, len(items))
	for _, item := range items {
		path := firstNonEmpty(
			item.TargetDisplay,
			item.BaseDisplay,
			item.BaseRawTarget,
			item.TargetOriginal,
			item.DeclaringDisplay,
		)
		if path == "" {
			path = "<unknown>"
		}
		reason := "inherited scene"
		key := path + "\x00" + reason
		group := groups[key]
		group.path = path
		group.reason = reason
		group.occurrences += item.Occurrences
		groups[key] = group
	}

	result := make([]inheritedGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].path != result[right].path {
			return result[left].path < result[right].path
		}

		return result[left].reason < result[right].reason
	})

	return result
}

func sortedDiagnostics(items []diagnostic.Diagnostic) []diagnostic.Diagnostic {
	result := append([]diagnostic.Diagnostic(nil), items...)
	sort.Slice(result, func(left, right int) bool {
		leftItem := result[left]
		rightItem := result[right]
		if severityRank(leftItem.Severity) != severityRank(rightItem.Severity) {
			return severityRank(leftItem.Severity) < severityRank(rightItem.Severity)
		}
		if leftItem.Code != rightItem.Code {
			return leftItem.Code < rightItem.Code
		}
		if leftItem.File != rightItem.File {
			return leftItem.File < rightItem.File
		}
		if leftItem.Line != rightItem.Line {
			return leftItem.Line < rightItem.Line
		}
		if leftItem.Resource != rightItem.Resource {
			return leftItem.Resource < rightItem.Resource
		}
		if leftItem.Column != rightItem.Column {
			return leftItem.Column < rightItem.Column
		}
		if leftItem.Message != rightItem.Message {
			return leftItem.Message < rightItem.Message
		}

		return leftItem.Occurrences < rightItem.Occurrences
	})

	return result
}

func writeEvidence(output *strings.Builder, result analysis.RecursiveResult, style styler) {
	unresolved := projectUnresolved(result.Summary.Unresolved)
	if len(unresolved) > 0 {
		output.WriteString("\nUnresolved scene instances\n")
		for _, group := range unresolved {
			fmt.Fprintf(output, "  %-34s ×%-4s %s\n", group.path, formatInteger(group.occurrences), group.reason)
		}
	}

	inherited := projectInherited(result.Summary.InheritedTargets)
	if len(inherited) > 0 {
		output.WriteString("\nInherited scene instances\n")
		for _, group := range inherited {
			fmt.Fprintf(output, "  %-34s ×%-4s %s\n", group.path, formatInteger(group.occurrences), group.reason)
		}
	}

	diagnostics := sortedDiagnostics(result.Diagnostics)
	if len(diagnostics) > 0 {
		output.WriteString("\nDiagnostics\n")
		for _, item := range diagnostics {
			severity := style.status(strings.ToUpper(string(item.Severity)))
			location := diagnosticLocation(item)
			fmt.Fprintf(output, "  %s %s", severity, item.Code)
			if location != "" {
				fmt.Fprintf(output, " %s", location)
			}
			fmt.Fprintf(output, ": %s", item.Message)
			if item.Occurrences > 1 {
				fmt.Fprintf(output, " ×%s", formatInteger(item.Occurrences))
			}
			output.WriteByte('\n')
		}
	}
}

func unresolvedDisplayPath(item analysis.UnresolvedInstance) string {
	path := firstNonEmpty(
		item.TargetDisplay,
		item.RawTarget,
		item.TargetOriginal,
		item.TargetCanonical,
		item.DeclaringDisplay,
		item.DeclaringScene,
	)
	if path == "" {
		return "<unknown>"
	}

	return path
}

func displayClassification(classification analysis.TargetClassification, reason string) string {
	switch classification {
	case analysis.TargetImportedScene:
		return "imported PackedScene"
	case analysis.TargetInheritedScene:
		return "inherited scene"
	case analysis.TargetMissingExternalResource:
		return "missing scene resource"
	case analysis.TargetUnresolvedPath:
		if reason != "" {
			return "unresolved path (" + strings.ReplaceAll(reason, "_", " ") + ")"
		}
		return "unresolved path"
	case analysis.TargetUnsupportedScene:
		return "unsupported scene"
	case analysis.TargetSubResource:
		return "sub-resource instance"
	case analysis.TargetPlaceholder:
		return "placeholder instance"
	case analysis.TargetUnavailableScene:
		return "unavailable scene"
	default:
		value := strings.ReplaceAll(string(classification), "_", " ")
		if value == "" {
			return "unresolved scene"
		}
		return value
	}
}

func diagnosticLocation(item diagnostic.Diagnostic) string {
	location := item.File
	if item.Line > 0 {
		location += fmt.Sprintf(":%d", item.Line)
		if item.Column > 0 {
			location += fmt.Sprintf(":%d", item.Column)
		}
	}
	if item.Resource != "" {
		if location != "" {
			location += " "
		}
		location += "[" + item.Resource + "]"
	}

	return location
}

func severityRank(severity diagnostic.Severity) int {
	if severity == diagnostic.SeverityError {
		return 0
	}

	return 1
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
