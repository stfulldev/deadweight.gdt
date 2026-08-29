## Why

Issues #8–#11 retain exact local, recursive, graph, and cache evidence, but the returned root summary intentionally leaves the two unique-union metric fields at zero. Issue [#12](https://github.com/stfulldev/deadweight.gdt/issues/12) and Draft PR [#37](https://github.com/stfulldev/deadweight.gdt/pull/37) complete the frozen MVP metric contract by publishing all eight non-negative `int64` values together in canonical order.

## What Changes

- Finalize root analysis metrics after graph validation and recursive expansion, preserving the six existing occurrence/maximum values.
- Set `external_resources` from the unique resolved/unresolved resource identity union and `scene_dependencies` from the checked graph dependency count.
- Keep local and cached one-occurrence summaries' unique fields zero so repeated application cannot multiply them.
- Validate final metric non-negativity and preserve the fixed `nodes` through `scene_dependencies` order already shared by metrics, budgets, presets, config, and later reports.
- Add table-driven and integration coverage for literal type/default-shadow rules, repeated/diamond uniqueness, unresolved/imported exclusions, resolved/inherited root special cases, and the §20.7 aggregation example.
- Non-goals: inherited effective-tree aggregation (#14), reliability/diagnostic synthesis (#13), CLI/report/config/budget behavior, Godot execution, or roadmap metrics.

## Capabilities

### New Capabilities

- `scene-metric-aggregation`: Final root publication and deterministic ordering of the eight frozen MVP scene metrics.

### Modified Capabilities

None.

## Impact

- Affected production area: `internal/analysis` finalization plus existing `internal/metrics` value/order contracts.
- `RecursiveAnalyzer.Analyze` and its `Expand` projection will return populated unique metric fields for successful root analysis; cached child summaries remain unchanged.
- No new dependencies, persistence, process, network, Godot runtime, configuration keys, report format, or budget semantics are introduced.
