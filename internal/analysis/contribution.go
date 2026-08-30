package analysis

import (
	"fmt"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

// ContributionKind identifies how one contribution enters the analyzed root.
type ContributionKind string

const (
	ContributionRoot       ContributionKind = "root"
	ContributionScene      ContributionKind = "scene"
	ContributionInherited  ContributionKind = "inherited"
	ContributionUnresolved ContributionKind = "unresolved"
)

// Valid reports whether kind is part of the contribution contract.
func (kind ContributionKind) Valid() bool {
	switch kind {
	case ContributionRoot, ContributionScene, ContributionInherited, ContributionUnresolved:
		return true
	default:
		return false
	}
}

// ContributionValues contains the five additive occurrence metrics.
type ContributionValues struct {
	Nodes          int64
	SceneInstances int64
	MeshInstances  int64
	Lights         int64
	ShadowLights   int64
}

// Get returns one additive value. Tree depth and unique-union metrics are not additive.
func (values ContributionValues) Get(name metrics.Name) (int64, bool) {
	switch name {
	case metrics.Nodes:
		return values.Nodes, true
	case metrics.SceneInstances:
		return values.SceneInstances, true
	case metrics.MeshInstances:
		return values.MeshInstances, true
	case metrics.Lights:
		return values.Lights, true
	case metrics.ShadowLights:
		return values.ShadowLights, true
	default:
		return 0, false
	}
}

func (values ContributionValues) validate() error {
	for _, name := range []metrics.Name{
		metrics.Nodes,
		metrics.SceneInstances,
		metrics.MeshInstances,
		metrics.Lights,
		metrics.ShadowLights,
	} {
		value, _ := values.Get(name)
		if value < 0 {
			return &metrics.ValueError{Name: name, Value: value}
		}
	}

	return nil
}

// SceneContribution is one direct, context-aware contribution row.
type SceneContribution struct {
	Kind ContributionKind

	SceneCanonical string
	SceneDisplay   string
	SceneOriginal  string

	DeclaringScene   string
	DeclaringDisplay string
	MountName        string
	MountPath        string
	MountDepth       OptionalDepth
	ResourceID       string
	RawTarget        string
	Classification   TargetClassification

	Occurrences      int64
	Values           ContributionValues
	DepthCandidate   OptionalDepth
	Reliability      Reliability
	MetricConfidence MetricConfidence
}

// Validate checks one contribution without depending on presentation format.
func (item SceneContribution) Validate() error {
	if !item.Kind.Valid() {
		return fmt.Errorf("invalid contribution kind %q", item.Kind)
	}
	if !item.Reliability.Valid() {
		return fmt.Errorf("invalid contribution reliability %q", item.Reliability)
	}
	if err := item.MetricConfidence.Validate(); err != nil {
		return err
	}
	if summary := item.MetricConfidence.Reliability(); item.Reliability != summary {
		return fmt.Errorf("contribution reliability %q does not match metric summary %q", item.Reliability, summary)
	}
	if item.Occurrences <= 0 {
		return fmt.Errorf("contribution occurrences must be positive, got %d", item.Occurrences)
	}
	if err := item.Values.validate(); err != nil {
		return err
	}
	if item.DepthCandidate.Known && item.DepthCandidate.Value <= 0 {
		return fmt.Errorf("known contribution depth must be positive, got %d", item.DepthCandidate.Value)
	}
	if item.MountDepth.Known && item.MountDepth.Value <= 0 {
		return fmt.Errorf("known contribution mount depth must be positive, got %d", item.MountDepth.Value)
	}

	switch item.Kind {
	case ContributionRoot:
		if item.SceneCanonical == "" || item.DeclaringScene != "" {
			return fmt.Errorf("root contribution requires one scene identity and no declaring scene")
		}
	case ContributionScene, ContributionInherited:
		if item.SceneCanonical == "" || item.DeclaringScene == "" {
			return fmt.Errorf("%s contribution requires scene and declaring identities", item.Kind)
		}
	case ContributionUnresolved:
		if item.DeclaringScene == "" {
			return fmt.Errorf("unresolved contribution requires a declaring scene")
		}
		if item.SceneCanonical == "" && item.SceneDisplay == "" && item.SceneOriginal == "" && item.RawTarget == "" && item.ResourceID == "" {
			return fmt.Errorf("unresolved contribution requires target evidence")
		}
		if !item.Classification.Valid() || item.Classification == TargetInheritedScene {
			return fmt.Errorf("invalid unresolved contribution classification %q", item.Classification)
		}
	}

	return nil
}

// UniqueReferrer identifies one parsed declaration or validated graph edge.
type UniqueReferrer struct {
	SceneCanonical string
	SceneDisplay   string
	ResourceID     string
	RawTarget      string
	EdgeKind       EdgeKind
	Occurrences    int64
}

// UniqueEvidence retains one unique-union identity without assigning ownership.
type UniqueEvidence struct {
	Metric metrics.Name

	Canonical        string
	Display          string
	DeclaringScene   string
	ResourceID       string
	RawTarget        string
	ResolutionReason project.ResolutionReason

	Referrers []UniqueReferrer
}

// Validate checks one unique-union item and its referrers.
func (item UniqueEvidence) Validate() error {
	if item.Metric != metrics.ExternalResources && item.Metric != metrics.SceneDependencies {
		return fmt.Errorf("invalid unique evidence metric %q", item.Metric)
	}
	if item.Canonical == "" && item.Display == "" && item.RawTarget == "" {
		return fmt.Errorf("unique evidence %q requires an identity", item.Metric)
	}
	if item.Metric == metrics.SceneDependencies && item.Canonical == "" {
		return fmt.Errorf("scene dependency evidence requires a canonical identity")
	}
	if item.Metric == metrics.ExternalResources && item.Canonical == "" &&
		(item.DeclaringScene == "" || item.ResourceID == "" || !item.ResolutionReason.Valid()) {
		return fmt.Errorf("unresolved resource evidence is incomplete")
	}
	if len(item.Referrers) == 0 {
		return fmt.Errorf("unique evidence %q requires at least one referrer", item.Metric)
	}
	for _, referrer := range item.Referrers {
		if referrer.SceneCanonical == "" || referrer.Occurrences <= 0 {
			return fmt.Errorf("invalid unique evidence referrer: %#v", referrer)
		}
		if referrer.EdgeKind != "" && referrer.EdgeKind != EdgeInstance && referrer.EdgeKind != EdgeInheritance {
			return fmt.Errorf("invalid unique evidence edge kind %q", referrer.EdgeKind)
		}
	}

	return nil
}

type contributionKey struct {
	kind ContributionKind

	sceneCanonical string
	sceneDisplay   string
	sceneOriginal  string

	declaringScene   string
	declaringDisplay string
	mountName        string
	mountPath        string
	mountDepth       int64
	mountDepthKnown  bool
	resourceID       string
	rawTarget        string
	classification   TargetClassification
}

func keyForContribution(item SceneContribution) contributionKey {
	return contributionKey{
		kind:             item.Kind,
		sceneCanonical:   item.SceneCanonical,
		sceneDisplay:     item.SceneDisplay,
		sceneOriginal:    item.SceneOriginal,
		declaringScene:   item.DeclaringScene,
		declaringDisplay: item.DeclaringDisplay,
		mountName:        item.MountName,
		mountPath:        item.MountPath,
		mountDepth:       item.MountDepth.Value,
		mountDepthKnown:  item.MountDepth.Known,
		resourceID:       item.ResourceID,
		rawTarget:        item.RawTarget,
		classification:   item.Classification,
	}
}

func compactContributions(items []SceneContribution) ([]SceneContribution, error) {
	compacted := make(map[contributionKey]SceneContribution, len(items))
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return nil, err
		}
		key := keyForContribution(item)
		current, exists := compacted[key]
		if !exists {
			item.MetricConfidence = cloneMetricConfidence(item.MetricConfidence)
			compacted[key] = item
			continue
		}
		var err error
		current.Occurrences, err = checkedAdd(current.Occurrences, item.Occurrences)
		if err != nil {
			return nil, err
		}
		if err := addContributionValues(&current.Values, item.Values); err != nil {
			return nil, err
		}
		if item.DepthCandidate.Known && (!current.DepthCandidate.Known || item.DepthCandidate.Value > current.DepthCandidate.Value) {
			current.DepthCandidate = item.DepthCandidate
		}
		current.Reliability = conservativeReliability(current.Reliability, item.Reliability)
		for _, name := range metrics.OrderedNames() {
			addition, _ := item.MetricConfidence.Get(name)
			if err := current.MetricConfidence.merge([]metrics.Name{name}, addition.Reliability, addition.Reasons...); err != nil {
				return nil, err
			}
		}
		current.Reliability = current.MetricConfidence.Reliability()
		compacted[key] = current
	}

	result := make([]SceneContribution, 0, len(compacted))
	for _, item := range compacted {
		result = append(result, item)
	}
	sort.Slice(result, func(left, right int) bool {
		return contributionLess(result[left], result[right])
	})

	return result, nil
}

func contributionLess(left, right SceneContribution) bool {
	leftKey := keyForContribution(left)
	rightKey := keyForContribution(right)
	if leftKey.sceneCanonical != rightKey.sceneCanonical {
		return leftKey.sceneCanonical < rightKey.sceneCanonical
	}
	if leftKey.sceneDisplay != rightKey.sceneDisplay {
		return leftKey.sceneDisplay < rightKey.sceneDisplay
	}
	if leftKey.declaringScene != rightKey.declaringScene {
		return leftKey.declaringScene < rightKey.declaringScene
	}
	if leftKey.mountPath != rightKey.mountPath {
		return leftKey.mountPath < rightKey.mountPath
	}
	if leftKey.kind != rightKey.kind {
		return leftKey.kind < rightKey.kind
	}
	if leftKey.resourceID != rightKey.resourceID {
		return leftKey.resourceID < rightKey.resourceID
	}
	if leftKey.rawTarget != rightKey.rawTarget {
		return leftKey.rawTarget < rightKey.rawTarget
	}
	if leftKey.classification != rightKey.classification {
		return leftKey.classification < rightKey.classification
	}
	if leftKey.mountDepthKnown != rightKey.mountDepthKnown {
		return !leftKey.mountDepthKnown
	}
	if leftKey.mountDepth != rightKey.mountDepth {
		return leftKey.mountDepth < rightKey.mountDepth
	}
	if leftKey.mountName != rightKey.mountName {
		return leftKey.mountName < rightKey.mountName
	}
	if leftKey.declaringDisplay != rightKey.declaringDisplay {
		return leftKey.declaringDisplay < rightKey.declaringDisplay
	}

	return leftKey.sceneOriginal < rightKey.sceneOriginal
}

func addContributionValues(target *ContributionValues, value ContributionValues) error {
	fields := []struct {
		target *int64
		value  int64
	}{
		{target: &target.Nodes, value: value.Nodes},
		{target: &target.SceneInstances, value: value.SceneInstances},
		{target: &target.MeshInstances, value: value.MeshInstances},
		{target: &target.Lights, value: value.Lights},
		{target: &target.ShadowLights, value: value.ShadowLights},
	}
	for _, field := range fields {
		next, err := checkedAdd(*field.target, field.value)
		if err != nil {
			return err
		}
		*field.target = next
	}

	return nil
}

func scaleContribution(item SceneContribution, multiplicity int64) (SceneContribution, error) {
	if multiplicity < 0 {
		return SceneContribution{}, &OverflowError{Operation: ArithmeticMultiply, Left: item.Occurrences, Right: multiplicity}
	}
	result := item
	result.MetricConfidence = cloneMetricConfidence(item.MetricConfidence)
	var err error
	result.Occurrences, err = checkedMultiply(item.Occurrences, multiplicity)
	if err != nil {
		return SceneContribution{}, err
	}
	fields := []struct {
		target *int64
		value  int64
	}{
		{target: &result.Values.Nodes, value: item.Values.Nodes},
		{target: &result.Values.SceneInstances, value: item.Values.SceneInstances},
		{target: &result.Values.MeshInstances, value: item.Values.MeshInstances},
		{target: &result.Values.Lights, value: item.Values.Lights},
		{target: &result.Values.ShadowLights, value: item.Values.ShadowLights},
	}
	for _, field := range fields {
		*field.target, err = checkedMultiply(field.value, multiplicity)
		if err != nil {
			return SceneContribution{}, err
		}
	}

	return result, nil
}

func conservativeReliability(left, right Reliability) Reliability {
	if left == ReliabilityApproximate || right == ReliabilityApproximate {
		return ReliabilityApproximate
	}
	if left == ReliabilityLowerBound || right == ReliabilityLowerBound {
		return ReliabilityLowerBound
	}

	return ReliabilityExact
}

func cloneContributions(items []SceneContribution) []SceneContribution {
	if items == nil {
		return nil
	}
	cloned := make([]SceneContribution, len(items))
	for index, item := range items {
		cloned[index] = item
		cloned[index].MetricConfidence = cloneMetricConfidence(item.MetricConfidence)
	}

	return cloned
}

func cloneUniqueEvidence(items []UniqueEvidence) []UniqueEvidence {
	cloned := make([]UniqueEvidence, len(items))
	for index, item := range items {
		cloned[index] = item
		cloned[index].Referrers = append([]UniqueReferrer(nil), item.Referrers...)
	}

	return cloned
}

func validateContributions(items []SceneContribution, values metrics.Values, depthPartial bool) error {
	if len(items) == 0 {
		return fmt.Errorf("successful analysis requires contribution evidence")
	}
	additive := ContributionValues{}
	maximum := OptionalDepth{}
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return err
		}
		if err := addContributionValues(&additive, item.Values); err != nil {
			return err
		}
		if item.DepthCandidate.Known && (!maximum.Known || item.DepthCandidate.Value > maximum.Value) {
			maximum = item.DepthCandidate
		}
	}
	checks := []struct {
		name metrics.Name
		got  int64
		want int64
	}{
		{name: metrics.Nodes, got: additive.Nodes, want: values.Nodes},
		{name: metrics.SceneInstances, got: additive.SceneInstances, want: values.SceneInstances},
		{name: metrics.MeshInstances, got: additive.MeshInstances, want: values.MeshInstances},
		{name: metrics.Lights, got: additive.Lights, want: values.Lights},
		{name: metrics.ShadowLights, got: additive.ShadowLights, want: values.ShadowLights},
	}
	for _, check := range checks {
		if check.got != check.want {
			return fmt.Errorf("contribution sum for %q is %d, want %d", check.name, check.got, check.want)
		}
	}
	if !depthPartial {
		if !maximum.Known || maximum.Value != values.TreeDepth {
			return fmt.Errorf("contribution depth maximum is %#v, want %d", maximum, values.TreeDepth)
		}
	}

	return nil
}

// ValidateContributionEvidence checks contribution reconciliation and unique-union cardinality.
func ValidateContributionEvidence(result RecursiveResult) error {
	if err := validateContributions(
		result.Summary.Contributions,
		result.Summary.Metrics,
		result.Summary.DepthPartial,
	); err != nil {
		return err
	}

	return validateUniqueEvidence(result.UniqueEvidence, result.Summary.Metrics)
}
