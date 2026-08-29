# Secure Path Resolution Specification

## Purpose

Defines secure and deterministic conversion of scene and resource references into canonical in-project filesystem identities, normalized Godot display paths, and typed unresolved evidence.

## Requirements

### Requirement: Resolved paths preserve three identities
Every successful resolution SHALL retain the canonical absolute filesystem path used for I/O and cache identity, a normalized project-relative display path beginning with `res://`, and the original raw input used for diagnostics. Display paths MUST use forward slashes on every host platform, and the project root itself SHALL display as `res://`.

#### Scenario: Nested path identities
- **WHEN** an existing file beneath the canonical project root is resolved from a raw value containing platform-native separators or cleanable segments
- **THEN** the result retains that raw value unchanged
- **AND** its filesystem identity is canonical and absolute
- **AND** its display identity is a cleaned forward-slash `res://` path

#### Scenario: Project root display
- **WHEN** the canonical project root is converted to a display path
- **THEN** the result is exactly `res://`

### Requirement: Root scene inputs resolve to canonical project files
Root scene resolution SHALL accept absolute host paths, paths relative to a supplied absolute current working directory, and `res://` paths relative to the canonical project root. A successful root scene MUST identify an existing regular file with the case-sensitive `.tscn` extension and MUST remain inside the canonical project root after symlink evaluation. Invalid root-scene input SHALL return a typed fatal error rather than an unresolved resource result.

#### Scenario: Absolute root scene inside project
- **WHEN** an absolute host path identifies an existing regular `.tscn` inside the canonical project root
- **THEN** the system returns its resolved path identities without rebasing it to the cwd

#### Scenario: Relative root scene
- **WHEN** a cwd-relative path identifies an existing regular `.tscn` inside the canonical project root
- **THEN** the system resolves it from the supplied cwd and returns its canonical in-project identity

#### Scenario: Resource root scene
- **WHEN** a `res://` root-scene input identifies an existing regular `.tscn` beneath the canonical project root
- **THEN** the system resolves it from the project root independently of the cwd location

#### Scenario: Invalid root scene
- **WHEN** a root-scene input is empty, missing, non-regular, outside the project, inaccessible, or lacks the exact `.tscn` extension
- **THEN** resolution fails with a stable typed reason, relevant path, and wrapped filesystem cause when one exists

### Requirement: Resource references use deterministic bases
Resource resolution SHALL resolve `res://` references from the canonical project root, relative and parent-relative references from the declaring scene's directory, and absolute host references without rebasing. The declaring scene MUST be a canonical path inside the project. An existing regular target inside the project SHALL resolve regardless of its file extension so later layers can classify scenes, imported resources, and ordinary external resources without path-policy duplication.

#### Scenario: Project-relative resource
- **WHEN** a `res://assets/material.tres` reference identifies an existing regular project file
- **THEN** it resolves relative to the project root with canonical and display identities

#### Scenario: Declaring-scene-relative resource
- **WHEN** a relative reference identifies an existing regular file from the declaring scene's directory
- **THEN** it resolves from that directory rather than the process cwd or project root

#### Scenario: Parent-relative resource remains inside project
- **WHEN** a reference containing `../` cleans to an existing regular target that remains inside the project
- **THEN** it resolves successfully

#### Scenario: Absolute resource inside project
- **WHEN** an absolute host reference identifies an existing regular file inside the canonical project root
- **THEN** it resolves successfully without rebasing

### Requirement: Lexical and symlink containment is mandatory
Every candidate path SHALL pass segment-aware containment against the canonical project root before it can be returned for downstream I/O. Existing symlinks MUST be evaluated, and both direct targets and paths whose nearest existing ancestor traverses a symlink MUST be rejected when their canonical destination escapes the project. Containment MUST NOT use string-prefix tests, and the resolver MUST NOT open target contents while classifying a rejected path.

#### Scenario: Lexical escape
- **WHEN** a relative, `res://`, or absolute candidate cleans to a location outside the canonical project root
- **THEN** it is rejected with an outside-project reason before target contents are opened

#### Scenario: Existing symlink escape
- **WHEN** an apparently in-project path resolves through an existing symlink to a target outside the canonical project root
- **THEN** it is rejected with an outside-project reason and the outside target is not returned

#### Scenario: Missing target below an escaping symlink
- **WHEN** a missing candidate has an existing ancestor symlink whose canonical destination is outside the project
- **THEN** it is classified as outside-project rather than as an ordinary in-project missing target

#### Scenario: Segment boundary collision
- **WHEN** a candidate is under a sibling path whose text merely begins with the project-root text
- **THEN** it is rejected as outside the project

### Requirement: Unresolved resources carry stable classifications
Resource resolution SHALL represent expected nonfatal failures as typed unresolved results with the original raw value, stable reason, relevant candidate path when available, and wrapped filesystem cause when one exists. At minimum the reasons MUST distinguish empty input, `uid://`, `user://`, unsupported schemes or non-regular targets, missing targets, outside-project targets, inaccessible filesystem state, and invalid declaring scenes. Callers MUST be able to branch on these reasons without parsing display text.

#### Scenario: Unsupported Godot schemes
- **WHEN** a resource reference is empty, begins with `uid://`, or begins with `user://`
- **THEN** the unresolved result distinguishes the corresponding empty, UID-only, or user-data reason without filesystem access

#### Scenario: Unknown scheme or non-regular target
- **WHEN** a reference uses another URI-like scheme or resolves to a directory or other non-regular entry
- **THEN** the result is unresolved with an unsupported-target reason

#### Scenario: Missing in-project target
- **WHEN** a lexically and canonically in-project resource target does not exist
- **THEN** the result is unresolved with a missing-target reason and retains the candidate path

#### Scenario: Filesystem metadata is inaccessible
- **WHEN** the resolver cannot inspect a candidate or evaluate its existing ancestors for a reason other than non-existence
- **THEN** the result is unresolved with a filesystem reason and preserves the wrapped cause

#### Scenario: Declaring scene is invalid
- **WHEN** a relative resource is supplied with a declaring scene that is non-canonical or outside the project
- **THEN** the result is unresolved with an invalid-declaring-scene reason

### Requirement: Resolution never repairs case or depends on external runtimes
Resolution MUST use host filesystem case semantics and MUST NOT perform case-insensitive lookup or filename recovery. It MUST NOT parse target contents, launch Godot, access the network, print to process streams, or require external runtime services. Filesystem metadata and symlink evaluation boundaries SHALL be controllable by tests so failure behavior is reproducible.

#### Scenario: Wrong-case reference on a case-sensitive filesystem
- **WHEN** a reference differs from an existing path only by letter case on a case-sensitive host
- **THEN** it remains unresolved as missing and is not repaired by directory scanning

#### Scenario: Isolated resolution
- **WHEN** resolution is exercised with controlled filesystem and symlink operations
- **THEN** every success, escape, missing, and inaccessible path can be reproduced without Godot, network access, parser execution, console streams, or machine-specific project files
