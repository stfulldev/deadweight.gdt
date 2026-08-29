## Why

Recursive expansion now produces correct occurrence summaries, but its dependency evidence is only a path set and cyclic references still stop at a temporary safety error. Issue #10 and linked Draft PR #35 add the explainable graph and deterministic fatal cycle contract required before final metrics, completeness, budgets, and CLI orchestration can safely consume recursive analysis.

Tracking: [GitHub issue #10](https://github.com/stfulldev/deadweight.gdt/issues/10), [Draft PR #35](https://github.com/stfulldev/deadweight.gdt/pull/35).

## What Changes

- Build deterministic canonical graph nodes plus compacted instance and inheritance edges that preserve raw targets, resource IDs, resolution state, source identities, and checked occurrence multiplicity.
- Keep unresolved scene targets as graph evidence without inventing resolved nodes or including them in the dependency count.
- Traverse resolved edges with explicit unvisited, visiting, and visited DFS states, reconstructing deterministic self-cycle and multi-scene cycle chains.
- Replace the temporary recursive-reference safety error with a typed fatal cycle error exposing `SB2002` and no usable recursive result.
- Compute the unique resolved reachable scene-dependency count from graph nodes, excluding the analyzed root while including transitive instance and inheritance targets.
- Preserve inherited-scene metric deferral: the graph follows resolved inheritance for topology and dependency identity only, while issue #14 remains responsible for inherited effective-tree aggregation.

### Goals

- Produce one owned, deterministically ordered dependency graph for the analyzed root closure.
- Make cycle failures explainable and safe for the existing fatal-error/exit-code boundary.
- Provide the authoritative graph-backed dependency value that issue #12 can publish with the final eight metrics.

### Non-goals

- Publish the final whole-analysis result, run budget verdicts, or wire placeholder inspect/check commands; issues #12, #13, and #20 own those orchestration layers.
- Add cache statistics or parsed-coverage observability beyond the invocation reuse already needed for traversal; issue #11 owns cache hardening.
- Expand inherited base metrics or merge override semantics; issue #14 owns inherited-scene approximation.
- Render a separate graph output format or add concurrency, persistent caches, Godot, or network dependencies.

### Compatibility

This is an additive internal graph capability plus a deliberate replacement of the temporary internal recursive-reference error with stable `SB2002` cycle evidence. Acyclic recursive metrics and unresolved occurrence semantics remain compatible.

### Affected MVP Acceptance Criteria

- Criterion 9: `scene_dependencies` is derived from unique reachable resolved scene identities.
- Criterion 10: `A → B → C → A` produces exit-compatible `SB2002` evidence with the full cycle chain.
- Section 29.3: chain, diamond, repeated, self-cycle, multi-node cycle, unresolved-edge, and inheritance-edge graph tests become executable.
- Section 32 graph slice: cycle validation completes before final metric and budget publication.

## Capabilities

### New Capabilities

- `scene-dependency-graph`: Defines graph identities, compacted instance/inheritance edges, unresolved edge evidence, deterministic DFS cycle diagnostics, and unique resolved dependency counting.

### Modified Capabilities

- `recursive-scene-expansion`: Makes the graph authoritative for dependency identities, includes resolved inheritance topology without inherited metric expansion, and replaces the temporary recursive-reference error with graph-backed `SB2002` failure.

## Impact

- Affected areas: `internal/analysis` graph models/traversal, recursive analyzer result/error contracts, deterministic tests, and compatibility with the existing diagnostic/CLI fatal-error mapping.
- No project discovery, filesystem-open policy, report formatting, configuration, budget, or command behavior is added in this slice.
- Later final-metrics and completeness layers receive exact graph/dependency/cycle evidence without reopening TSCN node classification.
