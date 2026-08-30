## Why

The analyzer currently publishes one conservative reliability value for an entire report or contribution row, so unrelated unavailable evidence can make otherwise exact metrics look uncertain. Issue [#54](https://github.com/stfulldev/deadweight.gdt/issues/54) and Draft PR [#62](https://github.com/stfulldev/deadweight.gdt/pull/62) add deterministic per-metric confidence before MVP 0.2 consumers begin depending on the broader explainability surface.

## What Changes

- Add a stable `exact` / `lower_bound` / `approximate` confidence model for every frozen metric, with deterministic machine-readable reason codes tied to concrete static-analysis evidence.
- Preserve report-wide and contribution-wide reliability as conservative summaries of their per-metric confidence values.
- Emit confidence and reason codes on root metrics and contribution metric records in schema-v1 JSON.
- Use each metric's own confidence marker in text reports and add concise qualifications only when metric confidence materially differs from the report summary.
- Treat unavailable evidence as uncertainty, never as a known zero.
- Cover every frozen metric and each supported evidence class with deterministic analysis, JSON, and console-report tests.

Goals: make exact results visibly exact despite unrelated gaps, make uncertainty traceable to evidence, and keep all existing metric values and aggregation semantics unchanged.

Non-goals: probabilities, predictive confidence, empirical calibration, new metrics, dynamic Godot execution, or changes to budget thresholds and verdict policy.

Compatibility: existing reliability/status fields, metric identifiers, values, exit semantics, and schema version remain stable. JSON metric objects gain additive confidence metadata; deterministic text can gain mixed-confidence qualifications when required.

Affected MVP acceptance criteria: deterministic output, honest PARTIAL reporting, machine-readable schema-v1 reports, explainable contribution evidence, and the full Go quality-gate matrix.

## Capabilities

### New Capabilities

- `per-metric-confidence`: Defines the confidence taxonomy, reason codes, evidence-to-metric impact rules, validation, and conservative summaries.

### Modified Capabilities

- `analysis-completeness`: Makes report-wide reliability the conservative summary of complete per-metric confidence metadata.
- `scene-contributions`: Publishes confidence and reasons for every metric on every contribution record while retaining row-wide compatibility metadata.
- `deterministic-console-reports`: Formats values with their own confidence and explains only meaningful mixed-confidence differences.
- `versioned-json-reports`: Adds required confidence objects to root and contribution metric records in schema-v1 JSON.

## Impact

The change affects analysis domain models and completeness finalization, contribution compaction and cloning, inspect/check/contribution text formatting, JSON projections and schema, checked-in goldens, and end-user documentation. It adds no runtime dependency and preserves the standalone Go binary, offline operation, metric calculations, and CLI command surface.
