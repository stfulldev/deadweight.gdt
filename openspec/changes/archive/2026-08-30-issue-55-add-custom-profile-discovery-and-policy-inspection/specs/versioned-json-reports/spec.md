## ADDED Requirements

### Requirement: Profile-list JSON is a compatible version-one document kind
The version-one report schema SHALL accept kind `profiles` with the standard schema/tool envelope, selected portable configuration context, and exactly one profile-list payload. Each profile entry SHALL contain its ID, optional declared parent, and effective name and description, ordered by canonical profile ID. The document MUST NOT contain scene-analysis, policy-evaluation, dependency-tree, diff, fatal-error, or profile-detail payloads.

#### Scenario: Populated JSON list
- **WHEN** a selected configuration declares multiple custom profiles
- **THEN** kind `profiles` contains one canonically ordered entry per declaration and validates against the committed version-one report schema

#### Scenario: Empty JSON list
- **WHEN** the selected configuration declares no custom profiles
- **THEN** kind `profiles` contains an empty profile array rather than null or an omitted payload

### Requirement: Effective-profile JSON is a compatible version-one document kind
The version-one report schema SHALL accept kind `profile` with the standard schema/tool envelope, selected portable configuration context, and exactly one profile-detail payload. That payload SHALL contain the selected custom ID, ordered parent-chain layers, all effective metadata values and sources, all effective budgets and sources in frozen metric order, and effective `fail_on_partial` with its source. Each source SHALL use a stable layer kind and include an ID only for built-in preset or custom-profile layers. The document MUST NOT contain payloads owned by other kinds.

#### Scenario: Full effective JSON
- **WHEN** a custom profile inherits and overrides effective values
- **THEN** kind `profile` retains every value, ordered chain layer, and source and validates against schema version one

#### Scenario: Absent optional budget
- **WHEN** an effective profile has no value for one budget metric
- **THEN** the profile budget array omits that metric rather than emitting a zero value or provenance

### Requirement: Custom-profile JSON remains deterministic and additive
Profile-list and effective-profile JSON SHALL be deterministic UTF-8 with two-space indentation, no ANSI escapes, and exactly one trailing LF. Portable fields MUST use forward-slash project-relative configuration paths and MUST NOT expose canonical checkout prefixes. Adding the two kinds MUST NOT remove, rename, or reinterpret established version-one kinds or fields, so earlier version-one documents continue to validate.

#### Scenario: Equivalent checkout prefixes
- **WHEN** equivalent profile results originate from different checkout directory prefixes
- **THEN** their JSON documents are byte-identical after holding tool version constant

#### Scenario: Earlier version-one report
- **WHEN** a previously valid version-one inspect, tree, check, diff, or error report is validated against the evolved schema
- **THEN** it remains valid with unchanged field meanings

