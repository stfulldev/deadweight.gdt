## Why

Recursive graph discovery and expansion already reuse some invocation state, but issue [#11](https://github.com/stfulldev/deadweight.gdt/issues/11) requires the cache boundary to own physical open/parse effects, expose parsed-file coverage from unique cached documents, and prove every repeated-occurrence arithmetic path fails safely. Completing that contract in Draft PR [#36](https://github.com/stfulldev/deadweight.gdt/pull/36) keeps analysis linear in unique parsed scenes before final metric and completeness layers consume its result.

## What Changes

- Separate injectable scene-open and parse effects so tests can independently prove one physical read and one parse per canonical scene in chain, diamond, and repeated graphs.
- Centralize invocation-scoped parsed-document, local-summary, resource-resolution, loader-failure, and one-occurrence expanded-summary caches keyed by canonical absolute scene identity.
- Return parsed-scene-file coverage from successful parse-cache cardinality without multiplying it by instance occurrences.
- Preserve occurrence multiplication, unique dependency/resource unions, diagnostics, inherited/unresolved evidence, reliability inputs, and maximum depth when cached summaries are reused.
- Audit all non-negative recursive and graph count arithmetic so overflow or invalid negative operands return typed fatal `SB2004` errors without panic, wrapping, clamping, or usable partial results.
- Keep the existing no-Godot, no-network runtime and deterministic result ordering.
- Non-goals: persistent/on-disk invalidation, concurrent parsing, final unique metric publication, reliability verdict construction, inherited effective-tree aggregation, CLI/report/budget changes, or roadmap 0.2+ cache behavior.
- Acceptance evidence covers independently instrumented physical reads/parses, diamond and ×100 summary reuse, parse-cache-cardinality coverage, owned cached values, depth/unique-set preservation, invocation isolation, and overflow matrices.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `recursive-scene-expansion`: Strengthen invocation cache, injectable I/O/parse instrumentation, parsed-file coverage, one-occurrence summary reuse, and checked-arithmetic requirements.

## Impact

- Affected production area: `internal/analysis` recursive construction, invocation state/cache ownership, result models, and checked arithmetic.
- The internal analyzer construction API will accept independent open and parse hooks; repository call sites and tests will migrate together, with no shipped CLI compatibility promise affected.
- Existing scene graph, recursive metric, diagnostic-code, resolver-containment, and deterministic ownership contracts remain compatible.
- No new dependency, process, Godot runtime, persistent storage, network access, config key, report schema, or budget behavior is introduced.
