# Project Root Discovery Specification

## Purpose

Defines deterministic, standalone discovery of the Godot project root that later scene, resource, configuration, and application layers can share without duplicating filesystem policy.

## Requirements

### Requirement: Explicit project takes precedence
When an explicit project path is supplied, the system SHALL resolve it relative to the supplied current working directory when necessary and MUST use it instead of automatic discovery. The path SHALL be accepted only when it is either a directory containing a regular `project.godot` file or that regular `project.godot` file itself. The returned project root SHALL be an absolute cleaned directory path.

#### Scenario: Explicit project directory
- **WHEN** an explicit relative or absolute directory contains a regular `project.godot`
- **THEN** the system returns that directory as the project root without searching any scene or cwd ancestors

#### Scenario: Explicit project marker file
- **WHEN** an explicit path identifies a regular file named `project.godot`
- **THEN** the system returns the file's parent directory as the project root

#### Scenario: Invalid explicit project does not fall back
- **WHEN** an explicit path is missing, names another file, is not a directory or regular file, or names a directory without a regular `project.godot`
- **THEN** discovery fails with a typed invalid-explicit-project error
- **AND** the system does not fall back to an otherwise discoverable ancestor project

### Requirement: Filesystem scene discovery
For a scene input that does not begin with `res://`, the system SHALL resolve a relative path against the supplied current working directory, require the resulting path to be an existing regular file with the case-sensitive extension `.tscn`, and begin automatic project discovery in that file's directory. Absolute scene input SHALL follow the same validation and discovery rules without being rebased to the cwd.

#### Scenario: Relative filesystem scene
- **WHEN** a relative `.tscn` file exists beneath a Godot project and no explicit project is supplied
- **THEN** the system starts at the scene directory and returns the nearest valid project root

#### Scenario: Absolute filesystem scene
- **WHEN** an absolute `.tscn` file exists beneath a Godot project and no explicit project is supplied
- **THEN** the system starts at the scene directory and returns the nearest valid project root independently of the supplied cwd location

#### Scenario: Invalid filesystem scene
- **WHEN** the filesystem scene is missing, is not a regular file, or does not have the exact `.tscn` extension
- **THEN** discovery fails with a typed invalid-scene-input error before searching for a project root

### Requirement: Resource-path discovery starts at cwd
For a scene input beginning with `res://`, the system SHALL begin automatic discovery at the supplied current working directory because the scene filesystem path cannot be resolved before a project root is known. Root discovery SHALL NOT require the referenced resource path to exist; scene resolution and containment are separate downstream responsibilities.

#### Scenario: Resource path from inside a project
- **WHEN** the input begins with `res://`, no explicit project is supplied, and the cwd is inside a Godot project
- **THEN** the system searches from the cwd and returns the nearest valid project root

#### Scenario: Explicit project with resource path
- **WHEN** the input begins with `res://` and a valid explicit project is supplied
- **THEN** the explicit project root is returned without searching cwd ancestors

### Requirement: Nearest regular marker wins
Automatic discovery SHALL inspect the starting directory and then each parent up to the filesystem root. The first directory containing a regular file named exactly `project.godot` SHALL win. A non-regular entry with that name SHALL NOT identify a project. Traversal MUST terminate at the filesystem root and MUST NOT depend on string-prefix path tests.

#### Scenario: Nested projects
- **WHEN** both a starting directory ancestor and a higher ancestor contain regular `project.godot` files
- **THEN** the system returns the nearer ancestor

#### Scenario: Marker name is not a regular file
- **WHEN** a candidate directory contains a directory or other non-regular entry named `project.godot`
- **THEN** that candidate is not accepted as a project root
- **AND** discovery continues toward the filesystem root

#### Scenario: Filesystem root is reached
- **WHEN** no searched directory contains a regular `project.godot`
- **THEN** traversal terminates and returns a typed project-not-found error

### Requirement: Typed actionable discovery failures
Discovery failures SHALL expose a stable typed reason, the relevant input or inspected path, and an underlying filesystem cause when one exists, so callers can branch without parsing display text. Missing-project errors MUST suggest running from inside a Godot project or passing `--project`. These failures SHALL be fatal inputs to application orchestration and therefore retain the MVP fatal exit-code `2` behavior without a Go stack trace.

#### Scenario: Project cannot be found
- **WHEN** automatic traversal reaches the filesystem root without finding a valid marker
- **THEN** the typed error identifies the searched starting location
- **AND** its human-readable text suggests running inside a Godot project or passing `--project`

#### Scenario: Filesystem inspection fails
- **WHEN** discovery cannot inspect a scene, explicit project, or candidate marker for a reason other than non-existence
- **THEN** a typed filesystem error preserves the inspected path and wrapped cause
- **AND** the failure is not silently converted into project-not-found

#### Scenario: Discovery error reaches CLI orchestration
- **WHEN** a project discovery error is returned through command orchestration
- **THEN** the command exits with code `2`, writes an actionable message to stderr, and emits no Go stack trace

### Requirement: Isolated standalone discovery
Project-root discovery MUST NOT parse scene contents, read `project.godot` contents, resolve resources, print to process streams, launch Godot, use the network, or require external runtime services. The filesystem inspection boundary and cwd SHALL be controllable by tests so behavior does not depend on the developer's machine.

#### Scenario: Finder is tested in isolation
- **WHEN** discovery behavior is exercised with controlled filesystem metadata and a supplied cwd
- **THEN** every success and failure path can be reproduced without Godot, network access, console streams, or machine-specific project files
