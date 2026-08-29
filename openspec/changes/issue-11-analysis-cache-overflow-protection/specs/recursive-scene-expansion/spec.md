## ADDED Requirements

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

## MODIFIED Requirements

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
The expanded summary SHALL take its dependency identities from the authoritative scene-dependency graph and preserve every successfully loaded nested `.tscn` canonical path reachable through resolved instance or inheritance edges, excluding the analyzed root identity. Resolved inheritance traversal SHALL contribute topology and unique dependency evidence without claiming exact inherited metric aggregation. The summary SHALL preserve one external-resource identity for every declaration in every successfully parsed scene used by recursive or graph traversal: the canonical absolute target for a resolved declaration, or the tuple `(declaring canonical scene, document-local resource ID, raw path)` for an unresolved declaration. Applying a cached child more than once MUST union these identities rather than multiply them. Returned identity collections SHALL be owned and deterministically ordered.

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

#### Scenario: Caller mutation cannot alter cached identity sets
- **WHEN** a caller mutates dependency or resource slices returned from a completed invocation
- **THEN** no cached one-occurrence summary or independently repeated invocation is altered

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
