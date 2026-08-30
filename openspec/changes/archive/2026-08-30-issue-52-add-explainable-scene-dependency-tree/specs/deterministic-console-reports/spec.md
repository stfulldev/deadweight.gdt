## ADDED Requirements

### Requirement: Tree text explains dependencies without unbounded expansion
The dependency-tree text report SHALL render the tool version, portable root scene, project, analysis status, and accuracy followed by a rooted tree whose deterministic connector indentation exposes every edge's kind, occurrence count, target identity, reliability, and back-reference or unresolved classification when applicable. Resolved children SHALL appear as branches, repeated compacted edges SHALL show multiplicity, and later diamond paths SHALL use an explicit back-reference marker instead of duplicating an expanded subtree. The report SHALL retain grouped diagnostics and the established partial or approximate warning semantics after the tree.

#### Scenario: Complete chain
- **WHEN** a complete graph contains root A, instance child B, and inherited base C
- **THEN** text renders A as the root followed by nested instance and inheritance branches with their multiplicities and reliability

#### Scenario: Diamond and repeated branches
- **WHEN** a graph contains a repeated edge and a diamond target
- **THEN** text renders the checked repeated count once and marks the later diamond target as a back-reference
- **AND** no reachable edge disappears or expands without bound

#### Scenario: Partial unresolved branch
- **WHEN** a tree contains an imported or unavailable target
- **THEN** text names its portable target, stable classification, and non-exact reliability without relying on ANSI color

### Requirement: Tree text is portable, deterministic, and non-mutating
Tree rendering SHALL sort from a caller-owned projection using portable identities and stable edge context, SHALL use forward-slash resource identities, and MUST NOT expose canonical absolute paths, OS-specific separators, map order, locale-formatted values, or ANSI-dependent meaning. Repeated renders MUST be byte-identical, use exactly one trailing LF, and MUST NOT mutate the recursive result, graph, diagnostics, or contribution evidence.

#### Scenario: Equivalent checkouts
- **WHEN** equivalent tree results use different canonical checkout prefixes and Windows-style internal paths
- **THEN** their text output is byte-identical after holding the tool version and portable project identity constant

#### Scenario: Caller-owned graph
- **WHEN** text tree presentation receives graph edges in caller-owned storage
- **THEN** rendering leaves the original node, edge, diagnostic, and contribution order and values unchanged
