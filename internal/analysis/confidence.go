package analysis

import (
	"fmt"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
)

// ConfidenceReason identifies stable static evidence that qualifies a metric.
type ConfidenceReason string

const (
	ConfidenceUnresolvedSceneInstance ConfidenceReason = "unresolved_scene_instance"
	ConfidenceImportedScene           ConfidenceReason = "imported_scene"
	ConfidenceUnsupportedScene        ConfidenceReason = "unsupported_scene"
	ConfidenceSubresourceScene        ConfidenceReason = "subresource_scene"
	ConfidencePlaceholderInstance     ConfidenceReason = "placeholder_instance"
	ConfidenceUnavailableScene        ConfidenceReason = "unavailable_scene"
	ConfidenceUnavailableResource     ConfidenceReason = "unavailable_resource"
	ConfidenceUnsupportedResourcePath ConfidenceReason = "unsupported_resource_path"
	ConfidenceInheritedScene          ConfidenceReason = "inherited_scene"
	ConfidenceUnsupportedParent       ConfidenceReason = "unsupported_parent"
)

var confidenceReasonOrder = [...]ConfidenceReason{
	ConfidenceUnresolvedSceneInstance,
	ConfidenceImportedScene,
	ConfidenceUnsupportedScene,
	ConfidenceSubresourceScene,
	ConfidencePlaceholderInstance,
	ConfidenceUnavailableScene,
	ConfidenceUnavailableResource,
	ConfidenceUnsupportedResourcePath,
	ConfidenceInheritedScene,
	ConfidenceUnsupportedParent,
}

// Valid reports whether reason is part of the public confidence taxonomy.
func (reason ConfidenceReason) Valid() bool {
	for _, candidate := range confidenceReasonOrder {
		if reason == candidate {
			return true
		}
	}

	return false
}

// Confidence qualifies one metric value with owned machine-readable evidence.
type Confidence struct {
	Reliability Reliability
	Reasons     []ConfidenceReason
}

// Validate checks reliability, reason presence, uniqueness, and canonical order.
func (confidence Confidence) Validate() error {
	if !confidence.Reliability.Valid() {
		return fmt.Errorf("invalid metric confidence reliability %q", confidence.Reliability)
	}
	if confidence.Reliability == ReliabilityExact {
		if len(confidence.Reasons) != 0 {
			return fmt.Errorf("exact metric confidence must not have reasons")
		}
		return nil
	}
	if len(confidence.Reasons) == 0 {
		return fmt.Errorf("non-exact metric confidence requires a reason")
	}
	previous := -1
	for _, reason := range confidence.Reasons {
		rank := confidenceReasonRank(reason)
		if rank < 0 {
			return fmt.Errorf("invalid metric confidence reason %q", reason)
		}
		if rank <= previous {
			return fmt.Errorf("metric confidence reasons are not unique and canonical")
		}
		previous = rank
	}

	return nil
}

// MetricConfidence contains exactly one confidence value for every frozen metric.
type MetricConfidence struct {
	Nodes             Confidence
	TreeDepth         Confidence
	SceneInstances    Confidence
	MeshInstances     Confidence
	Lights            Confidence
	ShadowLights      Confidence
	ExternalResources Confidence
	SceneDependencies Confidence
}

// MetricConfidenceEntry pairs a frozen metric with its owned confidence value.
type MetricConfidenceEntry struct {
	Metric     metrics.Name
	Confidence Confidence
}

// ExactMetricConfidence constructs the complete exact eight-metric baseline.
func ExactMetricConfidence() MetricConfidence {
	exact := Confidence{Reliability: ReliabilityExact}
	return MetricConfidence{
		Nodes:             exact,
		TreeDepth:         exact,
		SceneInstances:    exact,
		MeshInstances:     exact,
		Lights:            exact,
		ShadowLights:      exact,
		ExternalResources: exact,
		SceneDependencies: exact,
	}
}

// UniformMetricConfidence constructs one classification for all frozen metrics.
func UniformMetricConfidence(
	reliability Reliability,
	reasons ...ConfidenceReason,
) (MetricConfidence, error) {
	result := ExactMetricConfidence()
	if reliability == ReliabilityExact && len(reasons) == 0 {
		return result, nil
	}
	if err := result.merge(allMetricNames(), reliability, reasons...); err != nil {
		return MetricConfidence{}, err
	}

	return result, nil
}

// With returns an owned copy with evidence merged into one frozen metric.
func (confidence MetricConfidence) With(
	name metrics.Name,
	reliability Reliability,
	reasons ...ConfidenceReason,
) (MetricConfidence, error) {
	result := cloneMetricConfidence(confidence)
	if err := result.merge([]metrics.Name{name}, reliability, reasons...); err != nil {
		return MetricConfidence{}, err
	}

	return result, nil
}

// Get returns an owned confidence value for one frozen metric.
func (confidence MetricConfidence) Get(name metrics.Name) (Confidence, bool) {
	var result Confidence
	switch name {
	case metrics.Nodes:
		result = confidence.Nodes
	case metrics.TreeDepth:
		result = confidence.TreeDepth
	case metrics.SceneInstances:
		result = confidence.SceneInstances
	case metrics.MeshInstances:
		result = confidence.MeshInstances
	case metrics.Lights:
		result = confidence.Lights
	case metrics.ShadowLights:
		result = confidence.ShadowLights
	case metrics.ExternalResources:
		result = confidence.ExternalResources
	case metrics.SceneDependencies:
		result = confidence.SceneDependencies
	default:
		return Confidence{}, false
	}
	result.Reasons = append([]ConfidenceReason(nil), result.Reasons...)

	return result, true
}

// Entries returns owned confidence entries in frozen metric order.
func (confidence MetricConfidence) Entries() []MetricConfidenceEntry {
	result := make([]MetricConfidenceEntry, 0, len(metrics.OrderedNames()))
	for _, name := range metrics.OrderedNames() {
		entry, _ := confidence.Get(name)
		result = append(result, MetricConfidenceEntry{Metric: name, Confidence: entry})
	}

	return result
}

// Validate checks every frozen metric confidence value.
func (confidence MetricConfidence) Validate() error {
	for _, entry := range confidence.Entries() {
		if err := entry.Confidence.Validate(); err != nil {
			return fmt.Errorf("metric %q: %w", entry.Metric, err)
		}
	}

	return nil
}

// Reliability returns the conservative summary across all frozen metrics.
func (confidence MetricConfidence) Reliability() Reliability {
	result := ReliabilityExact
	for _, entry := range confidence.Entries() {
		result = conservativeReliability(result, entry.Confidence.Reliability)
	}

	return result
}

func (confidence *MetricConfidence) merge(
	names []metrics.Name,
	reliability Reliability,
	reasons ...ConfidenceReason,
) error {
	if confidence == nil {
		return fmt.Errorf("metric confidence is nil")
	}
	addition, err := normalizedConfidence(reliability, reasons...)
	if err != nil {
		return err
	}
	for _, name := range names {
		current, ok := confidence.Get(name)
		if !ok {
			return fmt.Errorf("invalid confidence metric %q", name)
		}
		merged, mergeErr := mergeConfidence(current, addition)
		if mergeErr != nil {
			return mergeErr
		}
		confidence.set(name, merged)
	}

	return nil
}

func (confidence *MetricConfidence) set(name metrics.Name, value Confidence) {
	switch name {
	case metrics.Nodes:
		confidence.Nodes = value
	case metrics.TreeDepth:
		confidence.TreeDepth = value
	case metrics.SceneInstances:
		confidence.SceneInstances = value
	case metrics.MeshInstances:
		confidence.MeshInstances = value
	case metrics.Lights:
		confidence.Lights = value
	case metrics.ShadowLights:
		confidence.ShadowLights = value
	case metrics.ExternalResources:
		confidence.ExternalResources = value
	case metrics.SceneDependencies:
		confidence.SceneDependencies = value
	}
}

func normalizedConfidence(reliability Reliability, reasons ...ConfidenceReason) (Confidence, error) {
	if !reliability.Valid() {
		return Confidence{}, fmt.Errorf("invalid metric confidence reliability %q", reliability)
	}
	if reliability == ReliabilityExact {
		if len(reasons) != 0 {
			return Confidence{}, fmt.Errorf("exact metric confidence must not have reasons")
		}
		return Confidence{Reliability: ReliabilityExact}, nil
	}
	present := make(map[ConfidenceReason]struct{}, len(reasons))
	for _, reason := range reasons {
		if !reason.Valid() {
			return Confidence{}, fmt.Errorf("invalid metric confidence reason %q", reason)
		}
		present[reason] = struct{}{}
	}
	result := Confidence{Reliability: reliability}
	for _, reason := range confidenceReasonOrder {
		if _, ok := present[reason]; ok {
			result.Reasons = append(result.Reasons, reason)
		}
	}
	if err := result.Validate(); err != nil {
		return Confidence{}, err
	}

	return result, nil
}

func mergeConfidence(left, right Confidence) (Confidence, error) {
	if err := left.Validate(); err != nil {
		return Confidence{}, err
	}
	if err := right.Validate(); err != nil {
		return Confidence{}, err
	}
	reasons := append(append([]ConfidenceReason(nil), left.Reasons...), right.Reasons...)
	return normalizedConfidence(conservativeReliability(left.Reliability, right.Reliability), reasons...)
}

func confidenceReasonRank(reason ConfidenceReason) int {
	for index, candidate := range confidenceReasonOrder {
		if reason == candidate {
			return index
		}
	}

	return -1
}

func cloneMetricConfidence(confidence MetricConfidence) MetricConfidence {
	cloned := confidence
	for _, name := range metrics.OrderedNames() {
		value, _ := confidence.Get(name)
		cloned.set(name, value)
	}

	return cloned
}

func allMetricNames() []metrics.Name {
	return metrics.OrderedNames()
}
