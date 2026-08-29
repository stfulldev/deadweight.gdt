## Why

Issue [#20](https://github.com/stfulldev/deadweight.gdt/issues/20) made every application flow executable, but `inspect` and `check` still emit compact transitional summaries rather than the frozen MVP report contract. Issue [#21](https://github.com/stfulldev/deadweight.gdt/issues/21) and Draft PR [#44](https://github.com/stfulldev/deadweight.gdt/pull/44) complete the user-facing CLI by rendering reliable, deterministic evidence and preserving exact process outcomes for CI.

## What Changes

- Add a pure report layer for complete, lower-bound, and inherited-approximate `inspect` output with frozen metric groups, coverage, unresolved evidence, diagnostics, reliability markers, and warnings.
- Add passing, failing, and incomplete `check` tables with effective preset/profile metadata, observed comparisons, exceedance deltas, verdict summaries, and built-in preset disclaimers.
- Move preset list/show formatting into the same deterministic report boundary while preserving product order and stable locale-independent integers.
- Sort every user-visible metric, diagnostic, and unresolved group by the frozen rules without mutating application results.
- Apply ANSI only when output is a terminal and neither `--no-color` nor `NO_COLOR` disables it; retain textual `PASS`, `FAIL`, `WARNING`, `FAILED`, and `INCOMPLETE` signals in every mode.
- Render fatal errors deterministically on stderr without stack traces and preserve centralized priority `2 > 3 > 1 > 0`, including report-first exit `1` and `3` outcomes.
- Add byte-for-byte golden coverage for complete, lower-bound, approximate, passing, failing, incomplete, preset, and error states.
- Preserve existing command syntax, application orchestration, analysis semantics, policy precedence, and preset contents; no breaking compatibility change is introduced.

Goals are exact, readable, screenshot-ready output and CI-safe outcome mapping. Non-goals are JSON/HTML formats, per-metric confidence, new diagnostics, README release prose, broad fixture-matrix completion, or changes to frozen domain calculations.

## Capabilities

### New Capabilities

- `deterministic-console-reports`: Defines report contents, ordering, reliability notation, color suppression, error presentation, and outcome-visible console behavior for every MVP command.

### Modified Capabilities

None.

## Impact

- Adds `internal/report` as a pure projection over existing application/domain results and replaces transitional rendering in `internal/cli`.
- Extends CLI composition with injectable environment/terminal detection while keeping `cmd/deadweight.gdt/main.go` thin.
- Adds report golden files and focused color/error/outcome tests without requiring Godot or new Go modules.
- Advances MVP acceptance criteria 19–22 and Milestone 6; the broader seven-project fixture matrix and full acceptance traceability remain scoped to Issue #22.
