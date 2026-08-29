## Context

See `proposal.md` for motivation and `specs/recursive-scene-expansion/spec.md` for the behavior contract. The repository now has three prerequisite boundaries: `project.Resolver` produces canonical/display/original resource identities and typed unresolved reasons; `tscn.Parse` produces the supported format-3 document or a typed `SB2001` failure; and `analysis.BuildLocalSummary` separates ordinary nodes, instance mounts, inherited roots, override stubs, local metric contributions, resource declarations, and optional mount depths.

Recursive expansion must consume those boundaries without reopening raw node classification. Issue #9 has intentional seams with later roadmap issues: it needs invocation-local memoization and checked arithmetic to make repeated expansion correct, but issue #10 owns the explainable graph and cycle diagnostic, issue #11 hardens cache observability/coverage, issue #12 publishes the final eight-metric result, and issues #13–#14 own diagnostic/reliability and inheritance policy.

## Goals / Non-Goals

**Goals:**

- Add a reusable recursive analyzer whose only external effects are supplied resource resolution and scene-loading functions.
- Represent one expanded scene occurrence independently from the number and depth of mounts where callers later apply it.
- Preserve enough deterministic structured evidence for graph, coverage, reliability, and diagnostics without assigning those later policies prematurely.
- Keep production and tests in separate focused commits while maintaining green commits.

**Non-Goals:**

- Add CLI composition, report models, budget verdicts, persistent caches, goroutine-based traversal, or parallel filesystem reads.
- Compact or render the final dependency graph, or expose the final user-facing cycle chain and exit behavior.
- Treat inherited scenes as exact recursive children; inherited-root evidence remains explicit for issue #14.

## Decisions

### 1. Extend `internal/analysis` with a resolver-and-loader driven engine

Add a narrow resource-resolver interface matching `project.Resolver.ResolveResource(fromScene, raw)` and a scene-loader function that accepts a `project.ResolvedPath` and returns a parsed `*tscn.Document` or error. A recursive analyzer value stores these dependencies, while each public expansion call allocates fresh memoization state. The root input is a fully resolved canonical `project.ResolvedPath`; the loader handles the root and every child uniformly.

The production package will not call `os.Open`, discover projects, or parse process arguments. Tests can use an in-memory loader and a fake resolver for exact chain/diamond counts, plus a small real temporary project to verify compatibility with `project.Resolver`. A later application service can compose an OS opener with `tscn.Parse` without changing recursive policy.

Alternative considered: accept raw paths and construct a project resolver inside analysis. Rejected because it would duplicate issue #7 policy and make declaring-scene-relative resolution hard to test.

Alternative considered: accept prebuilt local summaries only. Rejected because recursive targets are discovered during traversal and canonical load/parse reuse is an issue #9 acceptance criterion.

### 2. Keep expanded counters, unique identities, and unresolved evidence separate

Introduce an `ExpandedSummary` representing one occurrence of a canonical scene:

- a `metrics.Values` containing only the six occurrence/maximum fields (`ExternalResources` and `SceneDependencies` remain zero until issue #12 derives final values from sets);
- sorted external-resource identities and canonical dependency paths;
- resolved and unresolved scene-instance coverage counts;
- sorted unresolved instance records carrying an occurrence count;
- scene-scoped local parent findings and inherited-root evidence;
- a depth-partial flag.

A resource identity is either resolved (`Canonical` is the uniqueness key) or unresolved (`DeclaringScene`, `ResourceID`, and `RawPath` form the key). An unresolved instance record includes a finite target classification, `project.ResolutionReason` when applicable, declaring path, resource reference, raw target, mount path/depth/position, and occurrence count. These records remain lower-level evidence; issue #13 maps them to stable user diagnostics and reliability.

Alternative considered: store maps in the public result. Rejected because callers could mutate cached state and output would inherit map nondeterminism. Internal maps build sets; returned slices are owned and sorted.

Alternative considered: set the unique metric fields immediately. Rejected because issue #10 owns graph-backed dependency counting and issue #12 owns the complete public eight-metric collection; retaining exact sets avoids both premature publication and information loss.

### 3. Classify targets from local evidence before loading

For every `InstanceMount`, look up its referenced ID in `LocalSummary.ExternalResources` when it is an external reference. Classification order is:

1. placeholder and `SubResource` mounts become unresolved without resolver calls;
2. a missing external-resource declaration becomes unresolved with its ID;
3. every present declaration is resolved through the supplied secure resolver using the declaring scene's canonical path;
4. a path-resolution failure becomes unresolved with the stable resolution reason;
5. a resolved exact-lowercase `.tscn` target is loadable regardless of declared type;
6. a resolved `.glb`, `.gltf`, `.blend`, or `.scn` target is imported/binary unresolved evidence;
7. every other extension is unsupported instance evidence.

All declarations, including the one used by a scene mount, also contribute a resource identity. Resolution is performed from the declaring scene, so a child summary never inherits its parent's relative-path base.

When a loaded document has an inherited root, this slice preserves inherited-scene evidence and one known mounted root rather than claiming an exact expanded child. The inheritance edge, base aggregation, explicit additions, and approximate reliability remain for issue #14. A non-inherited malformed/format-unsupported `.tscn` returns its typed parse failure; a non-parse loader failure after successful metadata resolution is converted to unavailable unresolved evidence because an unreadable nested scene is nonfatal in the MVP matrix.

Alternative considered: trust `MountPackedSceneCandidate` and ignore other external mount kinds. Rejected because the MVP requires an existing `.tscn` with an incompatible declaration type to be attempted as a scene.

Alternative considered: classify by declared type alone. Rejected because imported scenes and wrong-type `.tscn` declarations are distinguished by the resolved target and extension.

### 4. Use per-invocation document, summary, and in-progress maps

The invocation state has canonical-path keyed document and expanded-summary maps plus an in-progress set. On the first identity, load and build its local summary, record the owned document, mark the identity in progress, recursively expand it, then cache the completed one-occurrence summary. Later repeated or diamond occurrences reuse the cached summary.

The in-progress set is a safety boundary: encountering an identity already being expanded returns a typed internal recursive-reference error and no summary, preventing hang or stack exhaustion. Issue #10 will replace this minimal signal with graph nodes/edges, full display-chain reconstruction, and `SB2002`; issue #9 tests only acyclic chain/diamond/repeated closures plus termination safety.

Caching failures is unnecessary in this slice: an invocation stops on fatal errors, while unresolved target evidence is part of the successfully cached parent summary. State is allocated for one call and discarded afterward.

Alternative considered: globally cache documents across analyzer calls. Rejected because the CLI is read-only and short-lived, and cross-invocation invalidation is explicitly outside MVP 0.1.

Alternative considered: launch recursive goroutines. Rejected because deterministic errors, simple recursion state, and small local project files matter more than speculative parallelism; concurrent cache coordination belongs outside this slice.

### 5. Apply child summaries through one checked multiplicity path

Start each expanded summary from the local six-field metrics. `LocalSummary.SceneInstances` already counts every non-inherited mount header once. Classify mounts and compact resolved applications by canonical child identity plus optional mount depth. For each group with occurrence count `N`:

- add `N * child.Nodes`, `MeshInstances`, `Lights`, and `ShadowLights`;
- add `N * child.SceneInstances` because the `N` parent mount occurrences are already local;
- add `N * (1 + child.ResolvedSceneInstances)` to resolved coverage and `N * child.UnresolvedSceneInstances` to unresolved coverage;
- scale each child unresolved/finding occurrence by `N` while retaining its source identity;
- union child resources/dependencies and add the child canonical path once;
- compare, never multiply, `mountDepth + child.TreeDepth - 1` when both depths are known.

Each unresolved local mount adds one node and one unresolved coverage occurrence; its scene-instance occurrence is already in the local metric. Known unresolved mount depths are already present in local tree depth. Unknown mount or child depth sets the depth-partial flag without fabricating a value.

All increments, products, and depth composition call shared checked helpers. An `OverflowError` records operation and operands, implements the diagnostic coded-error contract with `SB2004`, and aborts without returning a summary. The helpers accept only non-negative operands, making a negative intermediate an internal invariant failure rather than a valid result.

Alternative considered: loop once per occurrence and only use checked addition. Rejected because a single multiplicity path is both faster for repeated mounts and exercises the same overflow semantics later graph aggregation needs.

Alternative considered: add one node for every mount before replacing it with a child. Rejected because a resolved mount header is the child root and would be double-counted.

### 6. Canonicalize deterministic results at the cache boundary

Before caching an expanded summary, convert internal resource/dependency sets and evidence maps into owned sorted slices. Sorting keys use canonical/display declaring scene, classification/reason, resource ID/raw target, mount path, source line/column, and occurrence count only as a final stable tie-breaker. Applying a cached summary clones or merges its values; callers never receive aliases to cache-owned slices or candidate pointers.

Tests will verify exact chain, diamond, repeated ×100, unresolved target table, incompatible-type `.tscn`, unique union, mount-depth composition, inherited deferral, fatal nested parse error, invocation reset, recursive-reference termination, and overflow. Production code will be committed separately from `*_test.go` plus OpenSpec progress.

Alternative considered: rely on traversal order alone. Rejected because resource maps, future edge compaction, and diamond memoization can otherwise leak nondeterministic ordering into later reports.

## Risks / Trade-offs

- [Issue #9 memoization overlaps issue #11 cache hardening] → Implement only the minimal invocation maps required for correct repeated expansion; leave public cache statistics, parsed-file coverage derivation, broader instrumentation, and cache-specific error tests to issue #11.
- [A loaded inherited child cannot be exact yet] → Preserve explicit inherited evidence and one known mounted root without base expansion; issue #14 modifies recursive behavior after graph and completeness infrastructure exists.
- [Resolver metadata can succeed and file loading can still fail] → Treat non-parse loader failures as unresolved unavailable evidence with one known root; keep typed parser failures fatal.
- [Compacting resolved mounts can hide source locations] → Compact only summary application by target/depth; unresolved and local source evidence remains per source identity with checked occurrence counts.
- [The temporary recursive-reference error is less informative than `SB2002`] → Return no summary and guarantee termination now; issue #10 owns the stable graph-backed error and full chain.
- [Public summary slices can be mutated by callers] → Clone at cache/application boundaries and return owned slices; tests mutate returned summaries to prove later cache hits are unchanged.

## Migration Plan

1. Add recursive models, resolver/loader seams, checked arithmetic, target classification, and invocation state without changing existing consumers.
2. Add recursive expansion and one-occurrence summary application for chain and unresolved cases, then repeated/diamond compaction and unique unions.
3. Commit production-only files after existing build/test/vet gates pass; add all issue #9 tests in a separate test commit.
4. Run targeted/full/race/vet/lint and strict OpenSpec validation, archive the capability, and keep CLI placeholders unchanged.

Rollback is additive: revert the test commit, production commit, and archived specification. No persistent cache, configuration, user data, public CLI output, or external Go API requires migration.
