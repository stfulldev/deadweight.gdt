## Why

The parser, project resolver, recursive analyzer, policy resolver, and budget evaluator are implemented, but the `inspect` and `check` Cobra commands still stop at placeholders. Issue [#20](https://github.com/stfulldev/deadweight.gdt/issues/20) and Draft PR [#43](https://github.com/stfulldev/deadweight.gdt/pull/43) connect those existing domain services into the frozen MVP command flows so the CLI can execute useful end-to-end analysis without Godot.

## What Changes

- Add an application service that owns project/config discovery, scene resolution, recursive analysis, effective-policy resolution, partial-policy resolution, budget evaluation, and report-ready results.
- Replace the `inspect` and `check` placeholders with thin Cobra handlers that parse the frozen syntax and delegate to the application service.
- Route `presets` and `presets show <id>` through the same injected application boundary while preserving their ability to run outside a Godot project.
- Add global project/config/color switches and the full `check` selector, override, and partial-policy flag set, including exact argument validation and mutual-exclusion errors.
- Centralize command outcome-to-exit-code mapping while keeping `main.go` limited to build information, streams, arguments, and process exit.
- Add injected application and stream seams so command tests do not invoke Godot or depend on the host project layout.
- Keep deterministic final console formatting, ANSI/TTY policy, and golden report coverage as follow-up report work; this change produces complete report models and a minimal deterministic presentation boundary.
- Preserve all existing parser, analyzer, policy, budget, preset, and configuration behavior; no breaking compatibility change is introduced.

Goals are to make all four MVP command flows executable, preserve the CLI contract in §§25 and 27, and keep orchestration out of Cobra. Non-goals are changing analysis semantics, budget precedence, preset contents, configuration schema, or the frozen text report format planned for the next reporting slice.

## Capabilities

### New Capabilities

- `application-command-flows`: Defines application orchestration, CLI syntax, injected command seams, and centralized outcomes for `inspect`, `check`, `presets`, and `presets show`.

### Modified Capabilities

None.

## Impact

- Adds a focused application package and report-ready request/result types over existing `project`, `config`, `analysis`, `policy`, `budget`, and `preset` packages.
- Reworks `internal/cli` command construction and tests while preserving the public executable name and thin `cmd/deadweight.gdt/main.go` composition root.
- Adds no runtime dependency on Godot, OpenSpec, Node.js, networking, or new Go modules.
- Advances MVP acceptance criteria 1, 2, 14, 16–21, 23, and 24 and Milestone 6 command/exit-code wiring without completing the separate final report/golden-test slice.
