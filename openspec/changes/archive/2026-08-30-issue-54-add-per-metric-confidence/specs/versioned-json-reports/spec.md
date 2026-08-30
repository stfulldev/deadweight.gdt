## ADDED Requirements

### Requirement: JSON metrics expose confidence and reasons
Every root metric object and every contribution metric object emitted by the current producer in a successful schema-version-one inspect, check, or tree document SHALL contain a `confidence` object with required `reliability` and `reasons` fields. Reliability SHALL be `exact`, `lower_bound`, or `approximate`; reasons SHALL be a deterministic duplicate-free array of stable machine-readable reason codes; and every frozen metric SHALL remain present in canonical order. The field SHALL remain optional in the version-one schema so older version-one documents continue to validate. This compatible extension MUST NOT change metric IDs, values, aggregation modes, availability semantics, report-wide reliability, contribution-wide reliability, or schema version.

#### Scenario: Exact metric JSON
- **WHEN** a metric is unaffected by unavailable or approximate evidence
- **THEN** its confidence contains reliability `exact` and an empty reasons array

#### Scenario: Mixed-confidence inspect JSON
- **WHEN** one unavailable ordinary resource affects only `external_resources`
- **THEN** the root metrics retain canonical order, only `external_resources` is lower-bound with its reason, and the report-wide reliability remains lower-bound

#### Scenario: Unavailable contribution value
- **WHEN** a contribution metric uses maximum or unique-union aggregation without an owned numeric value
- **THEN** its JSON record still contains confidence and reasons
- **AND** it does not add a misleading zero-valued field

#### Scenario: Schema validation
- **WHEN** an inspect, check, or tree report with per-metric confidence is validated against the committed version-one schema
- **THEN** every root and contribution metric confidence object satisfies the required enum and reason-array constraints

#### Scenario: Earlier version-one document
- **WHEN** a valid document from an earlier version-one producer has no confidence fields
- **THEN** it continues to validate against the evolved version-one schema
