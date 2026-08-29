## Context

See [proposal.md](proposal.md) for motivation and scope. Recursive analysis currently caches one `ExpandedSummary` per canonical scene and applies its aggregate occurrence metrics at each mount. That gives correct root totals and parse reuse, but it discards the direct-scene attribution needed to explain those totals. The dependency graph already preserves canonical/display scene identities and edge multiplicity, while resource aggregation intentionally collapses declarations into unique closure identities.

The contribution contract must therefore preserve two different truths:

1. occurrence metrics have direct, additive scene-level sources; and
2. depth and unique-union metrics cannot be honestly explained as per-scene additive ownership.

Contribution selection is presentation-only. The application service must continue returning the same analysis result for text and JSON, and the default text bytes must not change. All portable output must use project-relative forward-slash identities and must not leak canonical checkout paths.

## Goals / Non-Goals

**Goals:**

- Derive contribution evidence during the existing cached recursive traversal rather than performing a second parse or analysis pass.
- Make the five additive occurrence metrics reconcile by construction and validate the invariant before publishing a result.
- Preserve immediate mount context while compacting equivalent occurrences with checked arithmetic.
- Represent maximum depth and unique resource/dependency unions without inventing additive owners.
- Keep analysis-domain data owned and presentation-independent, then project it deterministically into text and JSON.

**Non-Goals:**

- Do not expose the full dependency graph as the tree UI planned for #52.
- Do not assign a separate confidence value to every metric; #54 will build on row reliability and evidence.
- Do not alter parser behavior, root aggregate formulas, budget evaluation, or check exit priority.
- Do not add a shipped dependency or any runtime need for Godot, Node.js, OpenSpec, network access, or schema validation.

## Decisions

### 1. Store domain contributions separately from report wire objects

The analysis package will add owned contribution models to the successful recursive result. A row will retain:

- stable kind (`root`, `scene`, `inherited`, or `unresolved`);
- internal canonical identity plus portable display/original identity;
- immediate declaring scene and mount/resource context;
- checked occurrence multiplicity;
- direct additive values for nodes, scene instances, mesh instances, lights, and shadow lights;
- an optional root-relative depth candidate;
- row reliability and unresolved/inheritance classification where applicable.

The frozen eight-entry metric projection and JSON field layout will be built in the report package. Analysis models will not import JSON/schema concerns or represent meaningless unique values as zero.

Alternative considered: reuse `metrics.Values` directly for every row. Rejected because zero in `external_resources` or `scene_dependencies` would look like an owned additive value, and zero `tree_depth` cannot distinguish a real value from an unavailable candidate.

### 2. Attribute only direct values, and assign each mount occurrence to its target row

Each cached scene summary will contain a self row for literal nodes declared by that document. Its direct mesh/light/shadow values come from those same nodes. Nested expanded values are never copied into the parent self row.

When a non-inheritance mount is applied, the mounted target row receives one `scene_instances` contribution per occurrence. A resolved target receives its direct node values through its cached self row; an unresolved target receives exactly one known root node and one scene-instance value per occurrence. Descendant rows remain separate. Thus every mount and every supported literal node has exactly one additive owner, making reconciliation a structural property rather than an after-the-fact heuristic.

For inherited scenes, base rows are applied through an `inherited` context without adding a scene-instance occurrence. The derived scene's self row receives the incoming mount occurrence when the derived scene itself is nested. Base and derived rows are marked approximate because supported inheritance is deliberately not a full effective-scene merge.

Alternative considered: assign all expanded child totals to the immediate child row. Rejected because grandchildren would then either disappear or be double-counted when their own rows are shown.

Alternative considered: assign `scene_instances` to the declaring parent. Rejected because the top view is more actionable when the mounted target accounts for the cost of being instantiated, and target attribution makes unresolved rows reconcile naturally.

### 3. Cache relative rows and compose them at mounts

A one-occurrence cached summary will retain contribution rows relative to its scene root. Applying it at a mount will:

1. convert its self row to the direct mounted-scene context;
2. add the incoming scene-instance contribution;
3. multiply occurrence counts and additive values with checked arithmetic;
4. compose every known relative depth candidate at the parent mount using the existing checked depth rule; and
5. retain descendant immediate contexts while compacting equivalent row keys.

The compaction key will include kind, target identity or unresolved classification, declaring-scene identity, mount path, resource/raw target context, and local known mount depth. It will not include reachable ancestor path. Consequently, repeated and diamond paths reuse and aggregate the same immediate context, while distinct mounts remain distinguishable. When equivalent rows arrive through ancestors at different root depths, additive values and occurrences sum while the depth candidate keeps their checked maximum.

Alternative considered: materialize every full root-to-mount occurrence path. Rejected because repeated content can make output grow with total occurrence count, defeats useful compaction, and duplicates the dependency-tree concern planned for #52.

### 4. Compute direct depth candidates without treating depth as additive

The self row's relative depth candidate will be the maximum known depth of its directly attributed literal nodes. A mounted self row also considers the known mount/root depth because the mounted root exists even when it has no additional supported literals. An unresolved row uses its known mount depth. Descendant candidates are offset during recursive application. Inherited base candidates are applied through the existing approximation boundary.

Finalization will compare the maximum known contribution candidate with the published `tree_depth` only when depth evidence is complete. With unsupported parent evidence, known candidates remain useful but no exact depth reconciliation is claimed.

Alternative considered: put the entire expanded tree depth on each scene row. Rejected because those overlapping maxima obscure which row actually supplies the deepest known node and encourage summation.

### 5. Build unique-union evidence from graph/resource provenance, not contribution sums

The successful recursive result will also retain unique evidence for:

- external-resource identities, keyed by resolved canonical target or the existing unresolved declaration tuple; and
- scene-dependency identities, keyed by validated non-root graph node.

Each unique item will contain deterministic referrers derived from successfully parsed declarations or resolved graph edges. Internal canonical keys ensure correct deduplication; display/raw fields support portable reporting. A shared target appears once and lists every distinct referring scene/context. Finalization validates that unique-item cardinalities equal the authoritative root union metrics.

The report metric entries for these two IDs will use aggregation `unique_union` and omit a row value. Top selection rejects them before analysis because a ranked additive owner does not exist.

Alternative considered: divide one shared target fractionally or assign it to the first referring scene. Rejected because either changes integer metric semantics or makes ownership depend on traversal order.

### 6. Classify reliability at the row boundary

Self rows start exact. Unresolved/imported/unavailable rows are lower-bound. Rows representing inherited base or derived override evidence are approximate. Unsupported parent evidence makes affected depth-bearing rows lower-bound. When equivalent rows compact, reliability uses the same conservative precedence as whole analysis: approximate, then lower-bound, then exact.

Reliability belongs to the full row, not to individual metric entries in this change. This is deliberately conservative and leaves #54 free to refine individual metrics using the retained classifications without changing contribution identity or additive ownership.

### 7. Validate, clone, and sort before publication

Analysis finalization will reject invalid kinds, identities, negative counts, invalid reliability, additive mismatches, complete-depth mismatches, unique cardinality mismatches, and any checked arithmetic failure. Fatal failures return the existing zero `RecursiveResult` contract.

All cache writes and public returns will deep-clone contribution and referrer slices. Internal ordering will be deterministic by stable identity/context keys. Report projections will re-sort using portable fields so output does not depend on absolute checkout prefixes or platform separators.

### 8. Keep top selection in the CLI/report boundary

`inspect` will add paired local flags:

- `--metric` for one supported additive metric or `tree_depth`; and
- `--top` for a positive signed 64-bit limit.

Argument validation will run before the application call and require both flags together. The flags will be converted into a report selection in `report.Options`; they will not be added to `app.InspectRequest`. Check intentionally receives no top-selection flags.

Text rendering appends a section only when a selection exists. Ranking uses selected value descending and a complete portable tie-break key. Default text therefore follows the existing renderer without any conditional changes before its final byte.

JSON always contains full contribution and unique evidence for both inspect and check analysis payloads. An inspect top selection adds a projection descriptor and selected row identities while retaining the full collection. This makes JSON useful for failed budgets even though the interactive top UI remains inspect-only.

Alternative considered: filter the analyzer result to the requested top N. Rejected because presentation would affect domain semantics, JSON would lose auditable evidence, and the application request would diverge between formats.

### 9. Extend schema version one compatibly

The committed Draft 2020-12 schema will add definitions for contribution kinds, aggregation modes, contexts, nullable/omitted non-additive values, unique evidence, and optional top selection. The analysis object will require contribution and unique-evidence fields for new producers. No existing field, discriminator, enum meaning, or report kind changes.

Golden and schema tests will cover complete, lower-bound, approximate, shared-resource, failed-check, and selected-top documents. Portability tests will render equivalent results from two checkout roots and Windows-style source separators. Schema validation remains test-only.

## Risks / Trade-offs

- [Contribution rows increase memory and JSON size] → Compact equivalent immediate contexts, cache one-occurrence rows, avoid full root-to-leaf path materialization, and benchmark repeated/diamond fixtures.
- [Immediate context cannot identify every distinct ancestor chain] → Preserve declaring scene, mount path, multiplicity, and portable target now; #52 will expose the authoritative graph/tree and back-references.
- [Row-level reliability is conservative for some direct metrics] → Retain classifications and do not claim per-metric precision; #54 can refine confidence without changing row identity.
- [Resolved-resource provenance was previously discarded after union] → Capture referrers during the existing graph/resource traversal and validate unique cardinality against the established union.
- [Default text compatibility can regress through shared helpers] → Keep contribution rendering opt-in and retain byte-for-byte golden coverage for existing reports.
- [Version-one JSON becomes larger for all scene reports] → Use deterministic compact records and treat the fields as the compatible automation foundation required for 0.2 explainability.

## Migration Plan

1. Add analysis-domain contribution and unique-evidence models behind existing recursive APIs.
2. Populate and validate them while keeping established root metric tests unchanged.
3. Add paired inspect selectors and opt-in text rendering.
4. Extend JSON wire models and schema, then regenerate/update focused goldens.
5. Run compatibility, ownership, overflow, portability, race, lint, and full repository gates before archive.

Rollback is a normal PR revert: no persisted user data or configuration is migrated. Existing text behavior and JSON schema version remain valid throughout the change.
