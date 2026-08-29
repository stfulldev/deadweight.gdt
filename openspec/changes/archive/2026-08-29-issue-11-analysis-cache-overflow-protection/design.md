## Context

See `proposal.md` for motivation and the `recursive-scene-expansion` delta for the behavioral contract. Issue #9 introduced one-occurrence recursive summaries and checked occurrence aggregation; issue #10 added a graph phase that now shares document, local-summary, resolution, and expanded-summary maps through `invocationState`. The remaining #11 gap is architectural: the injected `SceneLoader` combines physical open and parse, cache ownership is spread across the state object, and `RecursiveResult` does not expose successful parsed-document cardinality.

The analyzer already depends on `project.ResolvedPath` and owned `tscn.Document` values, while secure path resolution remains in `internal/project`. All effects and caches must remain in memory for one `Analyze`/`Expand` call, deterministic, single-threaded, and independent of Godot, network access, OpenSpec, or Node.js.

## Goals / Non-Goals

**Goals:**

- Make independent physical-open and parse effects injectable and instrumentable.
- Give one invocation-owned cache a single canonical-path key contract across graph discovery and expansion.
- Expose successful parse-cache cardinality as checked `parsed_scene_files` coverage on `RecursiveResult`.
- Preserve one-occurrence summary ownership while applying occurrence counters and evidence multiplicity separately.
- Make every non-negative count mutation visibly pass through checked arithmetic and return typed `SB2004` failures transactionally.
- Preserve deterministic graph, identity, evidence, and error behavior established by #9 and #10.

**Non-Goals:**

- Persistent, incremental, cross-invocation, watch-mode, content-hash, or timestamp invalidation.
- Parallel scene opening or parsing, cache locking, cancellation, or benchmark-driven concurrency.
- Publishing final unique metrics, final `Analysis` reliability/status, inherited effective-tree metrics, diagnostics rendering, budgets, reports, or CLI composition.
- Changing secure resolver semantics, the TSCN subset parser, cycle selection, or unresolved-target classifications.

## Decisions

### 1. Replace the combined loader with independent opener and parser effects

`RecursiveAnalyzer` will accept an opener shaped around a resolved canonical scene and a parser shaped around an `io.Reader` plus stable display/source identity. The opener returns an `io.ReadCloser`; the cache layer owns exactly one close after the single parse attempt. Tests can therefore count opens and parses independently, inject open failures separately from typed parser failures, and use `io.NopCloser` for in-memory fixtures.

The internal constructor will require resolver, opener, and parser and validate all three. Repository call sites migrate atomically because `internal/analysis` is not a public module API. `tscn.Parse` remains the production parser function passed by composition code rather than being hidden in analysis.

Alternatives considered:

- Keep `SceneLoader`: rejected because one callback cannot prove whether a repeated call reopened, reparsed, or returned a prebuilt document.
- Inject only a byte-reader callback and call `tscn.Parse` directly: rejected because parser instrumentation and parser-failure cache tests would remain indirect.
- Add analyzer options around the existing constructor: rejected as unnecessary compatibility machinery for an internal API with no shipped caller yet.

### 2. Centralize invocation memoization behind one cache owner

Introduce an invocation-only cache type in `internal/analysis/cache.go`. It owns maps for successful documents, memoized document failures, local summaries, per-scene resource resolutions, and one-occurrence expanded summaries. `invocationState` retains traversal-only state: DFS colors/stack, expansion in-progress invariants, and the analyzer reference.

All cache keys are clean canonical absolute `.tscn` paths already validated or produced by the secure resolver. Display/original paths remain payload evidence and never become alternate keys. The document load path is the only place allowed to call opener or parser:

1. return a cached document or cached failure;
2. open once;
3. parse once using the normalized display identity;
4. close once;
5. cache either the successful document or stable typed/wrapped failure.

Typed `SB2001` parse errors remain fatal. Non-parse open/read failures retain `sceneLoadError` classification so downstream reachable targets remain unresolved according to the existing contract. A successful document is inserted only after parse and close complete, so unsuccessful targets cannot inflate parsed coverage.

Alternatives considered:

- Keep independent maps directly on `invocationState`: workable but leaves cache cardinality and effect ownership implicit, making later completeness integration fragile.
- Cache raw bytes in addition to parsed documents: rejected because it duplicates memory without an MVP consumer.
- Persist successful parses across invocations: rejected by the frozen invocation-only decision and unspecified concurrent file mutation behavior.

### 3. Report parsed coverage from the successful document cache

Add `ParsedSceneFiles int64` to `RecursiveResult`. After graph validation and recursive expansion both succeed, the analyzer converts the successful document-cache cardinality with a checked cardinality helper and assigns it to the result. The root counts once; every parsed reachable instance or inheritance scene counts once; repeated/diamond paths do not multiply it. Failed analysis still returns the zero result.

`ExpandedSummary` remains a one-occurrence reusable value and does not gain a global parsed-file field. This prevents child-summary application from accidentally multiplying cache cardinality. The summary-only `Expand` compatibility projection continues to return only `ExpandedSummary`; final coverage consumers will use `Analyze` in the later completeness integration slice.

Alternatives considered:

- Add parsed files to `SceneInstanceCoverage`: rejected because that structure is intentionally occurrence-composable.
- Derive the count from graph nodes: rejected because cache cardinality is the frozen source of truth and cache tests must detect divergence between topology and physical parsing.
- Return a full final `Analysis` model now: rejected as issue #13 scope.

### 4. Keep cached summaries immutable and clone at boundaries

Documents and local summaries are private cache inputs. Local summaries, expanded summaries, graphs, and final recursive results continue to be defensively cloned wherever a mutable slice crosses a cache or caller boundary. Applying a child summary reads its counters and appends cloned/grouped evidence to a fresh parent builder; it never mutates the cached child.

Repeated mounts with the same canonical target and equal depth may still be grouped before one application, but the checked multiplicity applies only occurrence fields. Maximum depth is composed once per mount-depth group, and dependency/resource identities are set unions. Diamond paths reuse the same one-occurrence cached summary while separate parents receive equivalent owned contributions.

Alternatives considered:

- Store caller-returned slices directly: rejected because caller mutation could corrupt later cache hits inside the same invocation.
- Cache summaries already multiplied for a parent edge: rejected because cache keys would need parent/multiplicity/depth context and would destroy canonical one-occurrence reuse.

### 5. Make overflow checks transactional and share one error contract

Retain `OverflowError` and diagnostic `SB2004`, but audit every graph and recursive mutation boundary. Checked helpers cover non-negative addition, multiplication, depth composition, edge occurrence compaction, dependency accumulation, parsed-cache cardinality conversion, scene-instance coverage, grouped evidence, and all metric contributions.

Builders calculate contributions into temporaries before assigning fields or appending evidence. On failure, the current parent/graph builder is discarded, cached children remain untouched, and `Analyze` returns a zero `RecursiveResult`. Negative operands use the same typed failure contract because they violate the non-negative model even when machine arithmetic would not overflow.

A small checked-cardinality helper accepts an unsigned/testable input and rejects values above `math.MaxInt64`; production passes the document-map length. This makes the otherwise memory-unreachable boundary unit-testable without enormous allocation.

Alternatives considered:

- Rely on Go integer wrapping and detect negatives afterward: rejected because some overflows remain non-negative and partial state may already be published.
- Saturate or clamp: rejected because it would produce false budget evidence.
- Introduce arbitrary-precision counters: rejected because the frozen output and metric contracts are signed `int64`.

## Risks / Trade-offs

- [Opening now returns a closer, increasing fixture ceremony] → provide compact in-memory test helpers based on `io.NopCloser` and keep ownership in one cache method.
- [A close failure can blur open versus parse classification] → treat close failure after a successful parse as an unavailable load failure, cache it, and never insert the document; preserve a parser error as the primary fatal error when parsing already failed.
- [Centralizing caches can create a large refactor diff] → move storage first without semantic change, then migrate effects/cardinality, with focused tests after each slice.
- [Cached mutable documents could be accidentally changed by future code] → keep documents private to analysis and continue cloning every domain model that crosses cache/caller boundaries.
- [Parsed coverage is unavailable through legacy `Expand`] → retain `Expand` only as the issue #9 summary projection and make later final-analysis composition consume `Analyze`.
- [Some overflow boundaries are impossible to reach through real allocation] → unit-test checked primitives directly and retain end-to-end overflow fixtures for reachable repeated contributions.

## Migration Plan

1. Add the cache owner and move existing invocation maps behind it without changing recursive outputs.
2. Split `SceneLoader` into opener/parser effects and migrate all repository test helpers and real-project fixtures.
3. Add parsed-cache cardinality to `RecursiveResult` and its cloning boundary.
4. Audit arithmetic mutations and add checked cardinality coverage.
5. Commit production-only changes, then add instrumentation/overflow tests in a separate test commit.
6. Run all quality gates, strict OpenSpec validation, sync/archive the change, and merge Draft PR #36 after CI.

Rollback is a normal revert of the focused PR: there is no persisted cache, schema migration, external API, or stored data to unwind.
