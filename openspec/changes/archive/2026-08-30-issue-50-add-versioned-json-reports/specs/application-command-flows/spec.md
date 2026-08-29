## ADDED Requirements

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
