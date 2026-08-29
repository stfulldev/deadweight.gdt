## Context

See `proposal.md` for motivation and the new `scene-metric-aggregation` contract. The local summary already computes literal nodes, depth, mounts, meshes, lights, and shadows; recursive expansion already applies one-occurrence summaries with checked arithmetic; the graph owns checked unique dependency count; and the root summary retains a deterministic union of resource identities. Its two unique metric fields are deliberately reset to zero inside every cached summary, so #12 needs a final root-only publication boundary rather than another traversal.

## Goals / Non-Goals

**Goals:**

- Finalize all eight values from already validated evidence without changing traversal or multiplicity.
- Keep cached/local child summaries safe for reuse with zero unique metric fields.
- Use checked cardinality and validate the completed value object before return.
- Prove the frozen type/default/order/example contracts end to end.

**Non-Goals:**

- Reimplementing local classification, graph discovery, caching, or arithmetic.
- Inherited effective-tree aggregation, completeness/reliability, diagnostic rendering, budgets, config, reports, or CLI composition.
- Adding metric registry entries beyond the frozen eight.

## Decisions

### 1. Add a pure final-metric publication function in analysis

Create a small `finalizeMetrics(summary, graph)` function that copies the six recursive values, converts the final resource-identity slice length through `checkedCardinality`, takes `SceneDependencies` from the graph, and validates the completed `metrics.Values`. It returns a new value and never mutates cached input.

`Analyze` will merge graph-wide resources and dependency identities first, call this finalizer, assign the completed value to the root summary, and only then construct `RecursiveResult`. `Expand` remains a projection of `Analyze`, so successful callers receive the same final eight metrics.

Alternatives considered:

- Populate unique fields in every cached child summary: rejected because repeated application could multiply root-global values.
- Recount dependencies from the summary slice: rejected because the checked graph count is authoritative.
- Add a second traversal: rejected because all required evidence is already retained and deterministic.

### 2. Preserve existing package ownership and metric order

`internal/analysis` owns aggregation; `internal/metrics` continues to own names, labels, `Values`, validation, and canonical order. No new registry or map is introduced. Tests will enumerate `metrics.OrderedNames()` and `Values.Get` to prove the finalized values align with the existing order.

Alternatives considered:

- Return metrics as a map: rejected because it weakens fixed order and compile-time field ownership.
- Duplicate ordering in analysis: rejected because it creates two sources of truth.

### 3. Treat unique evidence and occurrence evidence differently

The finalizer counts the already de-duplicated `ResourceIdentity` slice once. Resolved identities use canonical paths; unresolved identities retain declaring-scene/resource-ID/raw-path tuples. Scene dependencies reuse graph nodes excluding root. Six other fields remain exactly as recursive aggregation produced them; tree depth remains a maximum.

Imported/unavailable scene contents are not parsed and therefore cannot invent internal nodes or resources. Resolved inheritance participates in dependency/resource topology but inherited effective-tree occurrence values remain deferred to #14.

### 4. Keep failure behavior transactional

If resource cardinality cannot fit signed `int64` or the completed `metrics.Values` violates non-negativity, finalization returns an error and `Analyze` returns a zero `RecursiveResult`. Upstream cycle, parse, and arithmetic failures still stop before this boundary. No partially finalized summary is cached or returned.

## Risks / Trade-offs

- [Existing tests expect unique fields to stay zero] → update only root `Analyze`/`Expand` expectations; retain direct local/cache assertions that child fields remain zero.
- [Inherited graph resources appear before exact inherited metrics] → this is frozen topology/unique evidence behavior from #10; keep occurrence aggregation deferred and test the distinction explicitly.
- [The §20.7 fixture can be mis-modeled by local mount nodes] → construct exact one-occurrence Building and Lamp scenes and assert both documented equations plus shared-resource uniqueness.
- [Finalization looks deceptively simple] → retain table-driven tests across all eight fields, unsupported literal types, repeated/diamond graphs, and deterministic ordered enumeration.

## Migration Plan

1. Add the pure finalizer and call it after root identity merging.
2. Update existing root expectations and add focused aggregation fixtures.
3. Commit production and tests separately, run all gates, sync/archive OpenSpec, and merge Draft PR #37 after CI.

Rollback is a focused PR revert; no stored data or external schema changes exist.
