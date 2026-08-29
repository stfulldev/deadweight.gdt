## ADDED Requirements

### Requirement: Published metrics agree with contribution aggregation modes
After successful graph validation, the analyzer SHALL validate that contribution rows sum exactly to the final `nodes`, `scene_instances`, `mesh_instances`, `lights`, and `shadow_lights` values, and that the maximum known contribution depth candidate agrees with final `tree_depth` whenever depth is complete. `external_resources` and `scene_dependencies` SHALL remain finalized from their authoritative unique unions and MUST NOT be reconstructed by summing per-scene rows. Any invariant violation SHALL be fatal rather than publishing inconsistent analysis evidence.

#### Scenario: Exact additive reconciliation
- **WHEN** recursive analysis completes successfully with resolved repeated children
- **THEN** every additive contribution sum equals the corresponding root metric exactly

#### Scenario: Partial depth
- **WHEN** unsupported parent evidence prevents a complete set of depth candidates
- **THEN** known candidates remain available but the analyzer does not claim an exact contribution reconciliation for tree depth

#### Scenario: Unique totals remain authoritative
- **WHEN** resources or dependencies are shared by several scene contexts
- **THEN** root unique metrics equal their validated union cardinalities
- **AND** no sum of context references replaces those totals

