## Why

Issue [#17](https://github.com/stfulldev/deadweight.gdt/issues/17), tracked by Draft PR [#41](https://github.com/stfulldev/deadweight.gdt/pull/41), must turn the declarations delivered by strict configuration v1 and the frozen built-in catalog into one deterministic effective check policy. Without this layer, custom profiles, inheritance, project-wide overrides, and CLI budgets cannot be applied consistently or diagnosed safely.

## What Changes

- Add deterministic selection between CLI and config preset/profile selectors while preserving the distinction between built-ins and custom profiles.
- Resolve single-inheritance custom-profile graphs with built-in or custom parents, full-chain cycle diagnostics, a maximum inheritance depth of 32, missing-parent errors, and built-in-ID collision rejection.
- Merge metadata and all eight optional budgets field-by-field from ancestor to descendant, then apply top-level config budgets and repeated CLI budget overrides in frozen priority order.
- Supply the documented custom metadata defaults for a root custom profile and retain inherited or explicitly overridden metadata.
- Parse ordered CLI `metric=limit` overrides strictly, allowing duplicate metrics with the last value winning, and reject an effective check policy that contains no budget.
- Return owned, deterministic effective-policy values and actionable `SB2003` configuration failures without coupling policy resolution to Cobra, scene I/O, reporting, or budget evaluation.
- Cover MVP acceptance criteria 17 and the profile/config portion of 18 with table-driven merge, selection, inheritance, collision, and failure tests.
- Non-goals: command wiring, partial-analysis policy, metric evaluation, report rendering, config JSON decoding, and changes to frozen built-in values.

## Capabilities

### New Capabilities

- `effective-policy-resolution`: Resolves built-in/custom selectors, profile inheritance, metadata, project budgets, and ordered CLI overrides into one validated effective check policy.

### Modified Capabilities

None.

## Impact

- Affected code: a focused `internal/policy` resolver and tests, consuming the existing `internal/config`, `internal/preset`, and `internal/budget` contracts.
- Public behavior: establishes the deterministic policy contract later used by `check`; no existing CLI flow changes in this issue.
- Compatibility: additive Go-internal capability with no schema, preset-data, parser, analyzer, filesystem, or external dependency changes.
- Delivery: English OpenSpec artifacts and implementation remain scoped to issue #17 and Draft PR #41; full project quality gates and strict OpenSpec validation are required before archive and merge.
