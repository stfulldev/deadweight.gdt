# Repository Foundation Specification

## Purpose

Defines the stable domain catalogs, error interoperability, dependency boundaries, and executable repository-skeleton gates on which later MVP analysis slices depend.

## Requirements

### Requirement: Canonical metric catalog
The system SHALL expose exactly these eight stable metric IDs and console labels in this canonical order: `nodes` / `Nodes`, `tree_depth` / `Tree depth`, `scene_instances` / `Scene instances`, `mesh_instances` / `Mesh instances`, `lights` / `Lights`, `shadow_lights` / `Shadow lights`, `external_resources` / `External resources`, and `scene_dependencies` / `Scene dependencies`. Catalog iteration SHALL be deterministic and independent of map iteration, and callers SHALL NOT be able to mutate package-owned order state.

#### Scenario: Consumer enumerates metrics
- **WHEN** a consumer requests the metric catalog repeatedly
- **THEN** every result contains exactly the eight ID/label pairs in the canonical order
- **AND** mutating one returned collection does not affect any later result

#### Scenario: Consumer validates a metric ID
- **WHEN** a consumer validates each canonical ID and an unknown ID
- **THEN** all eight canonical IDs are accepted with their exact labels
- **AND** the unknown ID is rejected and has no label

### Requirement: Metric value invariants
Every metric value SHALL use a signed 64-bit integer representation and MUST be non-negative. The domain contract SHALL reject a negative value and identify the metric whose invariant was violated rather than silently accepting it.

#### Scenario: Metric values satisfy the domain
- **WHEN** all eight values are zero or positive signed 64-bit integers
- **THEN** the metric collection is valid

#### Scenario: Metric value is negative
- **WHEN** any one of the eight values is negative
- **THEN** validation fails and identifies that metric by its stable ID

### Requirement: Stable diagnostic taxonomy
The system SHALL expose typed diagnostic severities `warning` and `error` and typed stable codes `SB1001` through `SB1007` and `SB2001` through `SB2004` with the meanings fixed by MVP specification section 11.1. `SB1xxx` codes SHALL classify as warnings, `SB2xxx` codes SHALL classify as errors, and unknown severities or codes SHALL be invalid.

#### Scenario: Diagnostic catalog is inspected
- **WHEN** a consumer enumerates or validates the diagnostic taxonomy
- **THEN** each canonical code is accepted with its specified severity
- **AND** the order is deterministic and an unknown code is rejected

#### Scenario: Diagnostic record has an inconsistent severity
- **WHEN** a diagnostic record pairs an `SB1xxx` code with `error` or an `SB2xxx` code with `warning`
- **THEN** domain validation rejects the record as inconsistent

### Requirement: Typed fatal domain errors
Fatal domain failures SHALL expose their stable diagnostic code through a typed error contract so callers can inspect the code without parsing human-readable text. TSCN parse failures SHALL use `SB2001` through that shared code contract while preserving source position and readable error text. Ordinary CLI rendering MUST NOT include a Go stack trace.

#### Scenario: Caller handles an invalid TSCN root
- **WHEN** parsing fails because the root TSCN is invalid or unsupported
- **THEN** the caller can identify the typed failure as `SB2001` without parsing its message
- **AND** the failure still includes source, line, and column context when available

#### Scenario: CLI renders a typed fatal error
- **WHEN** a typed fatal domain error reaches command orchestration
- **THEN** the CLI returns the MVP fatal exit code and renders a stable code plus human-readable message to stderr
- **AND** stderr contains no Go stack trace

### Requirement: Layered package boundaries
The executable SHALL remain a thin process adapter that delegates command orchestration and process-code mapping to the CLI layer. Metric, diagnostic, parser, preset, and budget domain packages MUST NOT print to stdout or stderr, terminate the process, or read scene files directly. Parser logic MUST remain independent of filesystem, budget, and console concerns. New architecture packages SHALL be introduced only when they contain concrete behavior required by an implementation slice.

#### Scenario: Domain logic is exercised in isolation
- **WHEN** metric, diagnostic, parser, preset, or budget behavior is tested
- **THEN** the behavior can run with in-memory values or readers and returned results or errors
- **AND** it does not require a console, a process exit, or scene filesystem access

#### Scenario: Executable delegates execution
- **WHEN** the program entry point is inspected or built
- **THEN** it only supplies process arguments and streams to CLI orchestration and exits with the returned code

### Requirement: Standalone repository skeleton gates
The repository SHALL build and pass tests and static analysis as a Go 1.24+ module on Linux, macOS, and Windows. Root help, version output, and compile-time placeholder commands SHALL remain available while later analysis slices are unfinished. The shipped binary MUST NOT require Godot, Node.js, OpenSpec, an MCP server, network access, or another runtime service.

#### Scenario: Foundation quality gates run
- **WHEN** maintainers run the repository build, test, race-test, vet, and lint gates
- **THEN** the Go module passes without generated or external runtime services

#### Scenario: Skeleton CLI runs before analyzer completion
- **WHEN** a user invokes root help, version output, or a wired placeholder command
- **THEN** the command tree is available on every supported operating system
- **AND** unfinished analyzer behavior fails clearly rather than pretending to produce an analysis result

### Requirement: Binary version output has truthful build provenance
The executable SHALL choose one user-visible version by the following precedence: a non-development value explicitly injected at link time MUST win; otherwise a semantic version recorded in Go module build metadata SHALL be used with one optional leading `v` removed; otherwise the version SHALL remain `dev`. Tagged module installation, version selection, and `--version` output MUST remain self-contained and MUST NOT require network access, Godot, Node.js, OpenSpec, an MCP server, or loose runtime metadata files after the binary is built.

#### Scenario: Explicit linker version wins
- **WHEN** a source build injects a non-development version through the documented linker variable
- **THEN** `--version` and reports use that exact injected value even if Go module metadata contains another version

#### Scenario: Tagged Go module install
- **WHEN** Go builds the command from module version `v0.1.0` without an explicit linker override
- **THEN** `--version` and reports identify the binary as `0.1.0` rather than `dev`

#### Scenario: Untagged development build
- **WHEN** neither an explicit linker version nor a semantic Go module version is present
- **THEN** `--version` and reports identify the binary as `dev`
