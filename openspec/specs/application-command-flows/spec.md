# application-command-flows Specification

## Purpose

Defines the executable application orchestration and command contract that connect scene analysis, policy evaluation, built-in presets, and deterministic process outcomes without requiring Godot.

## Requirements

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

### Requirement: Scene presentation format is selected without changing domain flow
The `inspect` and `check` command boundaries SHALL carry one validated presentation format, `text` or `json`, to centralized report and fatal-error presentation. Format selection MUST remain outside project discovery, parsing, recursive analysis, policy resolution, and budget evaluation. The application service interfaces and report-ready domain results SHALL remain independent of JSON encoding, ANSI/color state, and process streams.

#### Scenario: Inspect formats one application result
- **WHEN** identical inspect arguments are executed once with text format and once with JSON format
- **THEN** both invocations make an equivalent application request and receive equivalent analysis evidence
- **AND** only centralized presentation bytes differ

#### Scenario: Check exit outcome is format independent
- **WHEN** identical check arguments produce passed, failed, incomplete, or fatal outcomes in text and JSON formats
- **THEN** both formats return the same centralized process exit code for each outcome

#### Scenario: Injected command test selects JSON
- **WHEN** a command test supplies an injected application, in-memory streams, and `--format json`
- **THEN** it can verify request translation, JSON presentation, stream routing, and exit outcome without filesystem analysis, Godot, Node.js, OpenSpec, or network access

### Requirement: Inspect validates top-contributor presentation selectors
The `inspect` command SHALL accept `--metric METRIC` together with `--top LIMIT` to request an opt-in top-contributors projection. Both flags MUST be supplied together, `LIMIT` MUST be a positive signed 64-bit integer, and `METRIC` MUST be one of `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, or `shadow_lights`. Missing pairs, non-positive or overflowing limits, unknown metrics, and unique-union metric selections MUST fail as actionable usage errors before project discovery or application analysis. The selectors SHALL remain presentation inputs and MUST NOT change the application request, analysis result, or exit outcome.

#### Scenario: Valid explicit selection
- **WHEN** inspect receives `--metric nodes --top 5`
- **THEN** it performs the ordinary inspect application flow once and supplies the resulting contribution evidence to the selected presentation

#### Scenario: Missing selector pair
- **WHEN** inspect receives only `--metric` or only `--top`
- **THEN** it exits `2` with actionable usage guidance before invoking the application service

#### Scenario: Unsupported unique metric
- **WHEN** inspect receives `--metric external_resources --top 5`
- **THEN** it exits `2` before analysis and explains that unique-union metrics have no additive top-owner ranking

#### Scenario: Check remains unchanged
- **WHEN** a user invokes `check`
- **THEN** top-contributor selectors are unavailable and check request, report, and exit semantics remain unchanged

### Requirement: Tree executes one focused analysis flow
The `tree` command SHALL accept exactly one scene input, reuse the same project discovery, optional strict configuration loading, secure scene resolution, and recursive analysis boundaries as `inspect`, and return a report-ready dependency-tree result without policy or budget evaluation. It SHALL support absolute, working-directory-relative, and `res://` scene inputs. A successful partial analysis SHALL remain reportable with exit code `0`, while usage and fatal analysis failures SHALL retain exit code `2`.

#### Scenario: Tree analyzes one resource scene
- **WHEN** a user runs `tree res://levels/city.tscn`
- **THEN** the command performs project/configuration resolution and recursive analysis once before tree presentation
- **AND** it does not perform budget or policy evaluation

#### Scenario: Invalid tree argument count
- **WHEN** `tree` receives zero or more than one positional scene argument
- **THEN** it returns actionable usage guidance before invoking the application service

#### Scenario: Partial tree result
- **WHEN** recursive analysis returns successful partial dependency evidence
- **THEN** the command presents the tree and diagnostics and exits `0`

### Requirement: Tree format selection changes only presentation
The `tree` command SHALL accept `--format text` and `--format json`, use text by default, and reject every other format before application analysis. Format selection MUST NOT alter the application request, recursive result, dependency projection semantics, diagnostics, or process exit outcome. Command tests SHALL be able to inject the tree application flow and in-memory streams without filesystem analysis, Godot, Node.js, OpenSpec, network access, or another runtime service.

#### Scenario: Text and JSON share one domain result
- **WHEN** equivalent tree arguments are executed with text and JSON formats
- **THEN** both invocations send equivalent application requests and preserve equivalent graph evidence
- **AND** only centralized presentation bytes differ

#### Scenario: Unknown tree format
- **WHEN** `tree` receives an unsupported format value
- **THEN** it exits `2` with actionable usage guidance before invoking the application service

#### Scenario: Injected tree command
- **WHEN** a command test supplies a fake tree application and in-memory streams
- **THEN** it can verify request translation, presentation selection, stream routing, and exit outcome without external runtime effects

### Requirement: Diff executes a project-independent baseline flow
The `diff` command SHALL accept exactly two positional JSON report paths, read and compare them through an injected application boundary, and avoid working-directory discovery, project discovery, configuration loading, scene resolution, analysis, policy resolution, and Godot. It SHALL accept `--format text|json`, repeatable `--fail-on-increase METRIC` for any frozen metric, and `--fail-on-reliability`; reject invalid format, argument count, duplicate selected metrics, and unknown metrics before reading files; render every non-fatal result before centrally returning exit `0`, `1`, or `3`; and map every read, decode, compatibility, or validation failure to exit `2` using the selected error representation.

#### Scenario: Compare two baselines offline
- **WHEN** a user runs `diff before.json after.json`
- **THEN** the application reads exactly those two files, performs no scene flow, and renders deterministic text with exit `0` unless an explicit policy triggers

#### Scenario: JSON diff selection
- **WHEN** the same valid comparison uses `--format json`
- **THEN** only presentation bytes change and the semantic result and exit outcome remain equivalent

#### Scenario: Invalid selected metric
- **WHEN** `--fail-on-increase` names an unknown or duplicate metric
- **THEN** diff returns actionable exit `2` guidance before reading either baseline

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
