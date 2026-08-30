## ADDED Requirements

### Requirement: Completeness finalization publishes per-metric evidence
Completeness finalization SHALL publish validated confidence for every frozen root metric from retained unresolved-scene, imported-scene, inherited-scene, unavailable-resource, unsupported-resource, and unsupported-parent evidence. It SHALL retain all applicable machine-readable reasons, sort them deterministically, and derive the existing report-wide reliability and status as the conservative summary of those entries.

#### Scenario: Mixed but unrelated partial evidence
- **WHEN** an unavailable ordinary resource affects only `external_resources`
- **THEN** completeness is `partial lower_bound`, `external_resources` is lower-bound, and unaffected metrics remain exact

#### Scenario: Mixed evidence precedence
- **WHEN** one metric is affected by inherited semantics and unresolved scene content
- **THEN** its confidence is approximate with both reasons and the report-wide reliability is approximate

