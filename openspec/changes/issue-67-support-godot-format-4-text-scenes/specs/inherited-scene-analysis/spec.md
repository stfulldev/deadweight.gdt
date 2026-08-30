## MODIFIED Requirements

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

