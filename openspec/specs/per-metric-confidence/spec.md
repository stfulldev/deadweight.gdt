# per-metric-confidence Specification

## Purpose

Defines deterministic confidence and machine-readable reasons for each frozen metric so consumers can distinguish exact values from lower bounds and inherited approximations.

## Requirements

### Requirement: Every metric has deterministic confidence
Every successful analysis and every contribution record SHALL expose one confidence entry for each frozen metric in canonical metric order. Each entry SHALL contain reliability `exact`, `lower_bound`, or `approximate` and a sorted, duplicate-free collection of stable reason codes. Exact confidence MUST have no reasons; non-exact confidence MUST have at least one reason; and a missing or invalid entry SHALL make the result invalid rather than defaulting unavailable evidence to an exact zero.

#### Scenario: Complete supported closure
- **WHEN** every reachable scene, resource declaration, parent relationship, and inheritance behavior is statically supported
- **THEN** all eight analysis metrics and all emitted contribution metrics are `exact` with empty reasons

#### Scenario: Invalid incomplete metadata
- **WHEN** a producer omits a frozen metric, repeats a reason, uses non-canonical order, or assigns non-exact confidence without a reason
- **THEN** validation fails before a report is emitted

### Requirement: Evidence affects only metrics it can change
The analyzer SHALL derive confidence from retained evidence without spreading unrelated uncertainty. Unresolved, imported, unavailable, placeholder, subresource, or unsupported scene content SHALL make every root metric that can contain hidden scene-closure evidence a `lower_bound`. An unresolved ordinary external-resource declaration SHALL make `external_resources` a `lower_bound` while leaving occurrence metrics, `tree_depth`, and `scene_dependencies` exact unless other evidence affects them. Unsupported parent semantics SHALL make `tree_depth` a `lower_bound` while leaving metrics independent of parent composition exact. Stable reason codes SHALL distinguish unresolved scene instances, imported scenes, unsupported scenes, subresource scenes, placeholders, unavailable scenes, unavailable resources, unsupported resource paths, and unsupported parents.

#### Scenario: Missing ordinary texture
- **WHEN** a parsed scene has one unavailable texture declaration and no other partial evidence
- **THEN** `external_resources` is `lower_bound` with an unavailable-resource reason
- **AND** the other seven frozen metrics remain `exact`

#### Scenario: Unsupported parent only
- **WHEN** all content is available but one local node parent cannot be composed statically
- **THEN** `tree_depth` is `lower_bound` with an unsupported-parent reason
- **AND** the other seven frozen metrics remain `exact`

#### Scenario: Imported scene mount
- **WHEN** an imported scene mount cannot be expanded statically
- **THEN** every metric that may omit the mounted closure is `lower_bound` with an imported-scene reason
- **AND** known root occurrences remain in the reported metric values rather than being replaced by zero

### Requirement: Inherited uncertainty is approximate where applicable
Inherited scene roots and override evidence SHALL make every metric whose value may change under unsupported inheritance semantics `approximate` with an inherited-scene reason. Approximate SHALL take precedence over lower-bound reliability for a metric affected by both classes, while all applicable reasons remain available in deterministic order.

#### Scenario: Inheritance mixed with missing content
- **WHEN** inherited override uncertainty and unavailable nested content can both affect one metric
- **THEN** that metric is `approximate` and retains both inherited and unavailable reason codes

### Requirement: Aggregate reliability is a conservative summary
Analysis-wide reliability SHALL equal the conservative maximum of all root metric confidence entries, and contribution-wide reliability SHALL equal the conservative maximum of that row's metric confidence entries, using `exact` then `lower_bound` then `approximate` precedence. Analysis status SHALL remain `complete` only when the summary is exact and SHALL otherwise remain `partial`.

#### Scenario: One lower-bound metric
- **WHEN** seven root metrics are exact and one is lower-bound
- **THEN** report-wide reliability is `lower_bound` and status is `partial`

#### Scenario: Approximate wins summary
- **WHEN** at least one root metric is approximate and other metrics are exact or lower-bound
- **THEN** report-wide reliability is `approximate` and status is `partial`
