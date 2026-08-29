## Purpose

Defines an explainable deterministic scene-dependency graph, including resolved and unresolved instance or inheritance edges, stable cycle diagnostics, and unique dependency counting.

## Requirements

### Requirement: Scene dependencies form an owned canonical graph
The analyzer SHALL represent every successfully loaded reachable text scene as one graph node identified by its canonical absolute path while retaining its normalized display path. It SHALL represent every statically observed scene instance or inherited-root target as a directed edge from the declaring scene, preserving the edge kind, raw target, external-resource ID when present, resolution status and reason, canonical/display target identities when resolved, and a non-negative occurrence count. Returned nodes and edges SHALL be owned and deterministically ordered.

#### Scenario: Resolved instance edge
- **WHEN** scene A declares an instance target that resolves to supported scene B
- **THEN** the graph contains canonical nodes for A and B plus a resolved instance edge from A to B
- **AND** the edge retains the declaring and target display identities and original resource reference

#### Scenario: Caller mutation cannot alter later analysis
- **WHEN** a caller mutates graph slices returned by one analysis invocation
- **THEN** a later analysis returns the original deterministic graph values

### Requirement: Equivalent edges compact without losing multiplicity
Edges with the same declaring scene, target identity or unresolved classification, raw target, resource ID, resolution reason, and edge kind SHALL be compacted into one graph edge whose occurrence count is their checked sum. Resolved node identity and dependency counting MUST remain unique while occurrence-based recursive metrics continue to apply every mount separately. Edge occurrence arithmetic SHALL use the recursive checked-arithmetic contract and expose `SB2004` on overflow.

#### Scenario: One hundred repeated instance edges
- **WHEN** one scene mounts the same canonical child through 100 equivalent declarations
- **THEN** the graph contains one resolved edge with 100 occurrences and one child node
- **AND** recursive occurrence metrics still include all 100 mounted copies

#### Scenario: Edge occurrence overflow
- **WHEN** compacting equivalent edges would exceed the maximum signed 64-bit value
- **THEN** graph construction returns a typed `SB2004` failure without a wrapped graph

### Requirement: Unresolved targets remain edges without graph nodes
A scene target that is missing, unavailable, imported or binary, placeholder-backed, `SubResource`-backed, unsupported, or otherwise unresolved SHALL remain a graph edge with an empty canonical target and complete unresolved classification evidence. It MUST NOT create a resolved graph node or contribute to the resolved dependency count.

#### Scenario: Missing nested text scene
- **WHEN** a declared nested `.tscn` target cannot be resolved or loaded
- **THEN** the graph retains one unresolved instance edge with its reason and occurrence count
- **AND** no target node or resolved dependency is invented

#### Scenario: Unresolved inherited base
- **WHEN** an inherited root references a base scene that cannot be resolved
- **THEN** the graph retains an unresolved inheritance edge from the inherited scene
- **AND** the edge remains distinguishable from an unresolved nested instance

### Requirement: Resolved inheritance participates in graph traversal
A parsed inherited-root scene SHALL produce an inheritance edge to its resolved supported text-scene base. The analyzer SHALL follow that edge for graph topology, transitive dependency identities, cache reuse, and cycle validation while MUST NOT claim inherited effective-tree metric aggregation before the inherited-scene capability is applied.

#### Scenario: Resolved inherited base dependency
- **WHEN** root A mounts inherited scene B and B inherits supported base C
- **THEN** the graph contains an instance edge A to B and an inheritance edge B to C
- **AND** both B and C contribute unique dependency identities while C's metrics are not merged as an exact inherited result

#### Scenario: Inheritance reaches a transitive instance
- **WHEN** inherited base C mounts supported scene D
- **THEN** D is present in the reachable graph and unique dependency set
- **AND** traversal reuses the same canonical scene work if D is reached elsewhere

### Requirement: DFS reports deterministic complete cycle chains
Before publishing a successful graph-backed recursive result, the analyzer SHALL traverse resolved instance and inheritance edges with explicit unvisited, visiting, and visited states. An edge to a visiting node SHALL stop analysis with a typed fatal error exposing `SB2002`, the complete canonical and display cycle chain with the repeated start node at the end, and no usable graph or expanded summary. Outgoing traversal order and the selected cycle SHALL be deterministic for equivalent inputs.

#### Scenario: Self-cycle
- **WHEN** scene A contains a resolved edge to A
- **THEN** the error chain is `A → A` and exposes `SB2002`

#### Scenario: Multi-scene cycle
- **WHEN** resolved edges form `A → B → C → A`
- **THEN** the fatal error retains the full canonical and display chain `A → B → C → A`
- **AND** no truncated metric or budget-ready result is returned

#### Scenario: Diamond without cycle
- **WHEN** A reaches D through both B and C without a back edge
- **THEN** D transitions to visited once and the graph succeeds without a false cycle

### Requirement: Resolved dependency count is a unique graph value
The graph SHALL expose `scene_dependencies` as the number of unique successfully loaded text-scene nodes reachable from the root through resolved instance or inheritance edges, excluding the root node. Repeated and diamond-shaped paths MUST NOT multiply this value, while transitive and inheritance targets SHALL be included. The value SHALL use checked non-negative `int64` accumulation.

#### Scenario: Repeated dependency counts once
- **WHEN** the same supported child is mounted 100 times
- **THEN** `scene_dependencies` is one for that child

#### Scenario: Transitive diamond dependency count
- **WHEN** root A reaches B and C and both reach D
- **THEN** `scene_dependencies` is three for B, C, and D

#### Scenario: Root is excluded after a cycle-free back-free traversal
- **WHEN** a cycle-free graph contains only the analyzed root
- **THEN** `scene_dependencies` is zero
