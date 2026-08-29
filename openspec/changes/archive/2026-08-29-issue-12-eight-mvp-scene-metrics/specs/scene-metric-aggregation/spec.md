## Purpose

Defines final publication and deterministic ordering of the eight frozen MVP 0.1 scene metrics from validated local, recursive, graph, and unique-identity evidence.

## ADDED Requirements

### Requirement: Successful root analysis publishes all eight metrics
A successful root analysis SHALL publish non-negative `int64` values for `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, `shadow_lights`, `external_resources`, and `scene_dependencies`. The first six values SHALL preserve the recursive occurrence/maximum aggregation contracts, while the last two SHALL be finalized only after graph validation and closure-wide identity union. Fatal analysis MUST return no usable metric result.

#### Scenario: Root-only scene
- **WHEN** a supported root has one ordinary node and no resources or dependencies
- **THEN** nodes and tree depth are one and the other six metrics are zero

#### Scenario: Cached children do not retain root unique counts
- **WHEN** a one-occurrence child summary is cached and applied repeatedly
- **THEN** its occurrence metrics multiply according to mounts
- **AND** unique root metric fields are finalized once after application rather than stored or multiplied in the child cache

### Requirement: Literal occurrence metrics keep frozen type semantics
Final occurrence metrics SHALL count only ordinary literal node types retained by local summaries: exact `MeshInstance3D` for `mesh_instances`, and exact `DirectionalLight3D`, `OmniLight3D`, or `SpotLight3D` for `lights`. `shadow_lights` SHALL count only those supported literal lights with explicitly parsed `shadow_enabled = true`; absent or false SHALL count zero. Custom subclasses, `MultiMeshInstance3D`, 2D lights, mounts, inherited roots, override stubs, and imported-scene contents MUST NOT be inferred as supported literal occurrences.

#### Scenario: Supported and unsupported literal types
- **WHEN** a closure contains the supported mesh and three supported 3D lights alongside custom, multi-mesh, 2D-light, and imported-scene evidence
- **THEN** only the exact supported literal ordinary nodes contribute to mesh and light metrics

#### Scenario: Default shadow value
- **WHEN** a supported light omits `shadow_enabled` or sets it to false
- **THEN** it contributes zero shadow lights

### Requirement: External resources are a canonical unique union
`external_resources` SHALL equal the checked count of closure-wide resource identities: one canonical absolute target for every resolved declaration, or one tuple `(declaring canonical scene, document-local resource ID, raw path)` for every unresolved declaration. Repeated occurrences and declarations in different parsed scenes that resolve to the same canonical target MUST NOT multiply the resolved identity. Resources inside unparsed imported, missing, or otherwise unavailable scenes MUST NOT be invented.

#### Scenario: Repeated shared resource
- **WHEN** repeated and diamond scene occurrences declare resources resolving to one canonical target
- **THEN** that target contributes one to `external_resources`

#### Scenario: Unresolved declarations remain document-local
- **WHEN** two parsed scenes contain unresolved declarations with the same raw path but different declaring scene or resource ID
- **THEN** both tuple identities contribute separately

### Requirement: Scene dependencies use the validated unique graph count
`scene_dependencies` SHALL equal the graph's checked number of unique successfully loaded non-root `.tscn` nodes reachable by resolved instance or inheritance edges. Repeated and diamond paths MUST NOT multiply it; transitive and inheritance nodes SHALL contribute; unresolved or imported targets MUST NOT contribute. A cyclic graph MUST fail before metrics are published.

#### Scenario: Repeated diamond dependencies
- **WHEN** root A reaches B and C and both reach D, with any edge repeated
- **THEN** `scene_dependencies` is three for B, C, and D

#### Scenario: Resolved inherited base
- **WHEN** a parsed reachable scene inherits a successfully loaded base
- **THEN** both the reachable scene and base contribute unique dependency identities
- **AND** this dependency count does not claim inherited effective-tree occurrence metrics

### Requirement: Metric order is fixed and independent of maps
Every ordered metric consumer SHALL use the canonical sequence `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, `shadow_lights`, `external_resources`, `scene_dependencies`. Equivalent inputs SHALL produce identical values and order regardless of source or map iteration, and final metric validation SHALL reject any negative value.

#### Scenario: Ordered lookup
- **WHEN** all eight metric values are enumerated for output, configuration, presets, or budgets
- **THEN** their identifiers appear in the canonical sequence with the corresponding values

### Requirement: Frozen aggregation example composes by occurrence
Recursive finalization SHALL apply one-occurrence child summaries per mount without double-counting resolved instance roots. For the MVP §20.7 City/Building/Lamp example, two Building occurrences whose summary contains 20 nodes and one nested Lamp instance plus three direct four-node Lamp occurrences SHALL produce 62 nodes and seven scene instances. Shared resource identities SHALL remain unique.

#### Scenario: City Building Lamp example
- **WHEN** City has 10 local ordinary nodes, two Building occurrences, and three direct Lamp occurrences using the documented child summaries
- **THEN** final nodes equal 62 and scene instances equal seven
- **AND** one shared canonical texture contributes one external resource
