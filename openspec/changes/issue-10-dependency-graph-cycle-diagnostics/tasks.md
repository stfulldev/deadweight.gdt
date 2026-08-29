## 1. Define Graph and Cycle Contracts

- [x] 1.1 Add owned `internal/analysis` models for graph nodes, instance/inheritance edges, dependency count, and the paired recursive result; verify the package compiles and zero values remain safe.
- [x] 1.2 Add a typed cycle error with owned canonical/display chains, deterministic code-free message text, and diagnostic code `SB2002`; verify it satisfies the existing coded/message error contracts without importing CLI behavior into analysis.
- [x] 1.3 Add a graph builder with comparable node/edge keys, checked occurrence compaction, checked dependency accumulation, deterministic sorting, and defensive cloning; verify direct builder tests can exercise overflow and copy isolation.

## 2. Share Invocation Resolution and Classification

- [x] 2.1 Add canonical-scene keyed external-resource resolution caching to the invocation state and refactor recursive expansion to consume it; verify each declaration is securely resolved once per analysis call even though graph and metrics both use it.
- [x] 2.2 Extract one target-classification boundary for local mounts and inherited-root references, preserving raw/resource/reason/display identities; verify exact lowercase `.tscn` still loads regardless of declared resource type and every unsupported form remains explicit.
- [x] 2.3 Preserve successful documents/local summaries and non-parse loader failures across both phases while keeping typed parse failures fatal; verify graph discovery and expansion never duplicate a loader or local-summary build.

## 3. Discover and Validate the Dependency Graph

- [x] 3.1 Implement deterministic outgoing candidate collection for every parsed scene, including local instance mounts and a distinct inherited-root edge; verify candidates sort independently of source/resource map order.
- [x] 3.2 Implement DFS discovery with explicit unvisited/visiting/visited state, ordered stack, and stack indices; verify acyclic chain and diamond graphs visit each canonical node once.
- [x] 3.3 Convert non-parse child loader failures to unavailable unresolved edges and retain all other unresolved target kinds without nodes; verify missing/imported/binary/SubResource/placeholder/unsupported targets never enter the resolved dependency count.
- [x] 3.4 Follow resolved inheritance bases and their transitive edges for topology/resources/cycles only; verify inherited base metrics remain absent from exact occurrence aggregation.
- [x] 3.5 Reconstruct self-cycle and multi-node canonical/display suffix chains on an edge to visiting state and return `SB2002` with zero graph/result; verify deterministic first-cycle selection for equivalent inputs.
- [x] 3.6 Finalize sorted owned nodes/edges and count every non-root graph node once through checked arithmetic; verify chain, repeated, diamond, inheritance, and root-only dependency values.

## 4. Integrate Graph and Recursive Expansion

- [x] 4.1 Add `RecursiveAnalyzer.Analyze` as the two-phase public operation and make `Expand` a compatibility projection; verify both validate roots identically and fatal errors return zero usable values.
- [x] 4.2 Run graph discovery/validation before occurrence expansion and retain only an internal invariant guard in expansion; verify cyclic inputs cannot publish truncated summaries and acyclic issue #9 metrics remain unchanged.
- [x] 4.3 Replace the root summary dependency slice with graph node identities excluding root and union graph-wide parsed resource identities; verify inherited topology contributes unique evidence without setting final unique metric fields before issue #12.
- [x] 4.4 Clone graph and summary slices at cache/application/return boundaries; verify caller mutation cannot affect a later invocation or a repeated/diamond cache application.

## 5. Commit the Production Feature

- [x] 5.1 Format production Go files, run `go build ./...`, `go test ./...`, and `go vet ./...`, then commit only non-test implementation files as `feat: add scene dependency graph`; verify the commit contains no `*_test.go`, CLI/report/config, or unrelated files.

## 6. Add Graph and Cycle Tests

- [x] 6.1 Add exact root-only, chain, diamond, and repeated ×100 graph tests covering node/edge ordering, compaction, dependency count, unchanged occurrence metrics, and one load/build/resolve per canonical scene; verify focused analysis tests pass.
- [x] 6.2 Add table-driven unresolved instance and inheritance edge tests covering missing IDs/paths, imported/binary, unsupported, placeholder, `SubResource`, and loader-unavailable cases; verify every unresolved edge has no target node/count contribution and retains stable evidence.
- [x] 6.3 Add resolved inheritance and transitive-inheritance tests proving instance versus inheritance kinds, graph-wide resources/dependencies, cache reuse, and deliberate inherited metric deferral; verify issue #14 behavior is not implemented early.
- [x] 6.4 Add self-cycle, multi-node mixed instance/inheritance cycle, diamond non-cycle, and multiple-cycle ordering tests; verify exact canonical/display chains, `SB2002`, zero results, termination, and compatibility with fatal exit-code-2 rendering.
- [x] 6.5 Add checked edge/dependency overflow, deterministic result, and caller-mutation tests; verify failures expose `SB2004` without wrap/panic and owned results remain stable.
- [x] 6.6 Add a temporary-project integration test using real `project.Resolver` and parsed format-3 scenes; verify relative instance/inheritance targets remain contained and no Godot/process/network dependency is introduced.

## 7. Verify and Deliver the Test Commit

- [x] 7.1 Run focused graph tests plus `go test ./...`, `go test -race ./...`, `go vet ./...`, and CI-pinned golangci-lint v2.12.0; verify every gate passes on deterministic no-Godot inputs.
- [x] 7.2 Commit only `*_test.go` files and OpenSpec task-progress updates as `test: cover scene dependency graph`, preserving the preceding production feature commit; verify the commits remain independently reviewable.
- [ ] 7.3 Run `openspec validate issue-10-dependency-graph-cycle-diagnostics --strict` and `git diff --check`, inspect Draft PR #35 through the GitHub connector, and verify its diff stays limited to issue #10 planning, graph/recursive analysis production, and tests with final metrics/CLI/report/budget/inherited aggregation absent.
