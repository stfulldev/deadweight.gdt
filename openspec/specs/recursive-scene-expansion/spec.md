## Purpose

Defines deterministic recursive expansion of statically resolvable nested text scenes, with correct per-occurrence contributions, unresolved-root evidence, depth composition, closure-wide unique identities, and safe arithmetic.

## Requirements

### Requirement: Every instance mount receives a target classification
The recursive analyzer SHALL classify every local instance mount without silently discarding any branch. A matching external-resource declaration with an existing canonical in-project target whose exact extension is `.tscn` SHALL be attempted as a text scene even when its declared type is not `PackedScene`. Missing or wrong external-resource IDs, path-resolution failures, `SubResource` references, `instance_placeholder` values, imported or binary scene extensions (`.glb`, `.gltf`, `.blend`, `.scn`), inherited-root documents awaiting the dedicated inheritance slice, and other unsupported targets SHALL produce structured unresolved evidence that preserves the declaring scene, resource ID when present, raw target, mount identity, mount depth when known, source position, and classification reason.

#### Scenario: Existing text scene candidate
- **WHEN** a mount's external-resource declaration resolves to an existing canonical `.tscn` file inside the project
- **THEN** that file is attempted as a recursive text-scene target
- **AND** the target is not rejected solely because its declared resource type is incompatible

#### Scenario: Unsupported instance forms
- **WHEN** mounts use a missing external-resource ID, a `SubResource`, an instance placeholder, an imported or binary extension, or an unresolved secure path
- **THEN** every mount produces explicit unresolved evidence with its original source and target context
- **AND** none is silently omitted from occurrence or coverage accounting

#### Scenario: Inherited target is deferred honestly
- **WHEN** a resolved nested `.tscn` parses successfully but its local summary identifies an inherited root
- **THEN** this slice retains inherited-target evidence and one known mounted root instead of claiming an exact child expansion
- **AND** base-scene aggregation remains deferred to the inherited-scene capability

### Requirement: Supported text scenes expand recursively
An existing canonical non-inherited `.tscn` target that parses as supported Godot format 3 SHALL be converted to its local summary and expanded recursively for its own nested mounts. A syntax or supported-format parse failure in a resolved nested `.tscn` SHALL remain a fatal typed analysis failure rather than being downgraded to an unresolved instance. Expansion MUST use canonical absolute scene identities for loading and memoization while retaining normalized display and original target identities for later presentation.

#### Scenario: Three-scene chain
- **WHEN** root scene A mounts supported scene B and B mounts supported scene C
- **THEN** the expanded root summary includes the per-occurrence contributions and closure evidence of B and C
- **AND** each load uses the canonical identity returned by secure path resolution

#### Scenario: Malformed resolved text scene
- **WHEN** a resolved `.tscn` target cannot be parsed as the supported format-3 subset
- **THEN** recursive expansion returns the typed parse failure and does not publish a truncated expanded summary

### Requirement: Resolved instances apply per-occurrence metrics without double-counting roots
For each resolved child occurrence, the analyzer SHALL add the child's expanded `nodes`, `mesh_instances`, `lights`, and `shadow_lights` contributions without adding a separate node for the mount header. It SHALL add one scene-instance occurrence for the mount plus all nested `scene_instances` inside that child occurrence. If the same child summary is applied `N` times, every occurrence-based child counter and nested resolved/unresolved coverage counter SHALL be applied `N` times; unique sets and tree depth MUST NOT be multiplied.

#### Scenario: Resolved child root
- **WHEN** a mount resolves to a child summary containing eight nodes and no nested instances
- **THEN** that occurrence contributes eight nodes and one scene instance
- **AND** it does not contribute a ninth node for the mount header

#### Scenario: One hundred repeated instances
- **WHEN** the same child scene is mounted 100 times and its summary contains nested scene-instance occurrences
- **THEN** its occurrence metrics and nested coverage are applied with multiplicity 100
- **AND** `scene_instances` includes 100 mount occurrences plus 100 copies of the child's nested occurrences

### Requirement: Unresolved instances retain one known root occurrence
Every unresolved non-inherited mount occurrence SHALL contribute exactly one known node and one scene instance, increment unresolved scene-instance coverage, and retain its structured unresolved evidence. It MUST NOT contribute unknown inner nodes, mesh instances, lights, shadow lights, resource declarations, or resolved dependencies. Distinct occurrences remain countable even when their unresolved evidence is later grouped.

#### Scenario: Missing nested text scene
- **WHEN** a mounted `.tscn` target is missing or otherwise unresolved
- **THEN** the expanded summary adds one known node, one scene instance, and one unresolved coverage occurrence
- **AND** no inferred inner metric or dependency contribution is added

#### Scenario: Repeated unresolved target
- **WHEN** the same unresolved target is mounted multiple times
- **THEN** known root, scene-instance, coverage, and unresolved-evidence occurrence counts preserve the full multiplicity

### Requirement: Child tree depth composes at the mount
For a resolved child with known tree depth `C` mounted at known depth `M`, the expanded candidate depth SHALL be `M + C - 1`. For an unresolved child with known mount depth, that mount depth SHALL remain a known candidate maximum. Tree depth SHALL be the maximum across local known depths and every known child candidate; it MUST NOT be multiplied by occurrence count. If either a mount depth or a required child depth is unknown, the analyzer SHALL preserve partial depth evidence and MUST NOT guess a composed value.

#### Scenario: Resolved mounted depth
- **WHEN** a child with tree depth 4 is mounted at local depth 3
- **THEN** its deepest expanded node is considered at depth 6

#### Scenario: Unknown mount depth
- **WHEN** a resolvable child is attached through a mount whose local parent semantics left its depth unknown
- **THEN** the child occurrence metrics may still be expanded
- **AND** no composed depth is invented and partial depth evidence is retained

### Requirement: Canonical resource and dependency identities form unique closure sets
The expanded summary SHALL preserve a unique dependency set containing every successfully resolved nested `.tscn` canonical path in its transitive closure, excluding the analyzed root identity. It SHALL preserve one external-resource identity for every declaration in every successfully parsed scene: the canonical absolute target for a resolved declaration, or the tuple `(declaring canonical scene, document-local resource ID, raw path)` for an unresolved declaration. Applying a cached child more than once MUST union these identities rather than multiply them. Returned identity collections SHALL be deterministically ordered.

#### Scenario: Diamond dependency
- **WHEN** two child branches reach the same canonical descendant scene and declarations resolve to the same canonical external target
- **THEN** each shared scene and resource identity appears once in the expanded closure sets
- **AND** occurrence metrics still include both branches

#### Scenario: Repeated unresolved declaration
- **WHEN** separate parsed scenes declare unresolved resources with the same raw path
- **THEN** their declaring-scene and resource-ID tuple identities remain distinct

### Requirement: Scene work is reused without changing multiplicity
Within one recursive analysis invocation, each canonical `.tscn` identity SHALL be loaded and parsed at most once and its one-occurrence expanded summary SHALL be constructed at most once. Repeated and diamond-shaped occurrences SHALL reuse that cached summary but apply it independently at each mount. Memoized state MUST be invocation-scoped and MUST NOT persist on disk or require invalidation.

#### Scenario: Repeated child load
- **WHEN** one canonical scene is mounted 100 times in the reachable closure
- **THEN** its loader and local-summary construction each run once in that invocation
- **AND** its summary is applied 100 times

#### Scenario: Diamond summary reuse
- **WHEN** branches B and C both reach canonical scene D
- **THEN** D is loaded and expanded once while both occurrence paths receive D's contributions

### Requirement: Recursive arithmetic cannot wrap
Every non-negative `int64` addition and multiplication performed while counting occurrences, applying child summaries, composing known depths, or accumulating coverage SHALL use checked arithmetic. An overflow SHALL stop expansion with a typed fatal error exposing diagnostic code `SB2004`; the analyzer MUST NOT panic, wrap to a negative value, clamp, or publish a partial summary. Equivalent inputs SHALL produce equivalent summaries and unresolved evidence ordering.

#### Scenario: Repeated metric overflows
- **WHEN** applying a repeated child contribution would exceed the maximum signed 64-bit value
- **THEN** expansion returns a typed `SB2004` failure and no wrapped metric collection

#### Scenario: Deterministic expanded summary
- **WHEN** equivalent acyclic scene closures are analyzed repeatedly
- **THEN** metrics, coverage, unique identities, and unresolved evidence are returned in the same deterministic order and values
