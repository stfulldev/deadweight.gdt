package analysis

import (
	"fmt"
	"sort"

	"github.com/stfulldev/deadweight.gdt/internal/metrics"
	"github.com/stfulldev/deadweight.gdt/internal/project"
)

type uniqueEvidenceKey struct {
	metric metrics.Name

	canonical        string
	declaringScene   string
	resourceID       string
	rawTarget        string
	resolutionReason project.ResolutionReason
}

type uniqueReferrerKey struct {
	sceneCanonical string
	sceneDisplay   string
	resourceID     string
	rawTarget      string
	edgeKind       EdgeKind
}

func buildUniqueEvidence(graph DependencyGraph, cache *invocationCache) ([]UniqueEvidence, error) {
	items := make(map[uniqueEvidenceKey]UniqueEvidence)
	displays := graphDisplayPaths(graph)

	canonicals := make([]string, 0, len(cache.resourceResolutions))
	for canonical := range cache.resourceResolutions {
		canonicals = append(canonicals, canonical)
	}
	sort.Strings(canonicals)
	for _, canonical := range canonicals {
		resources := cache.resourceResolutions[canonical]
		ids := make([]string, 0, len(resources))
		for id := range resources {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			entry := resources[id]
			key := uniqueEvidenceKey{
				metric:           metrics.ExternalResources,
				declaringScene:   canonical,
				resourceID:       entry.resource.ID,
				rawTarget:        entry.resource.Path,
				resolutionReason: entry.resolution.Reason,
			}
			item := UniqueEvidence{
				Metric:           metrics.ExternalResources,
				DeclaringScene:   canonical,
				ResourceID:       entry.resource.ID,
				RawTarget:        entry.resource.Path,
				ResolutionReason: entry.resolution.Reason,
			}
			if entry.resolution.Resolved() {
				key = uniqueEvidenceKey{metric: metrics.ExternalResources, canonical: entry.resolution.Path.Canonical}
				item = UniqueEvidence{
					Metric:           metrics.ExternalResources,
					Canonical:        entry.resolution.Path.Canonical,
					Display:          entry.resolution.Path.Display,
					ResolutionReason: project.ResolutionResolved,
				}
			}
			referrer := UniqueReferrer{
				SceneCanonical: canonical,
				SceneDisplay:   displayPath(displays, canonical),
				ResourceID:     entry.resource.ID,
				RawTarget:      entry.resource.Path,
				Occurrences:    1,
			}
			if err := addUniqueEvidence(items, key, item, referrer); err != nil {
				return nil, err
			}
		}
	}

	for _, node := range graph.Nodes {
		if node.Canonical == graph.RootCanonical {
			continue
		}
		key := uniqueEvidenceKey{metric: metrics.SceneDependencies, canonical: node.Canonical}
		if _, exists := items[key]; !exists {
			items[key] = UniqueEvidence{
				Metric:    metrics.SceneDependencies,
				Canonical: node.Canonical,
				Display:   node.Display,
			}
		}
	}
	for _, edge := range graph.Edges {
		if !edge.Resolved || edge.ToCanonical == "" {
			continue
		}
		key := uniqueEvidenceKey{metric: metrics.SceneDependencies, canonical: edge.ToCanonical}
		item := items[key]
		referrer := UniqueReferrer{
			SceneCanonical: edge.FromCanonical,
			SceneDisplay:   edge.FromDisplay,
			ResourceID:     edge.ResourceID,
			RawTarget:      edge.RawTarget,
			EdgeKind:       edge.Kind,
			Occurrences:    edge.Occurrences,
		}
		if err := addReferrer(&item, referrer); err != nil {
			return nil, err
		}
		items[key] = item
	}

	result := make([]UniqueEvidence, 0, len(items))
	for _, item := range items {
		sortUniqueReferrers(item.Referrers)
		if err := item.Validate(); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	sort.Slice(result, func(left, right int) bool {
		return uniqueEvidenceLess(result[left], result[right])
	})

	return result, nil
}

func addUniqueEvidence(
	items map[uniqueEvidenceKey]UniqueEvidence,
	key uniqueEvidenceKey,
	item UniqueEvidence,
	referrer UniqueReferrer,
) error {
	current, exists := items[key]
	if !exists {
		current = item
	}
	if current.Display == "" && item.Display != "" {
		current.Display = item.Display
	}
	if err := addReferrer(&current, referrer); err != nil {
		return err
	}
	items[key] = current

	return nil
}

func addReferrer(item *UniqueEvidence, referrer UniqueReferrer) error {
	key := uniqueReferrerKey{
		sceneCanonical: referrer.SceneCanonical,
		sceneDisplay:   referrer.SceneDisplay,
		resourceID:     referrer.ResourceID,
		rawTarget:      referrer.RawTarget,
		edgeKind:       referrer.EdgeKind,
	}
	for index, existing := range item.Referrers {
		existingKey := uniqueReferrerKey{
			sceneCanonical: existing.SceneCanonical,
			sceneDisplay:   existing.SceneDisplay,
			resourceID:     existing.ResourceID,
			rawTarget:      existing.RawTarget,
			edgeKind:       existing.EdgeKind,
		}
		if key != existingKey {
			continue
		}
		occurrences, err := checkedAdd(existing.Occurrences, referrer.Occurrences)
		if err != nil {
			return err
		}
		item.Referrers[index].Occurrences = occurrences
		return nil
	}
	item.Referrers = append(item.Referrers, referrer)

	return nil
}

func sortUniqueReferrers(items []UniqueReferrer) {
	sort.Slice(items, func(left, right int) bool {
		first := items[left]
		second := items[right]
		if first.SceneCanonical != second.SceneCanonical {
			return first.SceneCanonical < second.SceneCanonical
		}
		if first.ResourceID != second.ResourceID {
			return first.ResourceID < second.ResourceID
		}
		if first.RawTarget != second.RawTarget {
			return first.RawTarget < second.RawTarget
		}
		if first.EdgeKind != second.EdgeKind {
			return first.EdgeKind < second.EdgeKind
		}
		if first.SceneDisplay != second.SceneDisplay {
			return first.SceneDisplay < second.SceneDisplay
		}

		return first.Occurrences < second.Occurrences
	})
}

func uniqueEvidenceLess(left, right UniqueEvidence) bool {
	if left.Metric != right.Metric {
		return uniqueMetricRank(left.Metric) < uniqueMetricRank(right.Metric)
	}
	if left.Canonical != right.Canonical {
		return left.Canonical < right.Canonical
	}
	if left.Display != right.Display {
		return left.Display < right.Display
	}
	if left.DeclaringScene != right.DeclaringScene {
		return left.DeclaringScene < right.DeclaringScene
	}
	if left.ResourceID != right.ResourceID {
		return left.ResourceID < right.ResourceID
	}
	if left.RawTarget != right.RawTarget {
		return left.RawTarget < right.RawTarget
	}

	return left.ResolutionReason < right.ResolutionReason
}

func uniqueMetricRank(name metrics.Name) int {
	if name == metrics.ExternalResources {
		return 0
	}
	if name == metrics.SceneDependencies {
		return 1
	}

	return 2
}

func validateUniqueEvidence(items []UniqueEvidence, values metrics.Values) error {
	counts := map[metrics.Name]uint64{
		metrics.ExternalResources: 0,
		metrics.SceneDependencies: 0,
	}
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return err
		}
		counts[item.Metric]++
	}
	for _, name := range []metrics.Name{metrics.ExternalResources, metrics.SceneDependencies} {
		count, err := checkedCardinality(counts[name])
		if err != nil {
			return err
		}
		want, _ := values.Get(name)
		if count != want {
			return fmt.Errorf("unique evidence for %q has %d items, want %d", name, count, want)
		}
	}

	return nil
}
