## Why

Portable JSON reports can be archived today, but users still have to build their own tooling to decide what changed or whether a later scene regressed. Issue [#53](https://github.com/stfulldev/deadweight.gdt/issues/53) and Draft PR [#63](https://github.com/stfulldev/deadweight.gdt/pull/63) add a deterministic offline comparison flow that respects the confidence contract delivered by issue #54.

## What Changes

- Add `diff <before.json> <after.json>` for compatible schema-v1 `inspect`, `tree`, or `check` reports with text output by default and `--format json` as an alternative.
- Compare changed metric values and confidence, report reliability, coverage, diagnostics, unique scene-dependency identities, and budget evaluation when both compatible reports contain it.
- Publish signed absolute metric/coverage deltas and deterministic added, removed, or changed evidence without depending on console prose or checkout paths.
- Reject malformed reports, unsupported schema versions or kinds, different report kinds, and different portable root scene identities as fatal exit `2` outcomes with no partial diff.
- Qualify numerical changes using exact/lower-bound/approximate evidence so a smaller partial value is not claimed as a proven improvement.
- Add repeatable `--fail-on-increase METRIC` and `--fail-on-reliability` enforcement. A proven selected increase exits `1`; selected evidence that cannot be concluded safely, or an enforced reliability degradation, exits `3`; valid comparison without a triggered policy exits `0`.
- Document reproducible baseline capture, comparison, enforcement, and intentional update workflows for CI.

Goals: make report comparison portable, deterministic, safe under partial evidence, and usable in CI without external services.

Non-goals: remote baseline storage, automatic Git revision checkout, PR comments or annotations, SARIF, project-wide scans, percentages or statistical significance, and runtime profiling.

Compatibility: all existing commands, report kinds, fields, exit meanings, and default scene flows remain unchanged. `diff`, its flags, and schema-v1 kind `diff` are additive. Older schema-v1 inputs without per-metric confidence remain readable through conservative report-summary fallback; unknown compatible fields are ignored.

Affected MVP acceptance criteria: deterministic and portable JSON, explicit uncertainty, centralized exit taxonomy, report-first non-fatal policy outcomes, cross-platform behavior, and full repository quality gates.

## Capabilities

### New Capabilities

- `report-baselines-and-diffs`: Defines compatible baseline inputs, the semantic diff model, confidence-aware assessment, deterministic evidence comparison, and opt-in regression enforcement.

### Modified Capabilities

- `application-command-flows`: Adds the project-independent `diff` application/CLI flow, flags, argument validation, and exit mapping.
- `deterministic-console-reports`: Adds deterministic human-readable diff presentation and explicit uncertainty/regression outcomes.
- `versioned-json-reports`: Adds a compatible schema-v1 `diff` document kind and the machine-readable diff payload.

## Impact

The change adds an internal baseline/diff domain, injectable file-reading orchestration, one Cobra command, text and JSON presenters, schema-v1 definitions, deterministic goldens, CLI integration tests, and README CI guidance. It adds no external service, database, Godot, Node.js, OpenSpec, or network dependency to the shipped binary.
