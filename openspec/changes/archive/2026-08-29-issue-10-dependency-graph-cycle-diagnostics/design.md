## Context

See `proposal.md` for motivation and the two delta specs for behavior. Issue #9 introduced a resolver-and-loader driven recursive analyzer with per-invocation document, local-summary, completed-summary, and in-progress maps. It returns an `ExpandedSummary`, classifies unavailable nested scenes honestly, defers inherited metrics, and currently uses `RecursiveReferenceError` only to prevent recursion overflow. Its `Dependencies` slice is a unique path union, not an explainable graph, and resolved inherited bases are not traversed.

The existing diagnostic and CLI boundary already maps any typed coded error, including `SB2002`, to fatal exit code 2. This slice must provide the domain error and return no usable analysis result; placeholder command composition, final metrics, completeness, and budget verdicts remain later work.

## Goals / Non-Goals

**Goals:**

- Build and validate the full reachable instance/inheritance topology before occurrence aggregation publishes a successful result.
- Share one invocation's parsed documents, local summaries, resolution results, and loader failures between graph discovery and recursive expansion.
- Preserve the issue #9 metric behavior while making graph nodes authoritative for dependency identities and scene-dependency count.
- Keep graph/result/error values owned, deterministic, and safe under checked `int64` arithmetic.

**Non-Goals:**

- Merge inherited base metrics or override stubs into effective-tree metrics.
- Publish `scene_dependencies` into the final eight-field metric result before issue #12.
- Add graph rendering, command wiring, budget execution, persistent cache state, cache statistics, goroutines, or filesystem policy to `analysis`.

## Decisions

### 1. Add an analysis result that pairs the summary with an owned graph

Add `RecursiveResult` with `Summary ExpandedSummary` and `Graph DependencyGraph`, plus an `Analyze(root)` method on `RecursiveAnalyzer`. Preserve the current `Expand(root)` method as a compatibility wrapper that calls `Analyze` and returns only its summary. Both methods return zero values on fatal failure.

`DependencyGraph` contains the root canonical/display identity, sorted `GraphNode` and `GraphEdge` slices, and a checked `SceneDependencies` value. A node stores canonical and display paths. An edge stores canonical/display declaring identities; optional canonical/display target identities; raw target; resource ID; `instance` or `inheritance` kind; resolved state; target classification; resolution reason; and occurrences.

Alternative considered: attach graph fields directly to `ExpandedSummary`. Rejected because occurrence metrics and topology have different aggregation semantics, and later orchestration needs to consume or render the graph without treating it as another metric-evidence slice.

Alternative considered: replace `Expand` outright. Rejected because the additive wrapper keeps existing issue #9 callers and tests compatible while issue #12 migrates to the complete result.

### 2. Use one two-phase invocation: discover/validate graph, then expand metrics

One invocation state is allocated by `Analyze`. Phase one loads local summaries, resolves declarations, builds edges, follows every loadable instance and inheritance target, and validates cycles. Phase two runs the existing occurrence expansion only after graph validation succeeds. Documents, local summaries, non-parse load failures, and resource-resolution maps are shared across both phases, so graph discovery does not double-load or re-resolve scenes before expansion.

This ordering guarantees that inherited-only cycles and cycles hidden beyond a branch that occurrence expansion defers are fatal before a usable result exists. It also lets occurrence expansion retain the simpler completed-summary cache on a graph proven acyclic. The existing in-progress guard remains only as an internal invariant fallback; it no longer defines public cycle behavior.

Alternative considered: build graph opportunistically inside the current recursive aggregation. Rejected because an inherited scene deliberately stops exact metric expansion, so topology beyond its base would be missed, and cycle selection would depend on occurrence/group traversal details.

Alternative considered: build a graph after successful expansion. Rejected because current expansion stops on a temporary recursive-reference error before it can reconstruct a complete cycle and does not follow inheritance bases.

### 3. Reuse one target-classification boundary for graph edges and expansion

Refactor resource resolution into a canonical-scene keyed cache. A shared target classifier consumes a local mount or inherited-root reference plus its declaring-scene resource map and returns either one exact `.tscn` candidate or one unresolved classification. Secure resolution always happens before extension classification.

Graph discovery attempts every resolved `.tscn` candidate. A successful local summary produces a resolved edge and node. A non-parse loader failure changes that edge to unavailable/unresolved with no target node. A typed parse failure remains fatal. Imported/binary, unsupported, placeholder, `SubResource`, missing-ID, and secure-resolution failures become unresolved edges immediately. Expansion consumes the same cached resolution/load outcome so its existing unresolved evidence stays consistent with the graph.

For an inherited root, the declaring inherited scene produces a separate inheritance edge using the root reference. A resolved supported base is visited for topology, transitive resources, cache reuse, and cycle validation, but its summary is never applied as an exact child metric contribution in this slice. Explicit local mounts in the inherited document are still graph edges even though issue #14 owns their effective inherited aggregation semantics.

Alternative considered: derive graph edges from `ExpandedSummary.Unresolved` and `Dependencies`. Rejected because those collections lose edge kind, declaring-to-target relationships, inherited-base topology, and compacted occurrence identities.

### 4. Compact edges through a comparable key and checked accumulation

The internal graph builder owns node and edge maps. The edge key contains every stable edge identity field except occurrences: from/to canonical and display identities, raw target, resource ID, kind, resolved state, classification, and resolution reason. Adding an equivalent local edge increments occurrences with `checkedAdd`; a different unresolved reason or edge kind remains separate. Finalization materializes sorted owned slices.

Resolved edge compaction affects explanation only. Recursive metric aggregation continues to group and apply by child identity plus mount depth, so graph compaction cannot erase occurrence or depth semantics. Graph node insertion and dependency counting also use checked increments rather than unchecked integer conversion from map length.

Alternative considered: include mount path and source position in the graph edge key. Rejected because it would prevent the required repeated-edge compaction; occurrence-specific source evidence remains in recursive unresolved/inherited records while the graph retains the declaring scene and resource target identity.

### 5. Detect cycles during deterministic DFS graph discovery

Maintain canonical-path keyed `unvisited`, `visiting`, and `visited` states plus an ordered stack of resolved paths and a stack-index map. For each scene, gather and sort all candidate outgoing edges before traversal by resolved target canonical path, edge kind, raw target, resource ID, classification, and reason. A resolved edge to a visiting target slices the stack at that target and appends the target again, producing a complete deterministic cycle.

`CycleError` owns parallel canonical and display string slices, implements `diagnostic.CodedError` with `SB2002`, and provides a code-free diagnostic message containing the display chain. The error itself is sufficient evidence even though the graph/result return values are zero. Self-reference follows the same stack rule and yields two entries.

Alternative considered: retain the first repeated canonical identity without a stack. Rejected because it cannot reconstruct the complete cycle or distinguish the relevant suffix of a longer traversal path.

Alternative considered: detect cycles with topological sorting. Rejected because it signals a cycle but does not naturally reconstruct the required root-to-root display chain, and unresolved edges are intentionally outside the resolved node graph.

### 6. Derive dependencies and graph-wide resources at the successful boundary

After cycle-free discovery, every graph node except the root contributes once to `DependencyGraph.SceneDependencies` and to the summary's sorted `Dependencies` slice. Missing/unavailable/imported targets have no node and therefore do not count. Instance and inheritance paths share the same canonical identity set, so repeated and diamond reachability remain unique.

Resource identities collected while resolving every successfully parsed graph node are unioned into the root expanded summary. This includes declarations in inherited bases reached only for topology, while inherited occurrence/type metrics remain deferred. `ExpandedSummary.Metrics.SceneDependencies` stays zero until issue #12 publishes the complete eight-field metric collection.

Alternative considered: use the existing recursive child dependency union as the graph count. Rejected because it omits inherited bases and cannot prove that every counted identity corresponds to a graph node successfully loaded in the validated closure.

### 7. Rely on the existing CLI fatal mapping without command changes

`CycleError.DiagnosticCode()` returns `diagnostic.CodeSceneDependencyCycle`; `DiagnosticMessage()` returns deterministic code-free text. The existing CLI executor already renders coded errors once and returns exit code 2, so analysis and focused compatibility tests are sufficient in this slice. Because `Analyze` and `Expand` return zero results on the error, later budget orchestration cannot accidentally consume a truncated verdict.

Alternative considered: wire inspect/check commands now solely to demonstrate exit code 2. Rejected because the commands do not yet own complete result/config/report composition, and issue #20 is the designated CLI integration slice.

## Risks / Trade-offs

- [Graph discovery parses inherited bases earlier than metric aggregation] → Treat every supported resolved `.tscn` parse failure consistently as fatal, cache the result, and keep inherited metric application explicitly absent.
- [Two phases could duplicate resolver or loader work] → Cache resource resolutions, local summaries, successful documents, and load failures in the shared invocation state and test exact call counts.
- [Compacted edges hide individual mount locations] → Keep per-mount unresolved/inherited records in `ExpandedSummary`; graph edges retain declaring/target identities and checked occurrences for explanation.
- [Multiple reachable cycles can exist] → Sort outgoing candidates before DFS and test the selected exact chain across repeated runs.
- [The compatibility `Expand` wrapper could diverge from `Analyze`] → Implement it only as a projection of `Analyze`, with common root validation and zero-on-fatal behavior.
- [Graph-wide resource union broadens evidence through inherited bases] → Include only declarations from successfully parsed graph nodes; do not infer resources behind unresolved targets.

## Migration Plan

1. Add graph/result/error models and the checked deterministic graph builder without changing existing callers.
2. Refactor invocation resource/load reuse and implement graph discovery for instance and inheritance targets.
3. Run graph validation before current occurrence expansion, replace the public temporary recursion error with `CycleError`, and project graph dependency/resource identities onto the root summary.
4. Commit production-only changes after build/test/vet gates, then add focused graph/cycle/compatibility tests in a separate commit.
5. Run all quality gates, sync both capability specs, archive the change, and merge the linked PR.

Rollback is additive: revert the test, feature, and archive commits. No configuration, persistent cache, public CLI output format, or user data requires migration.
