## Context

See `proposal.md` for motivation and `specs/local-scene-summary/spec.md` for the behavior contract. `internal/tscn` already returns a minimal in-memory `Document` containing ordered node headers, a map of external-resource declarations, source positions, literal types, instance references, placeholders, and `shadow_enabled`. `internal/metrics` owns the frozen eight-field metric value type. No analysis package yet interprets node parent paths or separates local contributions from child-scene expansion.

The local layer must become the stable handoff between parsing and the later graph/cache slice. It must be deterministic, must not import project/filesystem policy, and must preserve enough evidence for later resolution and whole-analysis diagnostics without claiming that a `PackedScene` declaration has resolved successfully.

## Goals / Non-Goals

**Goals:**

- Introduce a small analysis-domain model that retains ordinary nodes, mounts, inheritance/override evidence, external-resource declarations, known depths, and local occurrence counters.
- Make classification and parent validation explicit enough that later DFS aggregation does not inspect raw TSCN nodes again.
- Keep every returned collection deterministic and every unknown depth distinguishable from numeric depth zero.
- Keep the production-only and test-only delivery commits separate while leaving both commits buildable.

**Non-Goals:**

- Define the recursive `SceneSummary`, graph, cache, resolution-to-diagnostic mapping, or checked multiplicity arithmetic needed by later issues.
- Add a new user-visible diagnostic code solely for local parent findings; the frozen diagnostic catalog has no parent-specific code, so whole-analysis policy will map structured evidence at its orchestration boundary.
- Change parser validation or make local analysis accept arbitrary malformed ASTs as if they came from a successful parse.

## Decisions

### 1. Add the local model to a focused `internal/analysis` package

Create the analyzer-facing types and builder in `internal/analysis`. The package consumes `*tscn.Document` and uses `metrics.Values` for the six known occurrence/depth contributions. `ExternalResources` and `SceneDependencies` remain zero in that value because final uniqueness depends on canonical path resolution and graph closure; complete local declaration records are stored separately.

The summary will expose source-ordered ordinary-node, mount, and override-stub records; an optional inherited-root record; ID-ordered external-resource records; deterministic local findings; a partial-depth flag; and the local metrics. Depth fields use an optional representation so an unsupported parent cannot be confused with the root or a zero value.

The metrics semantics at this boundary are:

- `Nodes`: ordinary local typed headers only;
- `TreeDepth`: maximum known depth among ordinary nodes and non-inherited mounts;
- `SceneInstances`: non-root instance and placeholder mount occurrences;
- `MeshInstances`, `Lights`, and `ShadowLights`: literal ordinary-node contributions;
- `ExternalResources` and `SceneDependencies`: zero, with their unresolved local evidence carried in dedicated records.

Alternative considered: put tree construction in `internal/tscn`. Rejected because parsing should report syntax, while ordinary/mount/inherited classification and metric semantics are analysis policy.

Alternative considered: define a second six-field metric struct. Rejected because it would duplicate frozen metric names and make later aggregation translate between structurally identical counters. Naming and documentation will make the zero unique fields explicit.

### 2. Classify by semantic precedence before calculating depth

Walk `Document.Nodes` once in source order and apply this precedence:

1. the root with an instance reference is an inherited root anchor;
2. a non-root placeholder is a placeholder mount;
3. a non-root instance reference is an instance mount;
4. a node with a non-empty literal type is ordinary;
5. a remaining node is an override stub.

Every record retains name, raw parent, position, and its classification-specific evidence. A non-root external instance is a candidate only when its ID matches an external-resource declaration whose literal type is exactly `PackedScene`; its raw path is copied into the mount evidence but not resolved. Missing IDs, other resource types, `SubResource`, and placeholders remain explicit non-candidate mount kinds for later partial handling. The inherited root is never added to `SceneInstances`, and no mount is added to `Nodes`.

Alternative considered: count every node header locally, then subtract resolved instance roots during aggregation. Rejected because subtraction complicates unresolved cases and is exactly how double-counting bugs enter repeated expansion.

Alternative considered: classify any external reference ending in `.tscn` as a resolved candidate. Rejected because local classification has a parsed literal resource type and must not infer type or filesystem success from an untrusted suffix.

### 3. Build path identities first, then resolve depths

Use two passes so forward-declared parents behave like parent-first documents.

The first pass validates the serialized path grammar and derives each non-root record's local path: `parent="."` yields `name`, otherwise `parent + "/" + name`. Supported non-root parent strings are nonempty relative slash-separated segments with no empty, `.` or `..` segment; only the exact `.` marker names the root. A leading slash is absolute and unsupported. The root receives the internal path `.` and acts as a depth-1 anchor even when it is inherited, allowing explicit local additions to have known positions without counting the inherited root locally.

Index path identities to all records rather than one record so duplicates are detected as ambiguous. The second pass resolves depth from the root anchor through exact parent identities. A missing or ambiguous parent seeds a structured finding and unknown depth. Nodes below an unknown-depth parent remain unknown without generating a misleading independent missing-parent finding. Known depths are parent depth plus one, and the maximum includes ordinary nodes and non-inherited mounts only.

Local findings have a finite reason kind, node name/path, raw parent, and TSCN source position. They are diagnostic evidence and set `DepthPartial`; they do not use `diagnostic.Diagnostic` yet because assigning an unrelated frozen `SB100x` code would corrupt its documented meaning. Findings are sorted by source position and stable tie-breakers. Later whole-analysis work can translate them together with scene identity and final reliability policy.

Alternative considered: derive depth solely from the number of parent segments. Rejected because it would silently accept missing parents and cannot detect ambiguity.

Alternative considered: require parents to precede children. Rejected because depth is a property of the serialized tree, and the issue requires deterministic construction rather than an incidental parser-order constraint.

### 4. Preserve detailed records as the aggregation handoff

Ordinary-node records retain literal type, shadow property, path, optional depth, and position. Mount records retain kind, reference kind/ID or placeholder target, candidate resource metadata, path, optional depth, and position. External-resource records clone the parser values into an ID-sorted slice, avoiding aliasing and map iteration. Override and inherited records retain enough source evidence for later approximate-inheritance policy.

The summary does not retain a pointer to the mutable input document. Returned slices and nested reference data are owned by the summary so callers can cache or compare it without parser-map ordering or pointer aliasing surprises.

Alternative considered: keep only counters and re-read `tscn.Document` during DFS. Rejected because it would duplicate classification policy in graph code and make cache identity depend on retaining parser internals.

### 5. Verify pure local behavior with table-driven in-memory tests

Tests will construct or parse small in-memory documents and cover ordinary root counting, multi-level and forward-declared parents, missing/ambiguous parents, `..`, absolute paths, placeholder and candidate mounts, inherited roots, override stubs, literal mesh/light/shadow rules, duplicate raw resource targets, and repeated deterministic output. A dependency-boundary test or direct package audit will ensure production code imports neither `internal/project` nor filesystem/process packages.

Production files will be committed first after existing repository tests pass; test files will follow in a separate commit and exercise the full contract. This keeps both requested commits independently buildable and reviewable.

Alternative considered: use filesystem fixtures. Rejected because this layer consumes a parsed document and filesystem use would blur the boundary established by issue #7.

## Risks / Trade-offs

- [Using `metrics.Values` leaves two fields intentionally zero] → Name the field as local/without-unique contributions, document the invariant, and test that resource records do not produce premature unique counts.
- [Children below an invalid parent have unknown depths without one finding each] → Retain optional depth on every record and attach the actionable finding to the first invalid or ambiguous relationship, avoiding cascaded noise while preserving partial state.
- [A `PackedScene` candidate name can be mistaken for resolution] → Use candidate terminology and retain only declaration evidence; require the project resolver and graph layer to establish target identity later.
- [Inherited root anchors make local child depths known while base depth is not yet aggregated] → Do not count the anchor as a local node or tree-depth contribution; later inheritance aggregation combines base depth and explicit additions under approximate reliability.
- [Duplicate local paths are unusual but can make parent lookup non-deterministic] → Index all matching records, mark the identity ambiguous, and never choose based on document order.

## Migration Plan

1. Add the local analysis models, classification, deterministic resource extraction, and two-pass depth builder without changing existing consumers.
2. Commit production-only files and verify the existing build/test/vet gates remain green.
3. Add the focused contract tests in a separate test commit and run full test, race, vet, lint, and OpenSpec validation gates.
4. Leave CLI placeholders and recursive analysis untouched; issue #9 and later graph work can consume the new summary model additively.

Rollback is a normal revert of the test and production commits plus this change's archived specification. No stored data, configuration, public CLI contract, or external Go API requires migration.
