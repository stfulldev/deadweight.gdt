## Why

The existing built-in preset catalog is not yet fully protected by the validation, copy-isolation, and exact-value tests required to freeze it for v0.1.0. Finalizing those contracts now makes the experimental guardrails reproducible without presenting them as certified performance targets.

Tracking: [GitHub issue #15](https://github.com/stfulldev/deadweight.gdt/issues/15), [Draft PR #29](https://github.com/stfulldev/deadweight.gdt/pull/29).

## What Changes

- Freeze the `mobile`, `steam-deck`, and `desktop` metadata and all eight budget limits to the catalog in MVP specification section 22.2.
- Require built-ins to load from version-controlled embedded JSON and reject unsupported renderer or quality identifiers, negative limits, incomplete budgets, and inconsistent preset identity or lifecycle metadata.
- Preserve deterministic product order and make values returned by both catalog retrieval and preset lookup independent from mutable package or catalog state.
- Return an actionable unknown-preset error that identifies the requested ID and lists the available IDs in product order.
- Add exact snapshot-style coverage for every metadata field and budget limit, lifecycle labels, ordering, validation failures, and copy isolation.
- Preserve the patch-release rule that a necessary frozen-catalog correction must be documented in `CHANGELOG.md`.

### Goals

- Make the built-in catalog an exact, deterministic, immutable-by-contract v0.1.0 interface.
- Fail clearly if shipped embedded data violates the frozen catalog schema.
- Keep all user-facing claims explicitly heuristic and experimental.

### Non-goals

- Benchmarking hardware, certifying frame-rate outcomes, or claiming official Valve endorsement.
- Adding new built-in presets, custom profiles, override resolution, or the `check` execution path.
- Changing the eight MVP metric definitions or calibrating new limit values.

### Compatibility

This change is non-breaking: it preserves the three public preset IDs, their product order, their documented metadata and limits, and existing CLI list/show behavior. Stricter loading affects only invalid embedded data shipped by the project; callers receive safer independent values and a more actionable lookup failure.

### Affected MVP Acceptance Criteria

- Criterion 14: built-in presets remain available to `presets` and `presets show`.
- Criterion 15: all three presets exactly match the frozen section 22.2 catalog.
- Criterion 22: preset ordering and related output remain deterministic.
- Section 29.4: tests pin exact built-in values and metadata.
- Section 33.3: wording continues to describe guardrails rather than performance guarantees or certification.

## Capabilities

### New Capabilities

- `built-in-heuristic-presets`: Defines the frozen embedded preset catalog, its validation and ordering rules, copy-isolated retrieval and lookup behavior, and its non-certification contract.

### Modified Capabilities

None. This repository has no existing OpenSpec capability for built-in presets.

## Impact

- Affected areas: `internal/preset`, its embedded JSON data, preset unit tests, and CLI callers that retrieve or look up built-ins.
- The implementation will audit and tighten the existing preset code rather than introduce a parallel catalog.
- No new runtime dependency, network access, Godot dependency, or CLI compatibility break is introduced.
