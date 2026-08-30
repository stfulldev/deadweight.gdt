## ADDED Requirements

### Requirement: Console metrics use their own confidence
Inspect metrics, check comparison actuals, and selected contribution values SHALL use the reliability marker of the rendered metric rather than blindly applying the report-wide or row-wide summary. When one or more metric classifications materially differ from the report summary, text output SHALL add one concise deterministic qualification section listing only those differing metrics, their confidence, and machine-readable reasons in canonical order. When every metric matches the summary, no redundant per-metric section SHALL be added.

#### Scenario: One uncertain resource metric
- **WHEN** only `external_resources` is lower-bound and the report summary is lower-bound
- **THEN** the seven exact metric values are unmarked, the resource value uses `+`, and the qualification section identifies the exact metrics that differ from the summary

#### Scenario: Uniform unresolved closure
- **WHEN** all eight metrics are lower-bound for the same unresolved scene closure
- **THEN** all values use `+` and no redundant mixed-confidence section is emitted

#### Scenario: Contribution metric differs from row summary
- **WHEN** a selected contribution has lower-bound row reliability but its selected additive metric is exact
- **THEN** the selected value is unmarked while the row's conservative reliability remains visible

