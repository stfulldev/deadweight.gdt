## ADDED Requirements

### Requirement: Inspect can append an opt-in top-contributors section
When valid top-contributor selectors are present, the inspect text report SHALL append a section after the existing analysis evidence and warnings that names the selected metric and requested limit and renders the deterministically ranked rows with value, occurrence count, reliability, portable scene identity, and mount context. The section SHALL use the selected metric's reliability marker semantics, SHALL remain understandable without color, and MUST NOT alter any existing report bytes before the appended section. When selectors are absent, the complete default text report MUST remain byte-compatible.

#### Scenario: Default inspect remains unchanged
- **WHEN** inspect is rendered without top-contributor selectors
- **THEN** its text bytes match the established report for the same application result

#### Scenario: Partial top row
- **WHEN** a selected contribution is lower-bound or approximate
- **THEN** its value and textual reliability identify that uncertainty without relying on ANSI color

#### Scenario: Stable top ordering
- **WHEN** tied contribution rows are rendered repeatedly from equivalent checkouts
- **THEN** the appended section is byte-identical and uses only portable identities and deterministic context order

