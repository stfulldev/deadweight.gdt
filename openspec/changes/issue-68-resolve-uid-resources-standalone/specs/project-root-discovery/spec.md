## MODIFIED Requirements

### Requirement: Filesystem scene discovery
For a scene input that begins with neither `res://` nor `uid://`, the system SHALL resolve a relative path against the supplied current working directory, require the resulting path to be an existing regular file with the case-sensitive extension `.tscn`, and begin automatic project discovery in that file's directory. Absolute scene input SHALL follow the same validation and discovery rules without being rebased to the cwd.

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
For a scene input beginning with `res://` or `uid://`, the system SHALL begin automatic discovery at the supplied current working directory because the scene filesystem path cannot be resolved before a project root and, for UID input, a project UID index are known. Root discovery SHALL NOT require the referenced resource path or UID to exist; scene resolution and containment are separate downstream responsibilities.

#### Scenario: Resource path from inside a project
- **WHEN** the input begins with `res://`, no explicit project is supplied, and the cwd is inside a Godot project
- **THEN** the system searches from the cwd and returns the nearest valid project root

#### Scenario: UID path from inside a project
- **WHEN** the input begins with `uid://`, no explicit project is supplied, and the cwd is inside a Godot project
- **THEN** the system searches from the cwd and returns the nearest valid project root without scanning UID metadata during discovery

#### Scenario: Explicit project with resource path
- **WHEN** the input begins with `res://` or `uid://` and a valid explicit project is supplied
- **THEN** the explicit project root is returned without searching cwd ancestors
