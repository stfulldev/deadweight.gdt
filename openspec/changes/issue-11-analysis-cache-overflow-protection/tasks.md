## 1. Define Injectable Cache Effects

- [x] 1.1 Replace the combined scene-loader boundary with required `SceneOpener` and `SceneParser` effects that retain canonical/display source identities; verify constructor validation rejects every nil dependency and the package compiles with no production filesystem or parser hidden inside analysis.
- [x] 1.2 Add an invocation-scoped cache owner for documents, document failures, local summaries, resource resolutions, and one-occurrence expanded summaries keyed only by clean canonical `.tscn` paths; verify a fresh cache is allocated for every `Analyze` call.
- [x] 1.3 Move existing memoized state access behind cache methods while leaving DFS colors/stacks and expansion in-progress guards on invocation traversal state; verify focused graph and recursive tests retain their pre-#11 outputs.
- [x] 1.4 Make the document cache own one open, one parse attempt, and one close per canonical miss, caching either the successful document or stable failure; verify `SB2001` remains fatal while non-parse load failures preserve unavailable-scene behavior.

## 2. Publish Cache-Cardinality Coverage

- [x] 2.1 Add checked `ParsedSceneFiles` coverage to `RecursiveResult` and its defensive clone boundary without adding global cardinality to one-occurrence `ExpandedSummary`; verify zero results remain safe and `Expand` remains a summary-only compatibility projection.
- [x] 2.2 Add a testable checked-cardinality conversion that rejects values above signed `int64` with `SB2004`; verify zero, maximum, overflow, and invalid arithmetic boundaries do not wrap or clamp.
- [x] 2.3 Populate parsed-file coverage only after graph validation and occurrence expansion succeed, using successful document-cache cardinality; verify root-only, transitive, repeated, diamond, inheritance, unavailable, and fatal-parse semantics match the delta spec.

## 3. Harden Summary Ownership and Arithmetic

- [x] 3.1 Audit cached local/expanded summaries and caller returns so every mutable slice crosses cache boundaries through an owned clone; verify caller mutation cannot affect another branch, cache hit, or later invocation.
- [x] 3.2 Audit repeated and diamond summary application so only occurrence counters/evidence multiply, unique identity sets union, and composed tree depth remains a maximum; verify no cache field is mutated during parent application.
- [x] 3.3 Audit graph edge compaction, dependency accumulation, occurrence grouping, metric contributions, scene-instance coverage, and evidence occurrence arithmetic through the shared checked helpers; verify failures return typed `SB2004` before partial graph/result publication.
- [x] 3.4 Keep opener/parser/cache mechanics single-threaded and memory-only with no global state, disk persistence, invalidation, process, Godot, network, CLI, report, config, budget, final-metric, reliability, or inherited-aggregation changes; verify the production diff remains inside `internal/analysis`.

## 4. Commit the Production Feature

- [x] 4.1 Format production Go files, migrate every repository production call site to the opener/parser constructor, and run `go build ./...`, `go test ./...`, and `go vet ./...`; verify existing behavior remains green before committing.
- [x] 4.2 Commit only non-test implementation files as `feat: harden recursive analysis caches`; verify the commit contains no `*_test.go`, OpenSpec planning progress, CLI/report/config/budget changes, or unrelated files.

## 5. Add Cache Instrumentation Tests

- [x] 5.1 Replace the memory loader fixture with independent counting opener/parser hooks and local-summary instrumentation; verify root-only and chain tests assert exactly one open, parse, and summary build per canonical scene.
- [x] 5.2 Add exact diamond and repeated ×100 tests proving one open/parse/local-summary/expanded-summary result per unique canonical scene, per-occurrence metrics/evidence, unique resources/dependencies, unmultiplied depth, and parsed coverage from cache cardinality.
- [x] 5.3 Add graph-then-expansion and two-invocation tests proving both phases share one cache while separate calls reopen/reparse independently; verify no memoized state persists on the analyzer or disk.
- [x] 5.4 Add open, parse, and close failure tests proving effects are attempted once, successful cache cardinality excludes failed targets, unavailable failures retain evidence, fatal `SB2001` returns a zero result, and cached failure identity remains stable.
- [x] 5.5 Add cache ownership tests that mutate returned graph, summary, dependency, resource, and evidence slices; verify cached one-occurrence values and repeated equivalent analyses remain deterministic.

## 6. Add Overflow and Integration Tests

- [x] 6.1 Add table-driven checked add, multiply, depth, and cardinality boundary tests for zero, signed-`int64` maximum, overflow, and negative operands; verify every failure exposes `SB2004` and zero usable values.
- [x] 6.2 Add end-to-end overflow fixtures for edge compaction, dependency/cardinality helpers, repeated metrics, scene coverage, unresolved/inherited evidence, and parent findings; verify no panic, wrap, clamp, cached mutation, or partial graph/result occurs.
- [x] 6.3 Migrate the temporary-project real resolver/parser integration test to an actual file opener plus injected `tscn.Parse`; verify secure canonical I/O, one physical parse per file, parsed coverage, and the no-Godot/no-process/no-network boundary.

## 7. Verify and Deliver the Test Commit

- [x] 7.1 Run focused cache/graph/recursive tests plus `go test ./...`, `go test -race ./...`, `go vet ./...`, and CI-pinned golangci-lint v2.12.0; verify all deterministic gates pass on Linux-compatible no-Godot fixtures.
- [x] 7.2 Commit only `*_test.go` files and OpenSpec task-progress updates as `test: cover analysis cache integrity`, preserving the preceding production feature commit; verify both commits remain independently reviewable.
- [ ] 7.3 Run `openspec validate issue-11-analysis-cache-overflow-protection --strict` and `git diff --check`, inspect Draft PR #36 through the GitHub connector, and verify its diff stays limited to #11 planning, `internal/analysis` cache/effect/result production, and tests with later MVP layers absent.
