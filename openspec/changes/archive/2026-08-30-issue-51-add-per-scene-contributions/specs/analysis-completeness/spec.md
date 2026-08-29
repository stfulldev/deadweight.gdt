## ADDED Requirements

### Requirement: Partial evidence qualifies affected contributions
Completeness finalization SHALL classify contribution rows from their retained unresolved, imported, inherited, unavailable, and unsupported-parent evidence while preserving the existing conservative whole-analysis status and reliability. A row whose reported direct values or depth candidate may omit nested evidence SHALL be `lower_bound`; a row affected by inherited or override semantics SHALL be `approximate`; and approximate SHALL take precedence for mixed evidence. Contribution qualification MUST retain stable evidence links or classifications without inventing unavailable values.

#### Scenario: Mixed reliable and unresolved contributions
- **WHEN** one supported child is fully analyzable and another mount is unavailable
- **THEN** the supported direct row may be exact, the unavailable row is lower-bound, and the whole analysis remains partial lower-bound

#### Scenario: Inherited and missing evidence mix
- **WHEN** one contribution is affected by inherited overrides and unavailable nested content
- **THEN** that contribution and the whole analysis are approximate

#### Scenario: Unknown parent depth
- **WHEN** a contribution has additive values but its root-relative depth cannot be composed because of unsupported parent evidence
- **THEN** the additive values remain reported with their justified reliability and no exact depth candidate is emitted

