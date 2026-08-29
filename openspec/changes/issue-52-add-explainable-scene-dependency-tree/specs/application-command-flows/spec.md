## ADDED Requirements

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
