## Why

Issue [#18](https://github.com/stfulldev/deadweight.gdt/issues/18), tracked by Draft PR [#42](https://github.com/stfulldev/deadweight.gdt/pull/42), must turn analyzed metrics and the effective policy from #17 into a deterministic check outcome. The existing comparison helper does not yet publish summary verdicts, reliability context, or the configured/CLI partial-analysis decision needed by later application and report layers.

## What Changes

- Preserve optional limits, explicit zero, inclusive upper bounds, canonical metric order, and actual/limit/delta/pass fields for all configured comparisons.
- Add an owned evaluation result with `PASSED`, `FAILED`, or `INCOMPLETE`, reliability, ordered comparisons, exceeded count, and effective partial-rejection policy.
- Keep known lower-bound and approximate comparisons visible and unchanged; reliability remains explicit so issue #21 can render `Observed`, `+`, `~`, and `FAIL*` accurately.
- Resolve a domain-level CLI partial override (`inherit`, `fail`, or `allow`) over config `fail_on_partial`, preserving the default false behavior.
- Give partial rejection higher final priority than budget failure while retaining every observed comparison and exceeded count.
- Validate metrics, limits, reliability, and override values; invalid input returns no partial evaluation and remains a fatal application concern for issue #20.
- Cover MVP acceptance criteria 19–20 and the budget portion of 16 with table-driven tests for all metrics, boundary values, reliability classes, override precedence, and verdict priority.
- Non-goals: Cobra flag mutual exclusion, process exit mapping, policy resolution, scene I/O, diagnostic aggregation, or console formatting.

## Capabilities

### New Capabilities

- `budget-evaluation`: Evaluates optional metric limits and reliability-aware partial policy into deterministic comparisons and summary verdicts.

### Modified Capabilities

None.

## Impact

- Affected code: additive validation/evaluation contracts in `internal/budget` plus focused tests, consuming `internal/metrics` and `internal/analysis` status taxonomy.
- Public behavior: provides the domain outcome later used by `check`; existing parser, analyzer, policy resolver, CLI placeholder, and report output remain unchanged in this issue.
- Compatibility: the existing `budget.Check` comparison helper and `Result` fields remain available; no external dependency, filesystem, Godot, network, or runtime schema requirement is added.
- Delivery: English OpenSpec artifacts and implementation stay scoped to issue #18 and Draft PR #42; full repository gates and strict OpenSpec validation are required before archive and merge.
