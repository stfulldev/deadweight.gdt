## Purpose

Defines the executable application orchestration and command contract that connect scene analysis, policy evaluation, built-in presets, and deterministic process outcomes without requiring Godot.

## ADDED Requirements

### Requirement: Inspect executes the analysis application flow
The system SHALL accept exactly one scene input for `inspect`, discover or validate its project root, resolve an optional configuration, securely resolve the scene, recursively analyze it, and return a report-ready result containing project, scene, configuration-source, analysis, coverage, and diagnostic evidence. The scene input SHALL support absolute paths, working-directory-relative paths, and `res://` paths. A missing implicit configuration SHALL be accepted, `fail_on_partial` SHALL NOT alter the inspect outcome, and a non-fatal partial analysis SHALL succeed.

#### Scenario: Inspect an absolute scene without configuration
- **WHEN** a user runs `inspect` with one absolute in-project `.tscn` path and no implicit configuration exists
- **THEN** the application discovers the containing project, analyzes the scene, and returns a successful report-ready result

#### Scenario: Inspect a relative or resource scene
- **WHEN** a user runs `inspect` with one working-directory-relative or `res://` scene and the project can be discovered or is supplied explicitly
- **THEN** the application resolves the scene through the project path boundary before analysis

#### Scenario: Inspect a partial analysis
- **WHEN** recursive analysis returns partial evidence without a fatal error and configuration sets `fail_on_partial` to true
- **THEN** the command preserves the partial report and exits successfully

#### Scenario: Inspect rejects an invalid argument count
- **WHEN** `inspect` receives zero or more than one positional scene argument
- **THEN** the command returns an actionable usage failure without invoking analysis

### Requirement: Check executes policy-aware budget evaluation
The system SHALL perform the same project, optional-configuration, scene-resolution, and analysis flow as `inspect`, then resolve the effective policy with CLI selectors and ordered budget overrides, resolve the effective partial policy, and evaluate all effective budgets into one report-ready check result. A CLI preset or profile SHALL replace the configured selector, project budgets SHALL override the selected base, and repeated CLI budget overrides SHALL be applied in order with the last duplicate metric winning. The system MUST reject a check whose effective policy contains no budget.

#### Scenario: CLI selector and budgets override configuration
- **WHEN** configuration selects one base and supplies project budgets while `check` supplies another CLI selector and repeated CLI budgets
- **THEN** the CLI selector replaces the configured selector, project budgets override that selected base, and the ordered CLI budgets are applied last

#### Scenario: Empty effective policy is rejected
- **WHEN** no preset, profile, project budget, or CLI budget produces an effective limit
- **THEN** `check` returns a fatal usage/configuration outcome with guidance to select or provide a budget

#### Scenario: Partial rejection takes verdict priority
- **WHEN** analysis is partial, effective `fail_on_partial` is true, and one or more observed budgets are exceeded
- **THEN** the report retains all observed comparisons and the command outcome is incomplete rather than failed

#### Scenario: Partial analysis is explicitly allowed
- **WHEN** analysis is partial and `--allow-partial` overrides configured `fail_on_partial=true`
- **THEN** budget evaluation uses `fail_on_partial=false` and the command outcome is passed or failed according to observed comparisons

### Requirement: Built-in preset flows are project independent
The system SHALL list built-in presets in product order and show one built-in preset by ID without performing working-directory, project, configuration, scene-resolution, or analysis operations. Unknown preset IDs MUST return a fatal outcome that lists the available IDs.

#### Scenario: List presets outside a project
- **WHEN** a user runs `presets` from a directory without `project.godot` or configuration
- **THEN** all built-in presets are returned in frozen product order

#### Scenario: Show an unknown preset
- **WHEN** a user runs `presets show` with an unknown ID
- **THEN** the command fails and names the available built-in preset IDs

### Requirement: CLI syntax matches the frozen command contract
The root command SHALL support `--project PATH`, `--config PATH`, `--no-color`, `--version`, and standard help. `check` SHALL support mutually exclusive `--preset ID` and `--profile ID`, repeatable `--budget METRIC=LIMIT`, and mutually exclusive `--fail-on-partial` and `--allow-partial`. Each scene command SHALL accept exactly one scene, `presets` SHALL accept no positional arguments, and `presets show` SHALL accept exactly one ID. Usage failures MUST be actionable and MUST NOT invoke the application flow.

#### Scenario: Conflicting selector flags
- **WHEN** `check` receives both `--preset` and `--profile`
- **THEN** the command returns a usage failure identifying the mutually exclusive flags before application work begins

#### Scenario: Conflicting partial flags
- **WHEN** `check` receives both `--fail-on-partial` and `--allow-partial`
- **THEN** the command returns a usage failure identifying the mutually exclusive flags before application work begins

#### Scenario: Repeated budget flags preserve order
- **WHEN** `check` receives multiple `--budget` values including duplicate metrics
- **THEN** the application request receives those values in command-line order

#### Scenario: Global project and config flags reach scene flows
- **WHEN** a scene command supplies explicit project and configuration paths as global flags
- **THEN** the application request receives both paths and uses them for discovery and loading

### Requirement: Command composition is injectable and Godot independent
The executable SHALL keep process setup in a thin composition root and place domain orchestration behind an injected application boundary. Command handlers SHALL only translate parsed command values into application requests, present returned models, and select the centralized outcome. Application and command tests MUST be able to inject services and streams and MUST NOT locate, launch, or require a Godot executable.

#### Scenario: Command test uses an injected application
- **WHEN** a command test supplies a fake application and in-memory output streams
- **THEN** the command can validate parsed requests, rendering, and exit outcomes without filesystem analysis or Godot

#### Scenario: Default application performs static analysis only
- **WHEN** the shipped CLI executes an inspect or check flow
- **THEN** it uses Go filesystem, parser, analysis, policy, and budget services without looking for or launching Godot

### Requirement: Application outcomes map centrally to process exit codes
The command runtime SHALL map fatal and usage errors to exit code `2`, rejected partial analysis to `3`, exceeded budgets to `1`, and every other successful result to `0`. This priority SHALL be applied centrally so `main.go` and individual Cobra handlers do not duplicate exit-code policy. Non-fatal results SHALL remain available for presentation before their exit code is returned.

#### Scenario: Failed budget returns one
- **WHEN** a complete or allowed-partial check report has one or more exceeded budgets
- **THEN** the report is presented and process exit code `1` is returned

#### Scenario: Rejected partial returns three
- **WHEN** a check report has incomplete status because partial analysis is rejected
- **THEN** the report is presented and process exit code `3` is returned

#### Scenario: Fatal error has highest priority
- **WHEN** discovery, configuration, resolution, parsing, analysis, policy resolution, or budget evaluation returns a fatal error
- **THEN** the process returns exit code `2` instead of any non-fatal verdict
