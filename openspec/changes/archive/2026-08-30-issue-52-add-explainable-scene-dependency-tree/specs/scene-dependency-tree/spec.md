## Purpose

Defines a bounded, deterministic, and portable rooted projection of the authoritative scene-dependency graph for human and machine explanation.

## ADDED Requirements

### Requirement: The dependency tree projects the authoritative graph
A successful dependency-tree result SHALL be derived from the recursive analysis result's authoritative dependency graph and MUST NOT reopen, reparse, rediscover, or independently resolve scene files. The projection SHALL identify the analyzed root and represent every retained resolved or unresolved graph edge exactly once. It MUST reject internally inconsistent graph evidence rather than silently omitting an edge, inventing a node, or publishing a partial projection.

#### Scenario: Root-only graph
- **WHEN** recursive analysis returns one root graph node and no edges
- **THEN** the dependency tree identifies the portable root and contains no edge entries

#### Scenario: Complete graph projection
- **WHEN** a validated graph contains resolved, inheritance, and unresolved edges reachable from its root
- **THEN** every graph edge appears exactly once in the dependency-tree projection
- **AND** no scene file is opened or parsed again for tree construction

#### Scenario: Inconsistent injected graph
- **WHEN** a graph edge names a missing source node or a resolved target without a graph node
- **THEN** tree presentation fails deterministically without publishing a truncated tree

### Requirement: Depth-first expansion is deterministic and bounded
The dependency tree SHALL use deterministic portable edge ordering and rooted depth-first expansion. The first resolved edge to a target SHALL expand that target's outgoing edges; a later resolved edge to an already expanded target SHALL remain visible as a back-reference and MUST NOT expand the target again. Equivalent repeated declarations SHALL remain one compacted edge with checked occurrence multiplicity. The projection SHALL remain bounded by graph nodes plus graph edges and MUST NOT recursively duplicate diamond or repeated subgraphs.

#### Scenario: Repeated child edge
- **WHEN** one compacted root edge reaches a child with 100 occurrences
- **THEN** the tree emits one child edge with occurrence count 100 and expands the child subtree once

#### Scenario: Diamond back-reference
- **WHEN** root A reaches D through both B and C
- **THEN** the first portable depth-first path expands D and the later edge to D is emitted as a back-reference
- **AND** D's outgoing subtree is not duplicated under the later path

#### Scenario: Stable sibling order
- **WHEN** equivalent graph evidence is produced from different checkout paths or supported operating systems
- **THEN** tree entry depth, sibling order, and back-reference placement are byte-stable after holding the tool version constant

### Requirement: Every edge retains explainable evidence and reliability impact
Each dependency-tree edge SHALL retain portable source and target identity, edge kind `instance` or `inheritance`, checked positive occurrence count, resolved state, and stable back-reference state. Resolved instance edges SHALL identify exact dependency evidence; unresolved non-inheritance edges SHALL identify lower-bound impact and retain their classification, resolution reason, resource ID, and safe raw target when available; inheritance edges SHALL identify approximate impact whether or not their target resolves. Unavailable identities MUST use retained portable target evidence and MUST NOT expose or invent canonical paths.

#### Scenario: Resolved instance and inheritance
- **WHEN** a scene has one resolved instance edge and one resolved inheritance edge
- **THEN** both targets are visible with multiplicity and edge kind
- **AND** the instance edge is exact while the inheritance edge is approximate

#### Scenario: Imported target
- **WHEN** an instance edge targets an imported scene that cannot be expanded statically
- **THEN** the edge remains visible as unresolved lower-bound evidence with classification `imported_scene`

#### Scenario: Unresolved inherited base
- **WHEN** an inheritance edge has no resolved target node
- **THEN** it remains visible as approximate inheritance evidence with its retained classification and target context

### Requirement: Tree projection preserves fatal and partial analysis boundaries
The tree command SHALL preserve recursive analysis status, reliability, coverage, diagnostics, and typed fatal errors. A successful partial analysis SHALL publish its complete dependency-tree evidence with the existing conservative reliability and exit `0`. A resolved dependency cycle SHALL remain fatal `SB2002`, SHALL publish no successful tree, and MUST NOT be converted into a back-reference or partial result.

#### Scenario: Partial tree succeeds honestly
- **WHEN** analysis succeeds with unresolved dependency edges and partial lower-bound evidence
- **THEN** the complete known tree and diagnostics are published with partial lower-bound status and exit code `0`

#### Scenario: Cycle remains fatal
- **WHEN** recursive graph validation detects `A → B → A`
- **THEN** the command returns the existing `SB2002` cycle chain with exit code `2`
- **AND** no text or JSON dependency tree is emitted as a successful result
