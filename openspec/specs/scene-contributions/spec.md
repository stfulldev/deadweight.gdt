# scene-contributions Specification

## Purpose

Defines portable, deterministic scene-level attribution that explains occurrence metrics, maximum depth, and shared unique evidence without overstating static-analysis precision.

## Requirements

### Requirement: Contributions retain portable scene and mount identity
A successful recursive analysis SHALL publish owned contribution rows for the analyzed root, every resolved scene mount, every inherited scene contribution, and every unresolved scene mount. Each row SHALL retain its stable kind, portable scene or target identity, immediate declaring-scene identity, local mount path and known mount depth when applicable, and a checked non-negative occurrence count. Canonical absolute paths MAY be retained internally but MUST NOT appear in portable report fields.

#### Scenario: Repeated child mount
- **WHEN** one canonical child is reached repeatedly through the same declaring scene and mount context
- **THEN** its equivalent contribution row retains one portable child identity and the checked total occurrence count
- **AND** the child document is not reparsed for each occurrence

#### Scenario: Distinct mount contexts
- **WHEN** the same child scene is mounted through two distinct local mount paths or declaring scenes
- **THEN** the contribution evidence keeps the contexts distinguishable even though the canonical target is shared

#### Scenario: Unresolved target identity
- **WHEN** a mount target has no resolved in-project scene identity
- **THEN** its contribution row retains the declaring scene, mount path, resource reference or raw target, and stable unresolved classification
- **AND** no canonical target is invented

### Requirement: Additive occurrence contributions reconcile exactly
Contribution rows SHALL use checked signed 64-bit arithmetic and SHALL reconcile exactly to the root values for `nodes`, `scene_instances`, `mesh_instances`, `lights`, and `shadow_lights`. Direct supported literal nodes SHALL be attributed to the scene document that declares them after applying that document's occurrence multiplicity. Each non-inheritance scene mount, resolved or unresolved, SHALL attribute its one scene-instance occurrence to the mounted-target row; an unresolved mount SHALL also attribute its one known root node there. Nested metrics MUST NOT be counted both in a parent row and a child row.

#### Scenario: Three-scene chain
- **WHEN** A mounts B and B mounts C once
- **THEN** the additive values across the A, B, and C contribution rows sum exactly to the published root occurrence metrics
- **AND** the B and C rows each account for their incoming scene-instance occurrence

#### Scenario: Repeated and diamond dependency
- **WHEN** one cached descendant is reached with multiplicity through repeated or diamond-shaped paths
- **THEN** its direct additive contribution is multiplied by the checked reachable occurrence count
- **AND** the sum of all rows includes every occurrence exactly once without reparsing or double-counting the cached summary

#### Scenario: Unresolved instance root
- **WHEN** a scene mount cannot be expanded statically
- **THEN** its unresolved contribution adds exactly its occurrence count to `nodes` and `scene_instances`
- **AND** it adds zero to unsupported unknown inner metrics

### Requirement: Maximum and unique-union evidence is explicitly non-additive
Every contribution projection SHALL retain the frozen metric order `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, `shadow_lights`, `external_resources`, `scene_dependencies` and SHALL identify the aggregation mode of each entry. `tree_depth` SHALL expose a known root-relative maximum candidate when available and the root depth SHALL equal the maximum known candidate rather than a sum. `external_resources` and `scene_dependencies` SHALL be marked `unique_union` with no per-row additive value. Unique evidence SHALL retain each portable union identity once together with a deterministic set of referring scene identities or contexts, so a target shared by several scenes is visible without assigning exclusive ownership.

#### Scenario: Deepest candidate wins
- **WHEN** several contribution rows contain known root-relative depth candidates
- **THEN** the published root tree depth equals the maximum candidate
- **AND** summing depth candidates is neither required nor presented as meaningful

#### Scenario: Shared external resource
- **WHEN** two parsed scenes declare resources that resolve to the same canonical target
- **THEN** unique evidence contains one portable resource identity and both referring scene identities
- **AND** no contribution row claims an additive resource count for that shared target

#### Scenario: Shared scene dependency
- **WHEN** a diamond graph reaches one canonical descendant through two declaring scenes
- **THEN** unique dependency evidence contains the descendant once with both referring contexts
- **AND** the root dependency metric remains the authoritative unique-union count

### Requirement: Contribution reliability is evidence-based
Each contribution row SHALL expose `exact`, `lower_bound`, or `approximate` row reliability as the conservative summary of confidence entries for all eight frozen metrics. Every metric entry SHALL expose its own reliability and deterministic machine-readable reasons derived from evidence that can affect that metric in that row. Unsupported or unavailable nested content MUST make potentially missing contribution values or candidates lower-bound rather than silently treating their unknown remainder as zero; unknown parent composition MUST make the depth candidate lower-bound without degrading unrelated additive values; and inherited roots or override evidence MUST make applicable metrics approximate. Approximate SHALL win when both approximate and lower-bound reasons affect the same metric. Unique-union entries SHALL expose confidence without inventing a per-row owned value.

#### Scenario: Imported child
- **WHEN** an imported scene mount contributes only a known unresolved root occurrence
- **THEN** potentially missing metrics on that unresolved row are `lower_bound` with an imported-scene reason and cannot appear exact
- **AND** its known nodes and scene-instance occurrences remain present

#### Scenario: Inherited contribution
- **WHEN** a scene contribution includes the supported inherited-base approximation or override evidence
- **THEN** applicable metric entries and the conservative row reliability are `approximate` even when every referenced path resolves

#### Scenario: Unrelated exact row
- **WHEN** one resolved scene's direct contribution is fully supported and another row is partial
- **THEN** the supported row's metrics may remain exact while the affected row and whole analysis retain their conservative classifications

#### Scenario: Unknown parent depth
- **WHEN** a row has exact additive values but its root-relative depth cannot be composed because of unsupported parent evidence
- **THEN** only its depth confidence is lower-bound and no exact depth candidate is invented

### Requirement: Top contributors are selected explicitly and deterministically
The top-contributors projection SHALL require an explicit supported metric and positive limit. It SHALL support the five additive occurrence metrics and the maximum `tree_depth` candidate, and MUST reject `external_resources` and `scene_dependencies` with an actionable non-additive explanation. Rows SHALL sort by selected value descending, then portable scene identity, declaring-scene identity, mount path, stable kind, and remaining context in ascending byte order; absent depth candidates SHALL sort after known candidates. The limit SHALL truncate only after this total ordering and SHALL NOT mutate the authoritative contribution collection.

#### Scenario: Stable tie breaking
- **WHEN** several rows have the same selected additive value
- **THEN** repeated projections on supported operating systems return the same portable identity and context order

#### Scenario: Limit exceeds row count
- **WHEN** the positive limit is larger than the number of eligible contribution rows
- **THEN** every eligible row is returned once in deterministic order

#### Scenario: Unique metric selection
- **WHEN** a user selects `external_resources` or `scene_dependencies` for top contributors
- **THEN** selection fails before rendering with guidance that the metric is a shared unique union rather than additive ownership

### Requirement: Contribution results are checked, deterministic, and owned
Every contribution occurrence, additive multiplication or sum, root-relative depth composition, and unique-evidence cardinality SHALL use the recursive checked-arithmetic contract and fail with `SB2004` on invalid negative input or overflow. Returned contribution and unique-evidence collections SHALL be deeply owned, sorted deterministically, and independent of cached one-occurrence summaries.

#### Scenario: Contribution multiplication overflows
- **WHEN** applying a cached direct contribution at its reachable multiplicity exceeds signed 64-bit range
- **THEN** analysis fails with `SB2004` and publishes no partial root result or contribution collection

#### Scenario: Caller mutates returned rows
- **WHEN** a caller mutates contribution metrics, contexts, or unique referrer slices returned by one invocation
- **THEN** cached summaries and later invocations retain their original deterministic values
