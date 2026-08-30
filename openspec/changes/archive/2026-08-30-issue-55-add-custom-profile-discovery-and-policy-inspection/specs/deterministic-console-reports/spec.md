## ADDED Requirements

### Requirement: Custom-profile list text is deterministic
The text presentation for `profiles` SHALL identify the selected configuration, render every custom profile exactly once in canonical profile-ID order, retain declared parent information, and end with exactly one LF. An empty declaration map SHALL produce an explicit empty-state message. Output MUST contain no ANSI escapes when color is disabled and MUST NOT expose checkout-dependent canonical paths beyond the selected user-facing configuration context.

#### Scenario: Reordered configuration members
- **WHEN** equivalent configurations declare the same profiles in different JSON member orders
- **THEN** their rendered profile-list bodies are byte-identical after holding the selected portable configuration context and tool version constant

#### Scenario: Empty profile map
- **WHEN** a valid selected configuration declares no custom profiles
- **THEN** text output clearly states that no custom profiles are declared

### Requirement: Effective profile text is complete and canonically ordered
The text presentation for `profiles show <id>` SHALL render the selected ID, ordered inheritance chain, all eight metadata fields with their sources, every effective budget in frozen metric order with its source, and `fail_on_partial` with its source. Empty strings and zero values MUST remain explicit rather than being omitted, the document SHALL end with exactly one LF, and repeated rendering MUST be byte-identical without mutating the result.

#### Scenario: Mixed provenance profile
- **WHEN** an effective profile contains built-in, custom-profile, default, and project-sourced values
- **THEN** text output labels every value with the corresponding source and preserves canonical field and metric order

#### Scenario: Repeated render
- **WHEN** one owned effective-profile result is rendered more than once
- **THEN** every output is byte-identical and the result remains unchanged

