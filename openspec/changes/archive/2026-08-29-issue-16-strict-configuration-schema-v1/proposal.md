## Why

Issue #16 and Draft PR #40 must add the frozen `.deadweight.gdt.json` version-1 contract before profile resolution and application wiring can safely consume project policy. Strict decoding, deterministic discovery, and a checked-in canonical schema prevent misspelled fields or malformed limits from silently changing CI guardrails.

## What Changes

- Discover configuration in frozen priority order: explicit `--config`, then `<project>/.deadweight.gdt.json`, otherwise no configuration; distinguish fatal missing explicit paths from normal missing implicit paths.
- Add an owned version-1 configuration model for selectors, `fail_on_partial`, all eight optional budgets, and custom-profile declarations.
- Decode exactly one JSON document with unknown fields rejected at every owned object boundary, preserving absent values separately from valid zero values.
- Validate version, selector exclusion, stable IDs, non-negative integers, renderer/quality enums, and profile field rules with actionable typed `SB2003` errors.
- Add `schema/deadweight.gdt.schema.json` using JSON Schema Draft 2020-12 and keep it behaviorally aligned with the Go model.
- Expose decoded declarations separately from dynamic profile-graph resolution; parent existence, built-in collisions, cycles, depth limits, inheritance merge, and effective-policy construction remain in issue #17.
- Preserve the standalone Go binary contract with no runtime schema engine, network, Node.js, OpenSpec, or Godot dependency.

## Capabilities

### New Capabilities

- `strict-configuration-v1`: Defines config discovery, strict JSON decoding, owned version-1 declarations, static semantic validation, canonical JSON Schema constraints, and typed failure behavior.

### Modified Capabilities

None.

## Impact

- Adds a focused `internal/config` package and a version-controlled `schema/deadweight.gdt.schema.json` artifact.
- Reuses the existing eight-field `budget.Limits`, built-in preset identifiers, and `SB2003` diagnostic taxonomy without changing preset values, project-root discovery, CLI flags, or budget evaluation.
- Covers MVP acceptance for missing explicit/implicit config behavior, full/minimal valid documents, strict field/type/version validation, selector exclusion, all eight optional limits including zero, ID/enumeration validation, and Go/schema parity.
- Does not implement custom-profile inheritance/merging (#17), partial policy evaluation (#18), command orchestration (#20), or report rendering (#21).
