## ADDED Requirements

### Requirement: Custom-profile commands execute a project-context flow
The `profiles` and `profiles show <id>` commands SHALL discover or validate a Godot project from the working directory or `--project`, strictly load the explicit `--config` or implicit project configuration, and invoke custom-profile discovery or inspection without resolving a scene, parsing TSCN, recursively analyzing resources, evaluating budgets, or launching Godot. A missing project, absent configuration, invalid configuration, or invalid profile graph MUST be fatal with exit code `2`.

#### Scenario: List profiles from inside a project
- **WHEN** a user runs `profiles` from a project descendant with an implicit configuration
- **THEN** the application discovers that project, loads its configuration, and returns the custom-profile list without scene analysis

#### Scenario: Explicit project and configuration
- **WHEN** `profiles show shipping` receives valid global `--project` and `--config` paths
- **THEN** the application uses those selections and returns the effective `shipping` profile

#### Scenario: Missing project context
- **WHEN** a user runs a custom-profile command outside a Godot project without `--project`
- **THEN** the command exits `2` with the existing actionable project-discovery guidance

### Requirement: Custom-profile CLI syntax and format are validated before application work
The `profiles` command SHALL accept no positional arguments and one `--format text|json` option defaulting to text. Its `show` subcommand SHALL accept exactly one custom-profile ID and its own `--format text|json` option defaulting to text. Unsupported formats and invalid argument counts MUST fail with actionable usage guidance before project or configuration access, and format selection MUST change only presentation bytes rather than application semantics or exit outcomes.

#### Scenario: JSON profile listing
- **WHEN** `profiles --format json` receives valid context
- **THEN** it invokes the same application list flow as text output and renders only the selected presentation differently

#### Scenario: Invalid show arguments
- **WHEN** `profiles show` receives zero or more than one ID
- **THEN** it exits `2` before invoking the application service

#### Scenario: Unsupported format
- **WHEN** either custom-profile command receives a format other than `text` or `json`
- **THEN** it exits `2` before project discovery

