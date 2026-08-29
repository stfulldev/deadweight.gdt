## Why

Issue #13 and Draft PR #38 must turn the analyzer's retained unresolved, inheritance, parent, resource, and cache evidence into one honest root result. Without an explicit completeness and reliability contract, downstream report and budget layers could present known partial metrics as exact.

## What Changes

- Add stable `complete`/`partial` analysis status and `exact`/`lower_bound`/`approximate` reliability values to successful recursive results.
- Publish checked coverage for parsed scene files, resolved and unresolved scene-instance occurrences, and inherited-scene occurrences.
- Convert every retained partial-analysis reason into validated warning diagnostics, including a stable unsupported-parent diagnostic, with deterministic grouping, occurrence counts, ownership, and ordering.
- Treat unresolved declared resources as partial while leaving successfully resolved ordinary non-scene resources complete without deep parsing.
- Preserve the frozen fatal/partial boundary: supported malformed text scenes, cycles, invalid roots, and arithmetic failure remain fatal and publish no usable result.
- Keep inherited effective-tree aggregation, CLI/report rendering, configuration, budgets, and `fail_on_partial` behavior out of this change; those remain in #14 and later tracker issues.

## Capabilities

### New Capabilities

- `analysis-completeness`: Defines successful root status, reliability precedence, coverage publication, grouped diagnostics, and the fatal/partial matrix.

### Modified Capabilities

None.

## Impact

- Affects `internal/analysis` result finalization and `internal/diagnostic` warning taxonomy.
- Extends the analyzer-domain result without changing parser, resolver, graph traversal, metrics, CLI, report, config, budget, preset, dependency, Godot, process, or network behavior.
- Covers the MVP acceptance criteria that imported or missing scenes never produce false `COMPLETE`, inheritance is visibly approximate, unresolved coverage is exact, normal resolved resource references remain complete, and fatal failures return no result.
- Adds no runtime dependency and remains compatible with the existing Go 1.24 build and standalone binary contract.
