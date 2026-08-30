## MODIFIED Requirements

### Requirement: Inspect executes the analysis application flow
The system SHALL accept exactly one scene input for `inspect`, discover or validate its project root, resolve an optional configuration, securely resolve the scene, recursively analyze it, and return a report-ready result containing project, scene, configuration-source, analysis, coverage, and diagnostic evidence. The scene input SHALL support absolute paths, working-directory-relative paths, `res://` paths, and canonical `uid://` values when the project is supplied or discoverable from the working directory. A missing implicit configuration SHALL be accepted, `fail_on_partial` SHALL NOT alter the inspect outcome, and a non-fatal partial analysis SHALL succeed.

#### Scenario: Inspect an absolute scene without configuration
- **WHEN** a user runs `inspect` with one absolute in-project `.tscn` path and no implicit configuration exists
- **THEN** the application discovers the containing project, analyzes the scene, and returns a successful report-ready result

#### Scenario: Inspect a relative or resource scene
- **WHEN** a user runs `inspect` with one working-directory-relative or `res://` scene and the project can be discovered or is supplied explicitly
- **THEN** the application resolves the scene through the project path boundary before analysis

#### Scenario: Inspect a UID scene
- **WHEN** a user runs `inspect` with a resolvable `uid://` scene and the project can be discovered from cwd or is supplied explicitly
- **THEN** the application builds the project UID index, resolves the canonical scene, and executes the ordinary analysis flow

#### Scenario: Inspect a partial analysis
- **WHEN** recursive analysis returns partial evidence without a fatal error and configuration sets `fail_on_partial` to true
- **THEN** the command preserves the partial report and exits successfully

#### Scenario: Inspect rejects an invalid argument count
- **WHEN** `inspect` receives zero or more than one positional scene argument
- **THEN** the command returns an actionable usage failure without invoking analysis

### Requirement: Tree executes one focused analysis flow
The `tree` command SHALL accept exactly one scene input, reuse the same project discovery, optional strict configuration loading, secure scene resolution, and recursive analysis boundaries as `inspect`, and return a report-ready dependency-tree result without policy or budget evaluation. It SHALL support absolute, working-directory-relative, `res://`, and canonical `uid://` scene inputs. A successful partial analysis SHALL remain reportable with exit code `0`, while usage and fatal analysis failures SHALL retain exit code `2`.

#### Scenario: Tree analyzes one resource scene
- **WHEN** a user runs `tree res://levels/city.tscn`
- **THEN** the command performs project/configuration resolution and recursive analysis once before tree presentation
- **AND** it does not perform budget or policy evaluation

#### Scenario: Tree analyzes one UID scene
- **WHEN** a user runs `tree uid://example` for a uniquely mapped project scene
- **THEN** the command resolves the same canonical scene and dependency graph that a path-based invocation would use

#### Scenario: Invalid tree argument count
- **WHEN** `tree` receives zero or more than one positional scene argument
- **THEN** it returns actionable usage guidance before invoking the application service

#### Scenario: Partial tree result
- **WHEN** recursive analysis returns successful partial dependency evidence
- **THEN** the command presents the tree and diagnostics and exits `0`
