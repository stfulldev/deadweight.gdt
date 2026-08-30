## MODIFIED Requirements

### Requirement: Root scene inputs resolve to canonical project files
Root scene resolution SHALL accept absolute host paths, paths relative to a supplied absolute current working directory, `res://` paths relative to the canonical project root, and canonical `uid://` values resolved by the selected project's UID index. A successful root scene MUST identify an existing regular file with the case-sensitive `.tscn` extension and MUST remain inside the canonical project root after symlink evaluation. Invalid, missing, ambiguous, stale, unsupported, or unsafe root-scene input SHALL return a typed fatal error rather than an unresolved resource result.

#### Scenario: Absolute root scene inside project
- **WHEN** an absolute host path identifies an existing regular `.tscn` inside the canonical project root
- **THEN** the system returns its resolved path identities without rebasing it to the cwd

#### Scenario: Relative root scene
- **WHEN** a cwd-relative path identifies an existing regular `.tscn` inside the canonical project root
- **THEN** the system resolves it from the supplied cwd and returns its canonical in-project identity

#### Scenario: Resource root scene
- **WHEN** a `res://` root-scene input identifies an existing regular `.tscn` beneath the canonical project root
- **THEN** the system resolves it from the project root independently of the cwd location

#### Scenario: UID root scene
- **WHEN** a canonical `uid://` root uniquely maps to an existing supported `.tscn` inside the selected project
- **THEN** the system returns that scene's canonical and `res://` identities while preserving the original UID

#### Scenario: Invalid root scene
- **WHEN** a root-scene input is empty, missing, ambiguous, stale, non-regular, outside the project, inaccessible, unsupported, or lacks the exact `.tscn` extension
- **THEN** resolution fails with a stable typed reason, relevant evidence, and wrapped filesystem cause when one exists

### Requirement: Resource references use deterministic bases
Resource resolution SHALL resolve `res://` references from the canonical project root, relative and parent-relative references from the declaring scene's directory, absolute host references without rebasing, and `uid://` references through the selected project UID index. When a declaration carries both UID and path evidence, a unique secure UID mapping SHALL take precedence; if the UID is unknown or malformed, the ordinary path SHALL remain the deterministic fallback. Ambiguous, stale, unsupported, or unsafe UID evidence MUST remain visible even when a usable path fallback exists. The declaring scene MUST be a canonical path inside the project. An existing regular target inside the project SHALL resolve regardless of its file extension so later layers can classify scenes, imported resources, and ordinary external resources without path-policy duplication.

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

#### Scenario: UID resource
- **WHEN** a canonical UID uniquely maps to an existing regular project resource
- **THEN** resolution returns that canonical resource independently of the declaring scene directory

#### Scenario: Unknown UID falls back to declared path
- **WHEN** an external-resource declaration has an unknown UID and a usable ordinary path
- **THEN** the ordinary path is resolved through the existing base and containment rules and the UID miss remains deterministic evidence

### Requirement: Unresolved resources carry stable classifications
Resource resolution SHALL represent expected nonfatal failures as typed unresolved results with the original raw value, stable reason, relevant UID evidence or candidate path when available, and wrapped filesystem cause when one exists. At minimum the reasons MUST distinguish empty input, missing, malformed, ambiguous, stale, unsupported-version, or unsafe UID evidence, `user://`, unsupported schemes or non-regular targets, missing targets, outside-project targets, inaccessible filesystem state, and invalid declaring scenes. Callers MUST be able to branch on these reasons without parsing display text.

#### Scenario: Unsupported Godot schemes
- **WHEN** a resource reference is empty or begins with `user://`
- **THEN** the unresolved result distinguishes the corresponding empty or user-data reason without target filesystem access

#### Scenario: Unresolvable UID
- **WHEN** a UID-only reference is missing, malformed, ambiguous, stale, unsupported, or maps unsafely
- **THEN** the unresolved result preserves the original UID and its specific typed reason without guessing a resource path

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
