## Purpose

Defines deterministic recursive expansion of statically resolvable nested text scenes, with correct per-occurrence contributions, unresolved-root evidence, depth composition, closure-wide unique identities, and safe arithmetic.

## Requirements

### Requirement: Every instance mount receives a target classification
The recursive analyzer SHALL classify every local instance mount without silently discarding any branch. A matching external-resource declaration with an existing canonical in-project target whose exact extension is `.tscn` SHALL be attempted as a text scene even when its declared type is not `PackedScene`. Missing or wrong external-resource IDs, path-resolution failures, `SubResource` references, `instance_placeholder` values, imported or binary scene extensions (`.glb`, `.gltf`, `.blend`, `.scn`), and other unsupported targets SHALL produce structured unresolved evidence that preserves the declaring scene, resource ID when present, raw target, mount identity, mount depth when known, source position, and classification reason. A successfully parsed inherited-root child SHALL use the limited inherited-scene analysis contract and retain approximation evidence instead of being downgraded to an unresolved one-known-root child.

#### Scenario: Existing text scene candidate
- **WHEN** a mount's external-resource declaration resolves to an existing canonical `.tscn` file inside the project
- **THEN** that file is attempted as a recursive text-scene target
- **AND** the target is not rejected solely because its declared resource type is incompatible

#### Scenario: Unsupported instance forms
- **WHEN** mounts use a missing external-resource ID, a `SubResource`, an instance placeholder, an imported or binary extension, or an unresolved secure path
- **THEN** every mount produces explicit unresolved evidence with its original source and target context
- **AND** none is silently omitted from occurrence or coverage accounting

#### Scenario: Inherited target is deferred honestly
- **WHEN** a resolved nested `.tscn` parses successfully and its local summary identifies an inherited root
- **THEN** the child applies its supported base and explicit local contributions through the inherited-scene contract
- **AND** its occurrence retains approximate inheritance evidence rather than claiming exact expansion

### Requirement: Supported text scenes expand recursively
An existing canonical non-inherited `.tscn` target that parses as supported Godot format 3 or format 4 SHALL be converted to its local summary and expanded recursively for its own nested mounts. A syntax or supported-format parse failure in a resolved nested `.tscn` SHALL remain a fatal typed analysis failure rather than being downgraded to an unresolved instance. Expansion MUST use canonical absolute scene identities for loading and memoization while retaining normalized display and original target identities for later presentation.

#### Scenario: Three-scene chain
- **WHEN** root scene A mounts supported scene B and B mounts supported scene C across any supported combination of format 3 and format 4
- **THEN** the expanded root summary includes the per-occurrence contributions and closure evidence of B and C
- **AND** each load uses the canonical identity returned by secure path resolution

#### Scenario: Malformed resolved text scene
- **WHEN** a resolved `.tscn` target declares format 3 or format 4 but cannot be parsed as the supported subset
- **THEN** recursive expansion returns the typed parse failure and does not publish a truncated expanded summary

#### Scenario: Unknown future nested format
- **WHEN** a resolved `.tscn` dependency declares an unknown format greater than 4
- **THEN** recursive expansion returns the typed unsupported-format failure rather than treating the dependency as an unresolved partial mount

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
For a resolved child with known tree depth `C` mounted at known depth `M`, the expanded candidate depth SHALL be `M + C - 1`. For an unresolved child with known mount depth, that mount depth SHALL remain a known candidate maximum. Tree depth SHALL be the maximum across local known depths and every known child candidate; it MUST NOT be multiplied by occurrence count or changed when a one-occurrence child summary is reused from cache. If either a mount depth or a required child depth is unknown, the analyzer SHALL preserve partial depth evidence and MUST NOT guess a composed value.

#### Scenario: Resolved mounted depth
- **WHEN** a child with tree depth 4 is mounted at local depth 3
- **THEN** its deepest expanded node is considered at depth 6

#### Scenario: Unknown mount depth
- **WHEN** a resolvable child is attached through a mount whose local parent semantics left its depth unknown
- **THEN** the child occurrence metrics may still be expanded
- **AND** no composed depth is invented and partial depth evidence is retained

#### Scenario: Repeated cached depth
- **WHEN** one cached child summary is applied 100 times at the same known mount depth
- **THEN** its composed depth is considered as one maximum candidate
- **AND** the depth is not multiplied by 100

### Requirement: Canonical resource and dependency identities form unique closure sets
The expanded summary SHALL take its dependency identities from the authoritative scene-dependency graph and preserve every successfully loaded nested `.tscn` canonical path reachable through resolved instance or inheritance edges, excluding the analyzed root identity. Resolved inheritance traversal SHALL contribute topology, unique dependency evidence, and the deliberately approximate base-summary contribution defined by the inherited-scene capability. The summary SHALL preserve one external-resource identity for every declaration in every successfully parsed scene used by recursive or graph traversal: the canonical absolute target for a resolved declaration, or the tuple `(declaring canonical scene, document-local resource ID, raw path)` for an unresolved declaration. Applying a cached child more than once MUST union these identities rather than multiply them. Returned identity collections SHALL be owned and deterministically ordered.

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
- **AND** the base's one-occurrence summary contributes once while the result remains approximate

#### Scenario: Caller mutation cannot alter cached identity sets
- **WHEN** a caller mutates dependency or resource slices returned from a completed invocation
- **THEN** no cached one-occurrence summary or independently repeated invocation is altered

### Requirement: Parsed scene coverage comes from successful cache cardinality
After a successful recursive analysis, the result SHALL expose `parsed_scene_files` as the checked non-negative `int64` count of unique canonical scene documents successfully stored in that invocation's parse cache, including the analyzed root and every successfully parsed reachable instance or inheritance scene. Repeated occurrences and diamond paths MUST NOT multiply this coverage value. Unresolved, unavailable, and parse-failed targets MUST NOT contribute a successful parsed-file entry, and a fatal analysis SHALL return no usable coverage result.

#### Scenario: Repeated parsed child counts once
- **WHEN** a root successfully reaches the same parsed child through 100 instance occurrences
- **THEN** `parsed_scene_files` is two for the root and child
- **AND** occurrence metrics still apply all 100 child instances

#### Scenario: Diamond cache cardinality
- **WHEN** root A reaches B and C and both branches reach successfully parsed D
- **THEN** `parsed_scene_files` is four for A, B, C, and D
- **AND** the two occurrence paths to D do not add a fifth parsed file

#### Scenario: Unavailable target is not parsed coverage
- **WHEN** a declared scene target resolves but cannot be opened or parsed successfully
- **THEN** it does not contribute a successful parse-cache entry
- **AND** fatal parse failures return no usable recursive result

### Requirement: Scene work is reused without changing multiplicity
Within one recursive analysis invocation, graph discovery and occurrence expansion SHALL share a canonical-path parse cache. Each canonical `.tscn` identity SHALL be physically opened at most once and have parsing attempted at most once through independently injectable effects; successful documents and deterministic open or parse failures SHALL be memoized for that invocation. Each successfully parsed document SHALL have its local summary and one-occurrence expanded summary constructed at most once. Repeated and diamond-shaped occurrences SHALL reuse cached work but apply occurrence metrics, coverage, and grouped evidence independently at every mount. Cached and returned values SHALL be owned copies. All memoized state MUST be invocation-scoped and MUST NOT persist on disk, require invalidation, or introduce concurrent parsing.

#### Scenario: Repeated child load
- **WHEN** one canonical scene is mounted 100 times in the reachable closure
- **THEN** its physical opener, parser, local-summary construction, and expanded-summary construction each run once in that invocation
- **AND** its one-occurrence summary is applied 100 times

#### Scenario: Diamond summary reuse
- **WHEN** branches B and C both reach canonical scene D
- **THEN** D is physically opened, parsed, locally summarized, and recursively expanded once
- **AND** both occurrence paths receive D's metric and evidence contributions

#### Scenario: Graph and expansion share parsed work
- **WHEN** graph discovery parses a canonical scene before occurrence expansion reaches it
- **THEN** occurrence expansion reuses the same parsed document and local summary without a second open or parse

#### Scenario: Invocation isolation
- **WHEN** the same analyzer runs two separate successful analysis invocations for the same root
- **THEN** each invocation allocates independent caches and performs its own physical reads and parses
- **AND** no persistent invalidation mechanism is required

#### Scenario: Cached failure is stable
- **WHEN** equivalent paths in one invocation reach a canonical scene whose open or parse effect fails
- **THEN** the effect is attempted once and the same typed failure classification is reused
- **AND** a fatal parse failure cannot publish a partial recursive result

### Requirement: Recursive arithmetic cannot wrap
Every non-negative `int64` addition and multiplication performed while compacting graph edges, counting dependencies or parsed files, counting occurrences, applying cached child summaries, composing known depths, accumulating coverage, or grouping evidence SHALL use checked arithmetic. Invalid negative operands or overflow SHALL stop analysis with a typed fatal error exposing diagnostic code `SB2004`; the analyzer MUST NOT panic, wrap to a negative value, clamp, mutate a reusable cached summary, or publish a partial graph or recursive result. Equivalent inputs SHALL produce equivalent errors or deterministically owned results.

#### Scenario: Repeated metric overflows
- **WHEN** applying a repeated cached child contribution would exceed the maximum signed 64-bit value
- **THEN** analysis returns a typed `SB2004` failure and no wrapped metric collection

#### Scenario: Edge or coverage accumulation overflows
- **WHEN** compacted edge occurrences, dependency counts, parsed-file coverage, or scene-instance coverage cannot be represented as a non-negative `int64`
- **THEN** analysis returns `SB2004` without a partial graph or recursive result

#### Scenario: Invalid negative arithmetic input
- **WHEN** an aggregation boundary receives a negative operand that violates the non-negative counter contract
- **THEN** it returns `SB2004` without panic, clamping, or cached-state mutation

#### Scenario: Deterministic expanded summary
- **WHEN** equivalent acyclic scene closures are analyzed repeatedly
- **THEN** metrics, coverage, unique identities, and unresolved or inherited evidence are returned in the same deterministic order and values

### Requirement: Recursive application preserves direct contribution evidence
The recursive analyzer SHALL construct one-occurrence direct contribution evidence from each cached local summary and SHALL apply child contribution rows with the same checked reachable multiplicity used for occurrence metrics. Nested rows SHALL be propagated independently from parent direct values, root-relative depth candidates SHALL be composed at each mount, and equivalent immediate contexts SHALL be compacted without losing occurrences. Graph discovery and contribution construction MUST continue sharing the invocation-scoped parse and local-summary cache.

#### Scenario: Cached repeated child contribution
- **WHEN** a supported child is mounted 100 times through equivalent context
- **THEN** its document and local summary are built once while its direct contribution occurrence count and additive values are applied 100 times

#### Scenario: Diamond descendant contribution
- **WHEN** two parent scenes reach the same canonical descendant
- **THEN** the descendant's cached direct evidence is reused for both reachable paths
- **AND** its context-aware contribution multiplicities account for both paths without duplicating parse work

#### Scenario: Mounted depth composition
- **WHEN** a child contribution has a known relative depth candidate and is applied at a known parent mount depth
- **THEN** its root-relative candidate is composed using checked depth arithmetic
- **AND** an unknown required depth remains unknown rather than guessed
