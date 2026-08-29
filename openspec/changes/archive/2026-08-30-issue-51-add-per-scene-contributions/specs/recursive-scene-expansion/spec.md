## ADDED Requirements

### Requirement: Recursive application preserves direct contribution evidence
The recursive analyzer SHALL construct one-occurrence direct contribution evidence from each cached local summary and SHALL apply child contribution rows with the same checked reachable multiplicity used for occurrence metrics. Nested rows SHALL be propagated independently from parent direct values, root-relative depth candidates SHALL be composed at each mount, and equivalent immediate contexts SHALL be compacted without losing occurrences. Graph discovery and contribution construction MUST continue sharing the invocation-scoped parse and local-summary cache.

#### Scenario: Cached repeated child contribution
- **WHEN** a supported child is mounted 100 times through equivalent context
- **THEN** its document and local summary are built once while its direct contribution occurrence count and additive values are applied 100 times

#### Scenario: Diamond descendant contribution
- **WHEN** two parent scenes reach the same canonical descendant
- **THEN** the descendant's cached direct evidence is reused for both reachable paths
- **AND** its context-aware contribution multiplicities account for both paths without duplicating parse work

#### Scenario: Mounted depth composition
- **WHEN** a child contribution has a known relative depth candidate and is applied at a known parent mount depth
- **THEN** its root-relative candidate is composed using checked depth arithmetic
- **AND** an unknown required depth remains unknown rather than guessed

