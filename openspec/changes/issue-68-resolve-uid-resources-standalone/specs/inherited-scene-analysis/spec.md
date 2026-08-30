## MODIFIED Requirements

### Requirement: Resolved base summary is applied exactly once
When an inherited root resolves by UID or path to a supported acyclic format-3 or format-4 `.tscn` base, the analyzer SHALL apply the base's one-occurrence expanded summary exactly once, then add explicit local typed nodes and expand explicit local nested mounts. It MUST NOT add a scene-instance occurrence for the inheritance edge or count a second inherited root node. Occurrence metrics SHALL add base and explicit local contributions; tree depth SHALL be the maximum known base/local contribution; unique resource and dependency identities SHALL remain canonical unions.

#### Scenario: Base plus explicit local additions
- **WHEN** a supported format-3 or format-4 base contributes five nodes and one light while the inherited scene declares two explicit typed nodes and one nested instance
- **THEN** the inherited one-occurrence summary contains the base metrics plus the two local nodes and expanded nested contribution
- **AND** inheritance itself adds neither a scene instance nor a duplicate root

#### Scenario: Transitive inheritance
- **WHEN** derived scene C inherits B and B inherits A through resolved supported format-3 or format-4 text scenes
- **THEN** each base summary is applied once in the effective chain
- **AND** all inheritance dependencies remain cycle-checked and unique

#### Scenario: UID-resolved base
- **WHEN** an inherited root's base declaration resolves only through a unique secure project UID mapping
- **THEN** the mapped canonical base follows the same one-occurrence expansion and cycle contracts as a path-resolved base

### Requirement: Unsupported bases retain one known inherited root
If an inherited base is missing, unreadable, imported, binary, `SubResource`-backed, unresolvable by UID, `user://`, outside the project, or otherwise unsupported, analysis SHALL remain successful and approximate. The unavailable base SHALL contribute exactly one known inherited root node, no inferred inner metrics or scene-instance occurrence, and complete target/classification evidence; explicit local typed nodes and supported local mounts SHALL still contribute independently.

#### Scenario: Missing base
- **WHEN** an inherited root references a missing `.tscn` base
- **THEN** one known root plus explicit local contributions are published as `partial approximate`
- **AND** no missing base dependency or inner metric is invented

#### Scenario: Imported base
- **WHEN** an inherited root references an existing `.glb`, `.gltf`, `.blend`, or `.scn` target
- **THEN** the imported contents remain unexpanded while one known inherited root and the canonical external resource identity are retained

#### Scenario: Ambiguous UID base
- **WHEN** an inherited base UID has conflicting direct project ownership claims and no usable path evidence
- **THEN** one known inherited root and deterministic UID ambiguity evidence are retained as `partial approximate`
