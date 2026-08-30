## ADDED Requirements

### Requirement: Diff JSON is a compatible version-one document kind
The version-one report schema SHALL accept kind `diff` as a compatible new document kind without removing, renaming, or reinterpreting established inspect, tree, check, or error fields. A diff document SHALL contain the standard schema/tool envelope and one required `diff` payload with compatible report kind, portable root scene, semantic change collections, before/after reliability, and enforcement outcome. It MUST NOT contain scene-analysis, policy, evaluation, dependency-tree, or fatal-error payloads belonging to other kinds. JSON framing SHALL remain deterministic UTF-8, two-space indented, ANSI-free, and exactly one trailing LF.

#### Scenario: Empty JSON diff
- **WHEN** compatible reports are semantically equal
- **THEN** kind `diff` contains empty change arrays, `changed: false`, and a schema-valid enforcement result

#### Scenario: Complete changed JSON diff
- **WHEN** compatible reports differ across metrics, coverage, diagnostics, dependencies, and check evaluation
- **THEN** kind `diff` retains every deterministic machine-readable change and confidence assessment with `changed: true`

#### Scenario: Existing version-one consumer
- **WHEN** an existing consumer ignores the newly introduced diff kind
- **THEN** every established version-one document kind and field meaning remains unchanged

### Requirement: JSON errors cover baseline input failures
After JSON format is selected, a fatal baseline read, decode, compatibility, or semantic validation error SHALL use the existing schema-version-one kind `error` document on stderr, leave stdout empty, and exit `2`. Raw baseline content, canonical file paths beyond the user-provided actionable path, and Go implementation details MUST NOT leak into the error document.

#### Scenario: Malformed candidate in JSON mode
- **WHEN** the candidate contains malformed JSON after `--format json` is validated
- **THEN** stderr receives one existing kind `error` document, stdout remains empty, and the process exits `2`

