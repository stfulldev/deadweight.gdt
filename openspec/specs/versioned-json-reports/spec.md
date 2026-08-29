# versioned-json-reports Specification

## Purpose

Defines a stable, deterministic, and portable JSON contract for automation to consume inspect, check, and fatal diagnostic outcomes without parsing human-readable console text.

## Requirements

### Requirement: Scene commands provide an explicit versioned JSON representation
The `inspect` and `check` commands SHALL accept `--format text` and `--format json`, SHALL use `text` when the flag is absent, and MUST reject every other value before invoking the application flow. Selecting JSON MUST change only representation; scene resolution, analysis, effective policy, budget evaluation, verdict priority, and process exit codes MUST remain identical to the corresponding text invocation.

#### Scenario: Text remains the default
- **WHEN** a user invokes `inspect` or `check` without `--format`
- **THEN** the command emits its existing deterministic human-readable report
- **AND** no JSON compatibility change alters its application request, report content, or exit outcome

#### Scenario: JSON is explicitly selected
- **WHEN** a user invokes `inspect` or `check` with `--format json`
- **THEN** the command emits the matching schema-versioned JSON document instead of human-readable report text
- **AND** the underlying application result and process exit outcome are unchanged

#### Scenario: Unknown format is rejected before analysis
- **WHEN** a scene command receives a format other than `text` or `json`
- **THEN** it returns an actionable usage failure with exit code `2`
- **AND** it does not invoke project discovery, configuration loading, or scene analysis

### Requirement: Every JSON document has a stable discriminated envelope
Every emitted JSON document SHALL contain integer `schema_version` equal to `1`, a `kind` discriminator, and tool name/version metadata. Successful inspect output MUST use kind `inspect`; every report-producing check outcome MUST use kind `check`; and a fatal outcome rendered after JSON mode is selected MUST use kind `error`. Each kind MUST have a committed Draft 2020-12 JSON Schema definition, and documents MUST NOT combine payloads belonging to different kinds.

#### Scenario: Inspect envelope
- **WHEN** inspect completes with exact, lower-bound, or approximate evidence in JSON mode
- **THEN** its document uses schema version `1`, kind `inspect`, and contains one inspect payload without check or fatal-error payloads

#### Scenario: Check envelope for a nonzero verdict
- **WHEN** check produces a budget-failed or partial-rejected report in JSON mode
- **THEN** its document uses kind `check`, retains the complete analysis and evaluation payload, and exits with code `1` or `3` respectively

#### Scenario: Fatal envelope
- **WHEN** JSON mode has been selected and project discovery, configuration, resolution, parsing, analysis, policy resolution, report validation, or another command operation fails fatally
- **THEN** stderr receives one kind `error` document containing the stable diagnostic code when available and actionable error text
- **AND** stdout remains empty and the process exits `2`

### Requirement: Inspect JSON retains complete semantic analysis evidence
An inspect document SHALL contain a portable preferred scene identity, configuration-presence/source metadata, analysis status, reliability, all eight metrics in canonical order, checked coverage, and all user-visible diagnostics in their deterministic order. Metric values and coverage counts MUST remain non-negative signed 64-bit JSON integers. Diagnostic records MUST retain stable code, severity, message, occurrence count, and available source/resource context without requiring a consumer to parse formatted prose.

#### Scenario: Complete exact inspect document
- **WHEN** inspect analyzes a complete exact scene in JSON mode
- **THEN** the document contains status `complete`, reliability `exact`, all eight ordered metrics, exact coverage, and an empty diagnostics collection

#### Scenario: Partial inspect document
- **WHEN** inspect succeeds with unresolved or inherited evidence in JSON mode
- **THEN** the document contains status `partial`, the authoritative `lower_bound` or `approximate` reliability, checked occurrence-aware coverage, and every grouped warning diagnostic
- **AND** the successful partial inspect exits `0`

### Requirement: Check JSON retains effective policy and evaluation semantics
A check document SHALL contain the complete inspect evidence plus the effective policy kind/identity, available policy metadata, effective partial policy, every configured comparison in canonical metric order, exceeded count, and final verdict. Each comparison MUST contain metric ID, observed value, limit, signed delta, and pass/fail state. Built-in preset output MUST retain its heuristic and experimental lifecycle metadata so machine consumers cannot mistake it for a performance guarantee.

#### Scenario: Passing built-in preset check
- **WHEN** a complete scene passes a built-in preset in JSON mode
- **THEN** the document identifies the preset, retains its heuristic/experimental metadata and all configured comparisons, reports verdict `PASSED`, and exits `0`

#### Scenario: Budget failure
- **WHEN** one or more configured limits are exceeded and partial evidence is not rejected
- **THEN** every comparison remains present, the exceeded count is exact, verdict is `FAILED`, and the process exits `1`

#### Scenario: Partial rejection wins without hiding comparisons
- **WHEN** analysis is non-exact, effective `fail_on_partial` is true, and any mix of comparisons pass or fail
- **THEN** the document retains every observed comparison and exceeded count, reports verdict `INCOMPLETE`, and exits `3`

### Requirement: JSON stream framing is machine-only and deterministic
JSON report output SHALL be UTF-8, use LF line endings, deterministic two-space indentation, and exactly one trailing LF. A non-fatal inspect or check outcome SHALL write exactly one complete JSON document to stdout and nothing to stderr. A JSON fatal outcome SHALL write exactly one complete JSON document to stderr and nothing to stdout. JSON output MUST NOT contain ANSI escapes, localized numbers, human table headings, disclaimer prose outside fields, or a Go stack trace.

#### Scenario: Repeated JSON render
- **WHEN** the same owned application result and tool version are rendered repeatedly
- **THEN** every emitted byte is identical and rendering does not mutate or reorder the application result

#### Scenario: Failed and incomplete checks remain report-first
- **WHEN** a JSON check result requires exit code `1` or `3`
- **THEN** stdout contains its complete document before the centralized process outcome is returned
- **AND** stderr remains empty

#### Scenario: JSON ignores color environment
- **WHEN** JSON is selected on a terminal with any combination of `--no-color` and `NO_COLOR`
- **THEN** the JSON bytes are identical and contain no ANSI escape sequence

### Requirement: Portable identity is independent of checkout location
Successful inspect and check documents SHALL identify in-project scenes and diagnostic sources by normalized project-relative display identity, using forward slashes and `res://` where available. Portable report fields MUST NOT expose canonical absolute filesystem paths, temporary-directory names, map iteration order, OS-specific separators, or locale-dependent formatting. Explicit input/configuration provenance MAY state how a value was selected but MUST NOT make semantic report identity depend on the checkout directory.

#### Scenario: Equivalent checkouts on different operating systems
- **WHEN** identical project contents and project-relative inputs are analyzed from different absolute checkout paths on supported operating systems
- **THEN** the semantic JSON document and its portable path fields are byte-identical after holding tool version constant

#### Scenario: Absolute input resolves inside the project
- **WHEN** a user supplies an absolute path to an in-project scene in JSON mode
- **THEN** the successful document identifies the scene by its normalized portable project-relative display path rather than the supplied canonical absolute path

### Requirement: Schema evolution preserves version-one consumers
The repository SHALL publish the authoritative report schema as `schema/deadweight.gdt.report-v1.schema.json`. Existing version-one kinds, required fields, enum values, and field meanings MUST NOT be removed, renamed, or reinterpreted. Compatible optional fields or new document kinds MAY be added during the version-one lifecycle, and consumers MUST ignore unknown object fields while rejecting unsupported schema versions. Any incompatible change MUST use a new schema version and a separately named schema file.

#### Scenario: Version-one document validation
- **WHEN** a generated inspect, check, or error document is validated against the committed version-one schema
- **THEN** validation succeeds and all required discriminators and payload fields are present

#### Scenario: Reader receives an unsupported schema version
- **WHEN** a report reader expecting version one receives another schema version
- **THEN** it rejects the document explicitly instead of interpreting it as version one

#### Scenario: Compatible optional field is introduced
- **WHEN** a later version-one producer adds an optional field without changing existing field meanings
- **THEN** existing consumers can ignore that field and continue interpreting the established version-one contract
