package analysis

import (
	"errors"

	"github.com/stfulldev/deadweight.gdt/internal/diagnostic"
	"github.com/stfulldev/deadweight.gdt/internal/project"
	"github.com/stfulldev/deadweight.gdt/internal/tscn"
)

type invocationCache struct {
	documents              map[string]*tscn.Document
	documentErrors         map[string]error
	localSummaries         map[string]LocalSummary
	localSummaryErrors     map[string]error
	resourceResolutions    map[string]map[string]resourceResolution
	expandedSceneSummaries map[string]ExpandedSummary
}

func newInvocationCache() *invocationCache {
	return &invocationCache{
		documents:              make(map[string]*tscn.Document),
		documentErrors:         make(map[string]error),
		localSummaries:         make(map[string]LocalSummary),
		localSummaryErrors:     make(map[string]error),
		resourceResolutions:    make(map[string]map[string]resourceResolution),
		expandedSceneSummaries: make(map[string]ExpandedSummary),
	}
}

func (cache *invocationCache) loadDocument(
	path project.ResolvedPath,
	opener SceneOpener,
	parser SceneParser,
) (*tscn.Document, error) {
	if document := cache.documents[path.Canonical]; document != nil {
		return document, nil
	}
	if loadErr, exists := cache.documentErrors[path.Canonical]; exists {
		return nil, loadErr
	}

	reader, err := opener(path)
	if err != nil {
		return nil, cache.storeDocumentError(path, err)
	}
	if reader == nil {
		return nil, cache.storeDocumentError(path, errors.New("scene opener returned a nil reader"))
	}

	document, parseErr := parser(reader, path.Display)
	closeErr := reader.Close()
	if parseErr != nil {
		return nil, cache.storeDocumentError(path, parseErr)
	}
	if closeErr != nil {
		return nil, cache.storeDocumentError(path, closeErr)
	}
	if document == nil {
		return nil, cache.storeDocumentError(path, errors.New("scene parser returned a nil document"))
	}

	cache.documents[path.Canonical] = document

	return document, nil
}

func (cache *invocationCache) storeDocumentError(path project.ResolvedPath, cause error) error {
	loadErr := cause
	if code, coded := diagnostic.CodeOf(cause); !coded || code != diagnostic.CodeInvalidTSCNRoot {
		loadErr = &sceneLoadError{path: path, cause: cause}
	}
	cache.documentErrors[path.Canonical] = loadErr

	return loadErr
}

func (cache *invocationCache) localSummary(canonical string) (LocalSummary, error, bool) {
	if summary, exists := cache.localSummaries[canonical]; exists {
		return cloneLocalSummary(summary), nil, true
	}
	if summaryErr, exists := cache.localSummaryErrors[canonical]; exists {
		return LocalSummary{}, summaryErr, true
	}

	return LocalSummary{}, nil, false
}

func (cache *invocationCache) storeLocalSummary(canonical string, summary LocalSummary) {
	cache.localSummaries[canonical] = cloneLocalSummary(summary)
}

func (cache *invocationCache) storeLocalSummaryError(canonical string, err error) {
	cache.localSummaryErrors[canonical] = err
}

func (cache *invocationCache) resources(canonical string) (map[string]resourceResolution, bool) {
	resources, exists := cache.resourceResolutions[canonical]
	if !exists {
		return nil, false
	}

	return cloneResourceResolutions(resources), true
}

func (cache *invocationCache) storeResources(
	canonical string,
	resources map[string]resourceResolution,
) {
	cache.resourceResolutions[canonical] = cloneResourceResolutions(resources)
}

func cloneResourceResolutions(resources map[string]resourceResolution) map[string]resourceResolution {
	cloned := make(map[string]resourceResolution, len(resources))
	for id, resolution := range resources {
		cloned[id] = resolution
	}

	return cloned
}

func (cache *invocationCache) expandedSummary(canonical string) (ExpandedSummary, bool) {
	summary, exists := cache.expandedSceneSummaries[canonical]
	if !exists {
		return ExpandedSummary{}, false
	}

	return cloneExpandedSummary(summary), true
}

func (cache *invocationCache) storeExpandedSummary(canonical string, summary ExpandedSummary) {
	cache.expandedSceneSummaries[canonical] = cloneExpandedSummary(summary)
}

func (cache *invocationCache) parsedSceneFiles() (int64, error) {
	return checkedCardinality(uint64(len(cache.documents)))
}
