## MODIFIED Requirements

### Requirement: Canonical resource and dependency identities form unique closure sets
The expanded summary SHALL take its dependency identities from the authoritative scene-dependency graph and preserve every successfully loaded nested `.tscn` canonical path reachable through resolved instance or inheritance edges, excluding the analyzed root identity. Resolved inheritance traversal SHALL contribute topology and unique dependency evidence without claiming exact inherited metric aggregation. The summary SHALL preserve one external-resource identity for every declaration in every successfully parsed scene used by recursive or graph traversal: the canonical absolute target for a resolved declaration, or the tuple `(declaring canonical scene, document-local resource ID, raw path)` for an unresolved declaration. Applying a cached child more than once MUST union these identities rather than multiply them. Returned identity collections SHALL be deterministically ordered.

#### Scenario: Diamond dependency
- **WHEN** two child branches reach the same canonical descendant scene and declarations resolve to the same canonical external target
- **THEN** each shared scene and resource identity appears once in the expanded closure sets
- **AND** occurrence metrics still include both branches

#### Scenario: Repeated unresolved declaration
- **WHEN** separate parsed scenes declare unresolved resources with the same raw path
- **THEN** their declaring-scene and resource-ID tuple identities remain distinct

#### Scenario: Inherited topology without inherited metric expansion
- **WHEN** a parsed nested scene inherits a resolved supported base scene
- **THEN** the nested and base canonical paths appear in graph-backed dependency identities
- **AND** the base's effective-tree metrics are not merged as an exact inherited contribution before the inherited-scene capability
