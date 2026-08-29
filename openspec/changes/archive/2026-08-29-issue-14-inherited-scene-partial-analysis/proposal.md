## Why

Issue #14 and Draft PR #39 must replace the temporary one-known-root fallback for inherited scenes with the frozen MVP approximate aggregation contract. The graph and completeness layers now expose enough topology and honesty evidence to reuse a resolved base safely without pretending to implement Godot override semantics.

## What Changes

- Expand resolved inherited roots by applying the base scene summary exactly once without adding a scene-instance occurrence or a duplicate root node.
- Add explicit local typed nodes and nested mounts while retaining typeless stubs and `[editable ...]` signals as unsupported override evidence.
- Preserve inheritance evidence through cached and repeated summaries, always produce `PARTIAL approximate`, and emit grouped `SB1003` warnings with the base target.
- Keep missing, unreadable, imported, binary, `SubResource`, UID/user, and otherwise unsupported bases successful approximate results with one known inherited root and explicit evidence.
- Preserve fatal supported-text parse failures and inheritance cycles, unique dependency/resource counts, checked arithmetic, and deterministic ownership.
- Keep full override merge semantics, CLI/report rendering, config, budget, preset, and runtime Godot behavior out of scope.

## Capabilities

### New Capabilities

- `inherited-scene-analysis`: Defines inherited-root aggregation, unsupported override evidence, approximate reliability, base fallback, coverage, and diagnostics.

### Modified Capabilities

- `recursive-scene-expansion`: Replaces deferred inherited targets with limited approximate child expansion and base-summary application.

## Impact

- Affects `internal/analysis` local evidence, recursive expansion, completeness coverage validation, and inheritance diagnostics.
- Does not change parser syntax, secure resolution, graph cycle detection, CLI/report/config/budget packages, dependencies, or the standalone Go runtime contract.
- Covers MVP acceptance for resolved bases, missing/imported bases, override stubs, editable signals, repeated inheritance, inherited cycles, and exact root/scene-instance counting boundaries.
