## Purpose

Defines deterministic discovery and strict version-1 decoding of project configuration, including optional policy declarations, static validation, canonical JSON Schema, and typed failures.

## ADDED Requirements

### Requirement: Configuration discovery follows frozen priority
After a project root is known, the system SHALL select a non-empty explicit config path before considering `<project>/.deadweight.gdt.json`. A missing explicit path MUST fail, while an absent implicit default MUST return a successful no-configuration result. A selected path that exists but is not a readable regular file SHALL fail rather than being treated as absent or skipped.

#### Scenario: Explicit config wins
- **WHEN** an explicit config path and an implicit project config both exist
- **THEN** discovery selects only the explicit path
- **AND** the implicit file is not used as a fallback

#### Scenario: Missing implicit config
- **WHEN** no explicit path is supplied and `<project>/.deadweight.gdt.json` does not exist
- **THEN** discovery succeeds with configuration absent

#### Scenario: Missing explicit config
- **WHEN** an explicit config path is supplied but does not identify a readable regular file
- **THEN** discovery returns an actionable fatal configuration error naming that path

### Requirement: Version-one JSON decoding is strict
A selected config SHALL contain exactly one JSON object document. Decoding MUST reject malformed JSON, trailing non-whitespace content, unknown fields at the top level or inside budgets and profiles, explicit `null` for any declared field, and values whose JSON type does not match the version-one model. The required `version` field SHALL be the integer `1`; missing, non-integer, or unsupported versions MUST fail.

#### Scenario: Minimal document
- **WHEN** the selected file contains only `{ "version": 1 }`
- **THEN** decoding succeeds with no selector, no budgets, no profiles, and `fail_on_partial` false

#### Scenario: Unknown or trailing content
- **WHEN** a document contains an unknown owned field or a second JSON value after the config object
- **THEN** decoding fails and identifies the invalid field or trailing content

#### Scenario: Unsupported version
- **WHEN** `version` is absent, is not an integer, or is not exactly `1`
- **THEN** decoding fails as invalid version-one configuration

### Requirement: All eight budget limits remain optional integers
Top-level and custom-profile budget objects SHALL independently support exactly `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, `shadow_lights`, `external_resources`, and `scene_dependencies`. Each present limit MUST be a non-negative signed 64-bit integer. Absence SHALL remain distinct from a present zero so downstream policy resolution can skip absent limits and enforce zero as a hard limit.

#### Scenario: Full zero budget
- **WHEN** all eight budget properties are present with value zero
- **THEN** all eight values are retained as configured limits equal to zero

#### Scenario: Invalid budget value
- **WHEN** any budget is negative, fractional, a string, a boolean, `null`, or outside the supported signed 64-bit range
- **THEN** decoding or static validation fails with the affected budget path

#### Scenario: Unknown metric
- **WHEN** a budget object contains a property outside the frozen eight metric IDs
- **THEN** decoding fails instead of ignoring the property

### Requirement: Selectors and profile declarations obey static rules
The top-level `preset` and `profile` selectors MUST NOT coexist, and every present selector, custom-profile map key, and `extends` value SHALL match the case-sensitive pattern `^[a-z0-9][a-z0-9._-]{0,63}$`. Profile metadata fields SHALL preserve whether they were omitted; a present `platform` MUST be non-empty, `target_fps` MUST be a non-negative integer, `renderer` MUST be one of `forward_plus`, `mobile`, `compatibility`, or `unspecified`, and `quality` MUST be one of `low`, `balanced`, `high`, or `custom`.

#### Scenario: Full profile declarations
- **WHEN** a version-one document contains valid top-level selectors, metadata, profile inheritance references, optional budgets, and `fail_on_partial`
- **THEN** every declaration is retained without applying inheritance or policy precedence

#### Scenario: Mutually exclusive selectors
- **WHEN** both top-level `preset` and `profile` are present
- **THEN** static validation fails with an actionable selector-conflict error

#### Scenario: Invalid profile declaration
- **WHEN** a profile key or reference violates the ID pattern, platform is empty, target FPS is negative, or renderer or quality is outside its enum
- **THEN** static validation fails with the exact declaration path

### Requirement: Dynamic profile rules remain a separate phase
Successful version-one decoding and static validation SHALL return owned declarations without requiring that selector or `extends` targets already exist. Built-in/custom target existence, custom-ID collision with built-ins, inheritance cycles, maximum depth, field-by-field merge, selector precedence, and effective-budget existence MUST be evaluated by the later dynamic profile-resolution phase rather than by JSON shape validation.

#### Scenario: Structurally valid unresolved reference
- **WHEN** a profile contains an `extends` ID that matches the pattern but is not resolvable in the current declarations
- **THEN** version-one decoding and static validation retain the reference successfully
- **AND** dynamic profile resolution remains responsible for rejecting it

### Requirement: Canonical schema matches the version-one contract
The repository SHALL include `schema/deadweight.gdt.schema.json` as valid JSON Schema Draft 2020-12. It MUST require `version` with `const: 1`, reject additional properties on every object, describe the frozen eight non-negative integer budgets, enforce selector exclusion, ID patterns, profile metadata types, renderer and quality enums, and optional profile maps. It MUST NOT encode dynamic reference existence, collision, cycle, inheritance-depth, merge, or effective-policy rules.

#### Scenario: Schema structural parity
- **WHEN** the canonical schema and Go version-one declaration model are inspected by automated tests
- **THEN** their field names, required version, metric set, ID pattern, enums, and additional-property constraints agree

#### Scenario: No runtime schema dependency
- **WHEN** the standalone binary loads a project configuration
- **THEN** it performs the version-one Go decoding and validation contract without requiring a schema engine, Node.js, network access, OpenSpec, or Godot

### Requirement: Configuration failures are typed and actionable
Every discovery, read, decode, or static-validation failure SHALL retain the selected source path when available, identify the relevant field or rule when known, expose diagnostic code `SB2003`, and return no usable configuration value. Equivalent inputs MUST produce deterministic messages without Go stack traces.

#### Scenario: Nested field failure
- **WHEN** `profiles.shipping.budgets.nodes` contains an invalid value
- **THEN** the fatal error exposes `SB2003`, the config source, and the nested field path
- **AND** no partially decoded configuration is published
