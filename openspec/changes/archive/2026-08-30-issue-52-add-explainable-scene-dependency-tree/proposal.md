## Why

The analyzer already retains an authoritative resolved and unresolved scene-dependency graph, but users can only see its aggregate dependency count. Issue #52 and linked Draft PR #61 add a bounded, portable explanation of how the analyzed root reaches each dependency so developers can inspect repeated, diamond, inherited, and unsupported branches without reparsing scenes or guessing from totals.

## What Changes

- Add a focused `tree <scene>` command that reuses the existing project discovery, secure scene resolution, and recursive analysis flow.
- Define a deterministic rooted projection of the authoritative dependency graph with one expanded occurrence of every graph edge and stable back-references for already-expanded nodes.
- Render resolved instance and inheritance edges with checked multiplicity, and keep unresolved, imported, missing, placeholder, sub-resource, unavailable, and unsupported targets visible with their stable classifications.
- Add portable human-readable tree output and a new compatible version-one JSON `tree` document kind without exposing canonical checkout paths or OS-specific separators.
- Preserve existing cycle failures, diagnostic codes, root metrics, `inspect`/`check` behavior, and exit-code meanings.
- Add complete, repeated, diamond, partial, inherited, cycle, portability, schema, and golden coverage.
- Keep Graphviz/DOT, HTML, interactive visualization, project-wide graphs, UID resolution, imported-scene expansion, and full inherited-scene merging out of scope.

## Capabilities

### New Capabilities

- `scene-dependency-tree`: Defines the rooted dependency-tree projection, deterministic expansion/back-reference rules, edge evidence, portability, and text/JSON semantics.

### Modified Capabilities

- `application-command-flows`: Adds the `tree <scene>` command flow and keeps format selection independent of analysis.
- `deterministic-console-reports`: Adds deterministic, color-independent dependency-tree text presentation.
- `versioned-json-reports`: Adds a compatible schema-version-one `tree` document kind with portable nodes, edges, back-references, and diagnostics.

## Impact

- Affects CLI composition and injected application interfaces, while reusing the existing inspect analysis request/result boundary.
- Adds report-domain tree projection and presentation models derived only from `analysis.DependencyGraph` plus existing status, reliability, coverage, and diagnostics.
- Extends `schema/deadweight.gdt.report-v1.schema.json` compatibly with a `tree` discriminator payload; existing inspect, check, and error meanings remain unchanged.
- Adds documentation and golden fixtures without introducing runtime dependencies, network access, Godot execution, persistent caches, or another parser pass.
- Satisfies the MVP 0.2 dependency-tree criteria in tracker #57 while preserving the frozen standalone static-analysis and honest-partial-evidence boundaries.
