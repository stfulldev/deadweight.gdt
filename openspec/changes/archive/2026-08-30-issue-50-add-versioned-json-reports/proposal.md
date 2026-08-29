## Why

The current CLI exposes deterministic human-readable reports, but automation must parse presentation text to retain metrics, completeness, diagnostics, and budget outcomes. GitHub issue [#50](https://github.com/stfulldev/deadweight.gdt/issues/50) and Draft PR [#58](https://github.com/stfulldev/deadweight.gdt/pull/58) establish the first versioned machine-readable contract required by the MVP 0.2 tracker [#57](https://github.com/stfulldev/deadweight.gdt/issues/57), before baselines, diffs, or ecosystem integrations depend on an unstable shape.

## What Changes

- Add `--format text|json` to `inspect` and `check`, with `text` as the compatible default.
- Define report schema version 1 for inspect results, check results, and fatal diagnostics selected after JSON mode is known.
- Publish a committed JSON Schema and deterministic, portable serialization rules for tool metadata, scene identity, analysis evidence, ordered metrics, diagnostics, effective policy, comparisons, and verdicts.
- Preserve process exit meanings and priority for success, exceeded budgets, fatal errors, and rejected partial evidence.
- Keep JSON streams machine-only: one complete document, no ANSI, no human headings, and no Go stack trace.
- Preserve all existing text-report bytes and semantics when JSON is not selected.

Goals:

- Give local scripts and CI a stable contract that does not parse console prose.
- Retain enough structured evidence to reconstruct every current inspect or check report.
- Make identical project-relative inputs byte-stable across supported operating systems and checkout paths.
- Establish an explicit schema-evolution boundary before later MVP 0.2 contribution and diff features.

Non-goals:

- Adding SARIF, JUnit, HTML, GitHub annotations, PR comments, baseline comparison, dependency-tree rendering, or project-wide scanning.
- Changing any metric definition, preset value, configuration schema, budget comparison, reliability classification, diagnostic code, or process exit code.
- Exposing canonical machine-specific absolute paths as portable report identity.
- Adding Godot, Node.js, OpenSpec, network access, or another shipped runtime dependency.

Compatibility impact: additive. Existing invocations continue to select deterministic text output, and the frozen 0.1 command, analysis, budget, diagnostic, and exit-code behavior remains authoritative. The directly affected MVP acceptance criteria are §30.19–§30.22; the change also preserves the Godot-free gate in §30.24 and exercises existing inspect and fatal-cycle behavior from §30.1 and §30.10.

## Capabilities

### New Capabilities

- `versioned-json-reports`: Defines report schema versioning, portable inspect/check documents, structured fatal diagnostics, deterministic serialization, and compatibility rules for machine-readable output.

### Modified Capabilities

- `application-command-flows`: Extends inspect/check CLI selection and centralized presentation so output format changes representation without changing application analysis, verdicts, or exit outcomes.

## Impact

- `internal/cli`: format parsing, stream selection, fatal-output routing, and centralized process behavior.
- `internal/report`: format-neutral projections plus deterministic JSON encoders alongside unchanged text renderers.
- `internal/app`, `internal/analysis`, `internal/budget`, `internal/policy`, and `internal/diagnostic`: existing report-ready domain models remain authoritative inputs; small owned projections may be added without moving serialization into domain packages.
- `schema/`: a committed report schema version 1 separate from configuration schema version 1.
- CLI/golden/schema/integration tests, README machine-readable usage, and MVP 0.2 acceptance evidence.
- No new production Go dependency or runtime service; schema validation tooling remains test-only.
