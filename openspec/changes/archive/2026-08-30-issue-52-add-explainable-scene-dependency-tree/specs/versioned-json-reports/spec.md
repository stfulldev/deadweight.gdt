## ADDED Requirements

### Requirement: Tree JSON is a compatible version-one document kind
The version-one report schema SHALL accept kind `tree` as a compatible new document kind without removing, renaming, or reinterpreting any established inspect, check, or error field. A successful tree document SHALL retain the portable scene and configuration identity plus the complete existing analysis payload, and SHALL add one required `dependency_tree` payload. Tree JSON SHALL preserve text-equivalent projection semantics while remaining machine-only, deterministic, UTF-8, two-space indented, ANSI-free, and framed by exactly one trailing LF.

#### Scenario: Complete tree document
- **WHEN** complete recursive analysis is rendered by `tree --format json`
- **THEN** the document uses `schema_version: 1`, kind `tree`, complete analysis evidence, and one schema-valid dependency-tree payload

#### Scenario: Existing version-one consumer
- **WHEN** an existing consumer understands inspect, check, and error but ignores the newly introduced tree kind
- **THEN** every established version-one field and meaning remains unchanged

#### Scenario: Failed tree analysis
- **WHEN** tree analysis fails fatally after JSON format is selected
- **THEN** stderr receives the existing kind `error` document and stdout remains empty

### Requirement: Tree JSON entries form one portable deterministic preorder
The `dependency_tree` payload SHALL contain one portable root identity and an ordered flat preorder of edge entries. Every entry SHALL contain positive signed-64-bit depth and occurrence values, portable source and target identities, edge kind, resolved state, row reliability, and explicit back-reference state. Unresolved entries SHALL additionally retain stable classification and available portable resource, raw-target, and resolution-reason evidence. Entry depth and order SHALL reproduce the text projection, and unique graph edges MUST NOT be duplicated or omitted.

#### Scenario: Repeated edge entry
- **WHEN** a compacted graph edge has 100 occurrences
- **THEN** JSON contains one entry with `occurrences: 100` rather than 100 duplicated entries

#### Scenario: Diamond back-reference entry
- **WHEN** a later preorder edge reaches an already expanded target
- **THEN** that entry sets its back-reference state and the target's descendants are absent from that later branch

#### Scenario: Unresolved entry evidence
- **WHEN** an imported target cannot resolve to a graph node
- **THEN** its JSON entry retains `resolved: false`, lower-bound reliability, `imported_scene` classification, and available portable source context

### Requirement: Tree JSON is checkout-independent and clone-safe
All tree JSON identities SHALL be normalized project-relative resource paths or stable retained unresolved identities. Canonical absolute paths, checkout directory names, and backslashes MUST NOT appear in portable fields. Equivalent results from different checkout prefixes and supported operating systems SHALL produce byte-identical documents after holding tool version constant, and repeated rendering MUST NOT mutate or reorder caller-owned results.

#### Scenario: Windows and Unix checkouts
- **WHEN** equivalent graph evidence carries Unix and Windows canonical checkout paths
- **THEN** generated tree JSON is byte-identical and contains only portable forward-slash identities

#### Scenario: Repeated JSON render
- **WHEN** one owned tree application result is rendered repeatedly
- **THEN** each document is byte-identical and the original graph and analysis evidence remain unchanged
