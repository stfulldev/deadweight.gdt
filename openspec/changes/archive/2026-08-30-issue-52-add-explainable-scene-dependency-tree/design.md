## Context

See `proposal.md` for motivation. Recursive analysis already returns an owned `DependencyGraph` containing every resolved and unresolved instance or inheritance edge, compacted occurrence counts, stable classification evidence, cycle validation, and normalized display paths. The command/report layers already provide shared project discovery, recursive analysis, text/JSON selection, portable-path helpers, deterministic diagnostics, and a version-one JSON envelope. The change must expose that evidence without another filesystem pass, without changing graph aggregation, and without making a diamond graph grow as if it were a tree.

## Goals / Non-Goals

**Goals:**

- Add an injected `tree` application flow that shares the scene-analysis pipeline with `inspect` while remaining independent of policy and budgets.
- Build one validated, portable, non-mutating projection that both text and JSON renderers consume.
- Guarantee that every authoritative graph edge appears once and that shared targets expand once through deterministic back-references.
- Keep incomplete and inherited evidence visible with conservative row reliability and existing whole-analysis diagnostics.
- Extend JSON schema version one compatibly and preserve every existing command and report meaning.

**Non-Goals:**

- Changing graph discovery, recursive metric aggregation, contribution attribution, or cycle detection.
- Calculating a globally multiplied occurrence count for a transitive path; each row reports its authoritative compacted declaring-edge multiplicity and the nesting explains composition.
- Emitting DOT, Graphviz, HTML, interactive output, or an unbounded nested JSON graph.
- Resolving UID targets, parsing imported scenes, implementing full inherited overrides, or scanning more than one root.

## Decisions

### 1. Share one scene-analysis application pipeline

The application package will factor the common project/configuration/scene/recursive work used by `Inspect` into one private operation. `Inspect` and a new `Tree` method will return distinct named request/result types backed by the same analysis effects. The injected CLI `Application` interface gains `Tree`, allowing command tests to prove that `tree` invokes exactly one intended flow rather than accidentally routing through policy or budget work.

Calling the existing `Inspect` method directly from the command was considered. It would reuse work but blur the command boundary in fakes, telemetry, and future evolution, so an explicit application method is preferred.

### 2. Keep the authoritative graph unchanged and project in the report layer

The graph model already contains all required edge evidence. A report-domain projection will validate graph/root/node/edge invariants, normalize portable identities, build outgoing adjacency from owned copies, and produce presentation rows. Text and JSON use that same row collection, preventing divergent traversal or back-reference behavior.

Adding a second analysis tree model was rejected because it would duplicate graph truth, require another ownership contract, and risk inconsistent edge counts. Reparsing scenes in the report or CLI layer is forbidden.

### 3. Use a flat portable preorder with expand-once semantics

Projection starts at the canonical root, orders each source's outgoing edges by portable target identity, edge kind, resolution state, resource/raw/classification context, and remaining stable fields, then performs depth-first traversal. Each graph edge produces one row. The first resolved edge to a canonical target marks that node expanded and recursively emits its outgoing edges. A later edge to the same target sets `back_reference=true` and stops at that row. Unresolved edges are leaves. The projector verifies that all graph edges were consumed and all non-root nodes were reached.

Fully nested output was considered, but recursive schema and duplicated diamond subtrees make it larger and harder to compare. A purely node-sorted flat graph was also considered, but it loses the explanatory root-to-child reading order. Preorder rows retain tree readability, have a simple bounded schema, and can reproduce connector indentation from depth.

### 4. Sort with portable identities and reject ambiguous portable nodes

Canonical paths remain internal lookup keys. Every emitted identity is derived through existing project-relative helpers and uses forward slashes. Sibling order is based on the emitted portable fields, not checkout prefixes. Two canonical nodes that collapse to the same portable identity make projection fail instead of introducing canonical tie-breaking that could leak platform-specific order.

Trusting the graph's existing canonical sort was rejected because canonical checkout prefixes and platform separators are not a portable presentation contract.

### 5. Derive conservative edge reliability from retained edge semantics

A resolved instance edge is `exact` as dependency evidence. An unresolved non-inheritance edge is `lower_bound`, because its unavailable subtree may contain further dependencies. Every inheritance edge is `approximate`, matching the existing inheritance boundary whether its base resolves or not. The document also retains whole-analysis status/reliability and diagnostics, so an exact edge inside a partial result is not mistaken for an exact whole tree.

Per-metric confidence is deliberately not inferred here; issue #54 owns that contract.

### 6. Reuse the existing analysis JSON payload and add `dependency_tree`

`TreeJSON` will build the same portable analysis payload used by inspect/check, change the discriminator to `tree`, and attach a required dependency-tree object with root plus preorder entries. Fatal JSON continues to use the existing `error` kind. The Draft 2020-12 schema adds the new kind and definitions without changing established kind schemas or `schema_version`.

A minimal tree-only analysis summary was considered, but it would create a second representation of status, diagnostics, coverage, contributions, and unique evidence. Reuse preserves one automation contract and makes partial-tree evidence self-contained.

### 7. Text output uses deterministic tree connectors without semantic color

Text renders the standard version/root/project/status/accuracy header, then one root line and preorder edge lines with Unicode tree connectors, kind, checked multiplicity, reliability, target, and optional back-reference or unresolved classification. Connector state is computed from row depth and sibling-last metadata retained only by the projection. Diagnostics and partial warnings reuse established report semantics. ANSI may emphasize established statuses when allowed but never changes tree meaning or golden bytes.

ASCII-only indentation was considered, but explicit tree connectors make ancestry materially easier to scan while remaining deterministic UTF-8 on every supported Go platform.

## Risks / Trade-offs

- [A flat preorder is less convenient than nested JSON for some consumers] → Include depth, source, target, and back-reference fields so consumers can reconstruct hierarchy without recursive parsing.
- [A graph edge's local multiplicity can be mistaken for total reachable path multiplicity] → Document it as declaring-edge multiplicity and preserve parent nesting; do not invent a multiplied total absent from the authoritative graph.
- [Portable path normalization could collapse invalid injected identities] → Validate portable node uniqueness and fail before output rather than fall back to canonical order or paths.
- [Unicode connectors can render differently by font] → Keep semantics in explicit labels and cover exact UTF-8 bytes on Linux, macOS, and Windows CI.
- [Adding `Tree` widens injected interfaces] → Update all fakes at compile time and keep the application implementation a thin wrapper around shared scene analysis.
- [JSON v1 growth increases successful tree document size] → Reuse existing owned evidence deliberately; keep the tree itself bounded to one row per graph edge.

## Migration Plan

1. Add the explicit application tree flow and CLI command without changing existing command behavior.
2. Add the shared projection and text renderer behind the new command.
3. Add JSON kind `tree`, schema definitions, and portable golden fixtures.
4. Run all repository gates and strict OpenSpec validation, archive the change, and merge through linked PR #61.

Rollback is removal of the new command, report entry points, and compatible schema kind; no stored data, configuration migration, or external runtime state is introduced.
