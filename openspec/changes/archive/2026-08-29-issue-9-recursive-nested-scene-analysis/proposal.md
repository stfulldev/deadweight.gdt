## Why

The analyzer can summarize one parsed scene but cannot yet follow its nested-scene mounts, so repeated instances, child metrics, unresolved roots, and closure-wide resource evidence remain unavailable. Issue #9 and linked Draft PR #34 add the recursive expansion boundary required before the dependency graph, cache hardening, final metric aggregation, and completeness layers can be completed.

Tracking: [GitHub issue #9](https://github.com/stfulldev/deadweight.gdt/issues/9), [Draft PR #34](https://github.com/stfulldev/deadweight.gdt/pull/34).

## What Changes

- Classify every local instance mount from its reference kind, matching external-resource declaration, secure path-resolution result, literal resource type, and target extension.
- Recursively load and summarize existing canonical in-project `.tscn` format-3 targets while preserving fatal nested parse failures.
- Reuse one document and one expanded summary per canonical scene identity within an analysis invocation while applying the cached summary separately to every occurrence.
- Aggregate resolved child occurrence metrics without adding an extra child root, and retain exactly one known root for every unresolved instance occurrence.
- Apply child depth at each mount with `mountDepth + child.tree_depth - 1` and preserve partial depth evidence when a mount depth is unknown.
- Preserve deterministic unique sets for resolved scene dependencies and resolved/unresolved external-resource identities rather than multiplying them with occurrences.
- Route repeated-instance additions and multiplications through checked non-negative `int64` arithmetic so recursive expansion cannot wrap.
- Return structured unresolved target and coverage evidence for later diagnostic/reliability policy instead of silently skipping unsupported branches.

### Goals

- Produce a deterministic expanded summary for one acyclic nested-scene closure and one root occurrence.
- Make repeated and diamond-shaped dependencies correct by separating cached per-scene summaries from per-occurrence application.
- Establish a lossless handoff for the later graph, final metrics, and completeness issues.

### Non-goals

- Build the explainable dependency graph, publish `SB2002` cycle chains, or compute the final graph-backed `scene_dependencies` metric; issue #10 owns those contracts.
- Finish cache instrumentation and parsed-file coverage policy beyond the invocation-local memoization required for correct repeated expansion; issue #11 hardens those behaviors.
- Publish the final eight-metric analysis result or CLI/report output; issue #12 owns the complete metric surface.
- Assign user-visible unresolved diagnostics, final `COMPLETE`/`PARTIAL` status, or reliability; issue #13 owns that policy.
- Expand inherited roots or merge overrides; issue #14 handles inherited-scene approximation.

### Compatibility

This is an additive internal capability. Existing local summaries, secure path resolution, parser behavior, commands, configuration, and reports remain compatible. The recursive layer consumes their existing contracts and adds no Godot, network, Node.js, OpenSpec, or persistent-cache runtime dependency.

### Affected MVP Acceptance Criteria

- Criterion 5: a repeated nested scene is loaded/expanded once and applied with correct multiplicity.
- Criterion 6: resolved instance roots are not double-counted, while unresolved roots contribute exactly one known node.
- Criterion 7: child tree depth follows the frozen mount formula.
- Criterion 8: recursive occurrence contributions and unique evidence are covered as inputs to the final eight-metric slice.
- Section 29.3: chain, diamond, repeated, unresolved, unique-union, cache-reuse, and overflow tests become executable at the recursive-summary boundary.
- Section 32, Step 5: resolved TSCN recursion and repeated-summary reuse are introduced before graph/cycle presentation.

## Capabilities

### New Capabilities

- `recursive-scene-expansion`: Defines target classification, recursive `.tscn` expansion, occurrence aggregation, unresolved-root accounting, mount-depth composition, invocation-local summary reuse, unique closure evidence, and checked arithmetic.

### Modified Capabilities

None. Existing capabilities define local summaries and secure paths but not recursive scene expansion.

## Impact

- Affected areas: `internal/analysis`, its boundary with `project.Resolver` and `tscn.Parse`, and in-memory/temp-project integration tests.
- The implementation will add recursive expanded-summary and unresolved-evidence models plus injectable scene loading, without wiring placeholder CLI commands.
- Later issues can add graph/cycle diagnostics, cache observability, final metric publication, and completeness policy without reclassifying raw node headers.
