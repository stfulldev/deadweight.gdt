## ADDED Requirements

### Requirement: Scene JSON includes portable contribution evidence
Every successful version-one inspect and check analysis payload SHALL include an ordered `contributions` collection and ordered unique-union evidence. Each contribution SHALL retain stable kind, portable scene or unresolved target identity, immediate portable mount context, checked occurrence count, row reliability, and all eight metric entries in frozen order with explicit `additive`, `maximum`, or `unique_union` aggregation. Additive entries SHALL contain non-negative signed 64-bit values, maximum entries SHALL contain a known candidate or explicit unavailability, and unique-union entries MUST NOT contain a misleading per-row additive value. Unique evidence SHALL represent each portable resource or dependency identity once with deterministic referrer identities or contexts.

#### Scenario: Exact inspect contributions
- **WHEN** a complete scene is rendered as inspect JSON
- **THEN** its contribution rows validate against the committed version-one schema and additive values reconcile to the root metrics

#### Scenario: Failed check contributions
- **WHEN** a check exceeds a configured budget in JSON mode
- **THEN** its complete contribution and unique-evidence collections remain available alongside comparisons before exit code `1`

#### Scenario: Shared resource evidence
- **WHEN** several parsed scenes refer to one resolved resource
- **THEN** JSON represents the resource once with every deterministic portable referrer
- **AND** contribution metric entries identify `external_resources` as `unique_union` without per-row ownership values

#### Scenario: Partial portable contribution
- **WHEN** an unresolved imported or inherited target contributes evidence
- **THEN** JSON retains its portable context and lower-bound or approximate reliability without exposing a canonical absolute path

#### Scenario: Compatible version-one extension
- **WHEN** an existing version-one consumer ignores the new contribution fields
- **THEN** all previously required inspect fields and meanings remain unchanged and consumable

### Requirement: JSON top selection is a deterministic projection
When inspect JSON is requested with valid top-contributor selectors, the payload SHALL retain the authoritative full `contributions` and unique-evidence collections and SHALL additionally identify the selected metric, requested limit, and deterministically truncated row identities or projections. Selecting top contributors MUST NOT change analysis, root metrics, contribution values, diagnostics, or process exit code.

#### Scenario: JSON top five nodes
- **WHEN** inspect uses JSON with `--metric nodes --top 5`
- **THEN** the payload retains full contribution evidence and identifies at most five rows in the same deterministic order as the text projection
