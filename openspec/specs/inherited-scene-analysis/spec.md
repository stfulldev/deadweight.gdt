## Purpose

Defines the deliberately limited approximate analysis of Godot inherited scenes, including base-summary reuse, explicit local additions, unsupported override evidence, and honest fallback behavior.

## Requirements

### Requirement: Inherited syntax is retained as approximation evidence
A parsed scene whose root carries an instance reference SHALL be classified as inherited. Its root MUST NOT contribute a new ordinary node or scene-instance occurrence; typeless node stubs MUST remain override evidence rather than new nodes, and `[editable ...]` syntax SHALL be retained as editable-child evidence. Any inherited root, override stub, or editable signal SHALL make the successful root result `partial approximate`.

#### Scenario: Inherited root with overrides
- **WHEN** a scene root references a base, later typeless nodes override base paths, and an editable entry is present
- **THEN** the root and stubs contribute no new ordinary nodes or scene instances
- **AND** the result retains inherited, override, and editable evidence as `partial approximate`

### Requirement: Resolved base summary is applied exactly once
When an inherited root resolves to a supported acyclic format-3 or format-4 `.tscn` base, the analyzer SHALL apply the base's one-occurrence expanded summary exactly once, then add explicit local typed nodes and expand explicit local nested mounts. It MUST NOT add a scene-instance occurrence for the inheritance edge or count a second inherited root node. Occurrence metrics SHALL add base and explicit local contributions; tree depth SHALL be the maximum known base/local contribution; unique resource and dependency identities SHALL remain canonical unions.

#### Scenario: Base plus explicit local additions
- **WHEN** a supported format-3 or format-4 base contributes five nodes and one light while the inherited scene declares two explicit typed nodes and one nested instance
- **THEN** the inherited one-occurrence summary contains the base metrics plus the two local nodes and expanded nested contribution
- **AND** inheritance itself adds neither a scene instance nor a duplicate root

#### Scenario: Transitive inheritance
- **WHEN** derived scene C inherits B and B inherits A through resolved supported format-3 or format-4 text scenes
- **THEN** each base summary is applied once in the effective chain
- **AND** all inheritance dependencies remain cycle-checked and unique

### Requirement: Unsupported bases retain one known inherited root
If an inherited base is missing, unreadable, imported, binary, `SubResource`-backed, UID-only, `user://`, outside the project, or otherwise unsupported, analysis SHALL remain successful and approximate. The unavailable base SHALL contribute exactly one known inherited root node, no inferred inner metrics or scene-instance occurrence, and complete target/classification evidence; explicit local typed nodes and supported local mounts SHALL still contribute independently.

#### Scenario: Missing base
- **WHEN** an inherited root references a missing `.tscn` base
- **THEN** one known root plus explicit local contributions are published as `partial approximate`
- **AND** no missing base dependency or inner metric is invented

#### Scenario: Imported base
- **WHEN** an inherited root references an existing `.glb`, `.gltf`, `.blend`, or `.scn` target
- **THEN** the imported contents remain unexpanded while one known inherited root and the canonical external resource identity are retained

### Requirement: Inheritance evidence composes by occurrence
Every inherited scene document SHALL retain one inheritance record containing the declaring scene, base resource ID, raw and resolved target identities when available, classification, resolution reason, source position, override presence, and editable presence. Applying an inherited child summary `N` times SHALL multiply inherited evidence and all occurrence-based base/local metrics by `N`, while the base document is parsed and summarized once, unique identities remain unique, and inherited-root coverage does not create scene-instance coverage.

#### Scenario: Repeated inherited child
- **WHEN** one inherited child is mounted 100 times
- **THEN** its effective occurrence metrics and `SB1003` occurrence count are multiplied by 100
- **AND** its canonical base is parsed once and inheritance itself does not add 100 scene instances

### Requirement: Inheritance diagnostics and fatal boundaries stay explicit
Every successful inherited result SHALL emit grouped `SB1003` warning evidence with the base target and `approximate` reliability, including when the base is unsupported. A malformed resolved supported `.tscn` base or any resolved instance/inheritance cycle SHALL remain fatal and return no usable result. Override properties, removals, reordering, type replacement, owner semantics, and editable-child effective changes MUST NOT be simulated or described as exact.

#### Scenario: Inherited cycle
- **WHEN** resolved inheritance edges form A to B to A
- **THEN** analysis fails with the deterministic cycle diagnostic and publishes no approximate result

#### Scenario: Unsupported override properties
- **WHEN** inherited stubs contain property overrides that could change counts in either direction
- **THEN** the known base/local signal remains approximate and no override merge is claimed
