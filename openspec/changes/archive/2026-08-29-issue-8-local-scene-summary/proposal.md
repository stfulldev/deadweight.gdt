## Why

The parser exposes the TSCN fields needed by the analyzer, but the repository has no domain layer that converts one parsed document into deterministic local node counts, depths, resource evidence, and nested-scene mount points. Issue #8 and linked Draft PR #33 add that non-recursive boundary so later graph traversal can expand child scenes without reinterpreting node paths or double-counting instance roots.

Tracking: [GitHub issue #8](https://github.com/stfulldev/deadweight.gdt/issues/8), [Draft PR #33](https://github.com/stfulldev/deadweight.gdt/pull/33).

## What Changes

- Build a local scene summary from one parsed TSCN document without reading files, resolving dependencies, or recursively expanding nested scenes.
- Classify ordinary nodes, nested-scene candidate mounts, inherited roots, instance placeholders, and override stubs so each later aggregation path has an explicit representation.
- Count exact local ordinary nodes and literal `MeshInstance3D`, supported 3D light, and enabled-shadow occurrences while preserving every declared external-resource key.
- Derive root-depth-1 local tree depths and instance mount depths from supported serialized parent paths, including multi-segment paths.
- Report unsupported, missing, or ambiguous parent relationships as typed partial evidence instead of guessing a depth; keep known local counts available.
- Ensure an instance header is represented as a mount rather than an ordinary node contribution, preventing its resolved child root from being counted twice.

### Goals

- Provide one deterministic, testable local-summary contract for the later dependency graph and aggregation layers.
- Preserve exact local counts and mount positions for supported ordinary TSCN scene trees.
- Make unsupported parent and inheritance/override semantics explicit in the result.

### Non-goals

- Resolve `ExtResource.path`, open or parse child scenes, build a dependency graph, detect cycles, cache summaries, or aggregate resolved child metrics.
- Implement inherited-scene merging or apply override stubs to a base scene.
- Wire `inspect` or `check`, finalize whole-analysis coverage/reliability policy, or change report output.
- Infer custom-script base classes, runtime-created nodes, or imported/binary scene contents.

### Compatibility

This change is additive and internal. Existing parser, project resolver, CLI, metric, preset, and budget behavior remains unchanged. The summary consumes existing `tscn.Document` values and introduces no runtime dependency on Godot, the network, Node.js, or OpenSpec.

### Affected MVP Acceptance Criteria

- Criterion 6: local instance nodes are represented so a resolved child root is not double-counted.
- Criterion 7: root depth is 1 and local/mount depths follow the serialized-parent formula required by later expansion.
- Criterion 8: exact local contributions are covered for nodes, tree depth, scene instances, mesh instances, lights, and shadow lights; unique and recursive aggregation remains for later issues.
- Section 29.3: root counting and mount-depth behavior gain focused unit coverage before graph aggregation.
- Section 32, Step 4: one scene receives a non-recursive local summary of ordinary nodes, types, depth, external-resource keys, and instance mounts.

## Capabilities

### New Capabilities

- `local-scene-summary`: Defines non-recursive node classification, exact local metric contributions, external-resource evidence, supported parent-path depth calculation, instance mount points, and partial evidence for unsupported local tree semantics.

### Modified Capabilities

None. Existing capabilities do not define analysis of a parsed TSCN document.

## Impact

- Affected areas: a new focused internal analysis package, existing `tscn.Document` and `metrics.Values` consumers, and table-driven unit tests built from in-memory documents.
- Later graph code will consume the summary's mount and external-resource records rather than repeat node classification and depth logic.
- No existing Go API is removed, no persistent data or configuration changes, and no CLI output changes in this slice.
