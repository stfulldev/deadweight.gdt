## 1. Define Recursive Analysis Contracts

- [x] 1.1 Add `internal/analysis` models for expanded summaries, resolved/unresolved coverage, deterministic target classifications, unresolved instance evidence, resource identities, inherited-target evidence, and scene-scoped parent findings; verify every model preserves canonical/display/source identities and uses owned value data.
- [x] 1.2 Add the narrow resource-resolver interface, parsed-scene loader function, analyzer construction validation, and canonical root input contract; verify `project.Resolver` satisfies the interface without importing project discovery, CLI, report, or filesystem-open policy into analysis.
- [x] 1.3 Add checked non-negative `int64` addition/multiplication/depth helpers and a typed overflow error exposing `SB2004`; verify the production package compiles and all existing domain tests remain green.

## 2. Classify Targets and Closure Resource Identities

- [x] 2.1 Implement deterministic lookup of each mount's external-resource declaration and classify placeholders, `SubResource`, missing IDs, path-resolution failures, imported/binary extensions, unsupported extensions, and exact `.tscn` candidates; verify every mount produces either one loadable target or one complete unresolved record.
- [x] 2.2 Attempt resolved exact `.tscn` targets regardless of declared resource type, preserve typed parse failures as fatal, and convert non-parse loader failures to unavailable unresolved evidence; verify classification never uses an extension suffix to bypass secure canonical resolution.
- [x] 2.3 Resolve every successfully parsed scene's external-resource declarations from its own canonical declaring path and build resolved-canonical or unresolved-tuple identities; verify deterministic unique ordering without multiplying repeated child declarations.
- [x] 2.4 Detect a successfully parsed child with an inherited root and retain inherited-target evidence plus one known mounted root without expanding its base; verify no exact inherited contribution or scene-instance double count is claimed before issue #14.

## 3. Implement Invocation-scoped Recursive Expansion

- [x] 3.1 Add fresh per-call document, expanded-summary, and in-progress maps keyed by canonical scene path; verify the root and every first-seen child load/build once while separate analyzer calls receive fresh state.
- [x] 3.2 Recursively expand an acyclic canonical child chain from local mounts, cache only completed one-occurrence summaries, and return a typed recursive-reference safety error without a summary for an in-progress identity; verify traversal terminates without implementing issue #10 graph rendering or `SB2002` chains.
- [x] 3.3 Clone cached summaries and their nested slices/pointers at application and return boundaries; verify caller mutation cannot alter a later repeated or diamond cache hit.

## 4. Aggregate Occurrences, Depth, Coverage, and Unique Sets

- [x] 4.1 Start from local six-field metrics and compact resolved applications by canonical child plus optional mount depth, accumulating group multiplicity only through checked arithmetic; verify local mount scene-instance counts remain present exactly once per occurrence.
- [x] 4.2 Apply `N` resolved child summaries with checked multiplication/addition for node/type metrics, nested scene instances, resolved/unresolved coverage, and evidence occurrences; verify child roots are not added separately and repeated nested instances keep full multiplicity.
- [x] 4.3 Add exactly one node and one unresolved coverage occurrence for every unresolved non-inherited mount while retaining the local scene-instance count; verify no unknown child metrics, resources, or dependencies are inferred.
- [x] 4.4 Compose resolved depth as `mountDepth + child.TreeDepth - 1`, retain known unresolved mount maxima, propagate unknown-depth evidence, and never multiply tree depth; verify all depth arithmetic is checked.
- [x] 4.5 Union each resolved child canonical path, transitive dependency path, and external-resource identity, then materialize deterministic owned slices; verify root exclusion and chain/diamond/repeated unique semantics.

## 5. Commit the Production Feature

- [x] 5.1 Format new production Go files, run `go build ./...`, `go test ./...`, and `go vet ./...`, then commit only non-test implementation files as `feat: add recursive scene expansion`; verify `git show --name-only HEAD` contains no `*_test.go`, CLI, report, config, or unrelated files.

## 6. Add Recursive Expansion Tests

- [x] 6.1 Add in-memory resolver/loader harnesses and exact chain tests proving canonical declaring-scene bases, child-root replacement, nested scene-instance counts, and mount-depth composition; verify the focused analysis test passes without process output or Godot.
- [x] 6.2 Add repeated ×100 and diamond tests with loader/build counters, occurrence metrics, coverage multiplicity, dependency/resource unions, invocation reset, and cache-copy isolation; verify each canonical child is loaded/expanded once per call but applied on every occurrence.
- [x] 6.3 Add a table-driven unresolved-target matrix for missing/wrong IDs, missing/outside/UID/user paths, `SubResource`, placeholder, imported/binary targets, unsupported extensions, loader failures, inherited targets, and wrong-type `.tscn`; verify every record retains source context and exactly one known root.
- [x] 6.4 Add tests for unknown mount depth, malformed nested `.tscn`, recursive-reference termination, deterministic ordering, checked helper boundaries, metric/coverage/depth overflow, and typed `SB2004`; verify fatal cases return no usable expanded summary and never panic or wrap.
- [x] 6.5 Add a small temporary-project integration test using the real `project.Resolver` and parsed format-3 scenes; verify relative child/resource paths stay within the project and the analysis package itself never opens files or discovers the project.

## 7. Verify and Deliver the Test Commit

- [x] 7.1 Run focused analysis tests plus `go test ./...`, `go test -race ./...`, `go vet ./...`, and CI-pinned golangci-lint v2.12.0; verify every gate passes with deterministic output and no runtime dependency on Godot, network, Node.js, OpenSpec, or persistent caches.
- [x] 7.2 Commit only `*_test.go` files and OpenSpec task-progress updates as `test: cover recursive scene expansion`, preserving the preceding production feature commit; verify the commits are independently reviewable and the test commit contains no production Go changes.
- [x] 7.3 Run `openspec validate issue-9-recursive-nested-scene-analysis --strict` and `git diff --check`, inspect Draft PR #34 through the GitHub connector, and verify its diff remains limited to issue #9 planning, recursive analysis production, and tests with graph/CLI/report/inheritance expansion absent.
