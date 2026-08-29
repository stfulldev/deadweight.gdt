# analysis-completeness Specification

## Purpose
Defines how successful static scene analysis communicates completeness, reliability, checked coverage, and deterministic warning diagnostics without weakening fatal error boundaries.

## Requirements

### Requirement: Successful analysis exposes status and reliability
Every successful root result SHALL expose status `complete` or `partial` and reliability `exact`, `lower_bound`, or `approximate`. Status SHALL be `complete` with `exact` reliability only when every reachable scene occurrence and every declared resource is statically accounted for and no inherited or unsupported parent semantics remain. Any non-inheritance partial reason SHALL produce `partial` with `lower_bound`; any inherited-scene or override evidence SHALL produce `partial` with `approximate`, and `approximate` MUST win when both classes occur.

#### Scenario: Fully resolved closure
- **WHEN** all reachable format-3 text scenes and all declared resources resolve successfully without inheritance or unsupported parent semantics
- **THEN** the successful result is `complete` and `exact`

#### Scenario: Mixed partial reasons
- **WHEN** one closure contains both a missing nested scene and inherited-scene evidence
- **THEN** the successful result is `partial` and `approximate`

### Requirement: Coverage is checked and occurrence-aware
A successful result SHALL publish non-negative `int64` coverage for unique successfully parsed scene files, resolved scene-instance occurrences, unresolved scene-instance occurrences, and inherited-scene occurrences. Repeated resolved or unresolved child summaries SHALL multiply occurrence coverage with checked arithmetic, parsed file coverage MUST remain the unique invocation-cache count, and inherited coverage SHALL preserve checked occurrence multiplicity without treating an inherited root as a new scene-instance occurrence.

#### Scenario: Repeated mixed coverage
- **WHEN** a cached child with one resolved and one unresolved nested occurrence is mounted 100 times
- **THEN** occurrence coverage is multiplied 100 times while every successfully parsed canonical scene file is counted once

#### Scenario: Coverage overflow
- **WHEN** grouping or coverage arithmetic would exceed non-negative `int64`
- **THEN** analysis fails with the typed arithmetic diagnostic and publishes no usable result

### Requirement: Unresolved declared resources affect completeness honestly
Every unresolved external-resource declaration in a successfully parsed scene SHALL make analysis partial, preserving its declaring canonical scene, document-local resource ID, raw path, and typed resolution reason. A successfully resolved ordinary `.tres`, texture, material, script, audio, or other non-scene resource SHALL contribute only its canonical resource identity and MUST NOT make analysis partial merely because it is not deeply parsed.

#### Scenario: Missing ordinary resource
- **WHEN** a parsed scene declares a texture path that is missing or outside the project
- **THEN** analysis is `partial` with `lower_bound` reliability and retains an unavailable-resource warning

#### Scenario: Existing ordinary resource
- **WHEN** a parsed scene declares an existing canonical texture, material, script, audio, or `.tres` path that is not a scene mount
- **THEN** that reference contributes to `external_resources` without changing `complete exact` status

### Requirement: Every partial branch produces grouped deterministic diagnostics
Every unresolved scene mount, imported or binary scene, inherited target, unavailable declaration, placeholder, UID-only or `user://` path, `SubResource` scene source, unsupported scene target, and unsupported parent finding SHALL produce validated warning evidence. Diagnostics SHALL use `SB1001` for otherwise unresolved scene instances, `SB1002` for imported or binary scenes, `SB1003` for inheritance, `SB1004` for missing, unreadable, empty, or outside-project resource paths, `SB1005` for placeholders, `SB1006` for UID-only or `user://` paths, and `SB1008` for unsupported parent semantics. Equivalent evidence SHALL be grouped by stable code, display path, classification or reason, and target/resource identity with checked occurrence totals; returned diagnostics SHALL be owned and sorted by severity, code, display path, line, column, resource, and message independently of map order.

#### Scenario: Repeated imported target
- **WHEN** one imported scene target occurs repeatedly through cached scene summaries
- **THEN** one `SB1002` warning retains the complete checked occurrence count

#### Scenario: Distinct unresolved reasons
- **WHEN** equivalent raw targets fail for different resolution reasons or declaring scenes
- **THEN** they remain distinguishable diagnostic groups in deterministic order

#### Scenario: Unsupported parent
- **WHEN** a parsed local node has an invalid, missing, ambiguous, absolute, or otherwise unsupported parent path
- **THEN** analysis is `partial lower_bound` and emits grouped `SB1008` evidence without inventing tree depth

### Requirement: Fatal analysis remains separate from partial success
Missing or unreadable root input, an invalid root extension or unsupported format, malformed supported root or nested `.tscn`, invalid project or configuration input, a resolved dependency cycle, invalid retained evidence, and arithmetic overflow SHALL remain fatal. A fatal analyzer call MUST return a zero result rather than status, reliability, metrics, coverage, graph, or warnings from a truncated traversal. Missing, unreadable, imported, binary, or otherwise unsupported nested targets SHALL remain successful partial analysis.

#### Scenario: Malformed resolved nested text scene
- **WHEN** a resolved `.tscn` target uses the supported text-scene class but cannot be parsed
- **THEN** analysis fails and returns no partial result

#### Scenario: Missing nested text scene
- **WHEN** a nested `.tscn` declaration cannot be resolved or opened
- **THEN** analysis succeeds as `partial lower_bound`, counts one known instance root, and emits grouped warning evidence

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
