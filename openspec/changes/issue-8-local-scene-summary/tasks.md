## 1. Define the Local Summary Domain

- [ ] 1.1 Add `internal/analysis` models for optional depth, ordinary nodes, mount kinds and candidate evidence, inherited roots, override stubs, external-resource records, parent findings, and the complete local summary; verify exported values retain source context without storing or aliasing the input `tscn.Document`.
- [ ] 1.2 Define and document the `metrics.Values` boundary for local occurrence/depth contributions, including zero `ExternalResources` and `SceneDependencies`; verify `go test ./...` and `go vet ./...` remain green with the additive production API.

## 2. Implement Classification and Resource Extraction

- [ ] 2.1 Implement the source-ordered classification precedence for inherited roots, placeholder mounts, instance mounts, ordinary typed nodes, and override stubs; verify the production build excludes inherited roots and mounts from `Nodes`, excludes inherited roots from `SceneInstances`, and preserves each classification record exactly once.
- [ ] 2.2 Implement exact `PackedScene` external-candidate matching plus explicit non-candidate reference kinds, and extract all external-resource declarations into an ID-sorted owned slice; verify duplicate raw targets under different IDs remain separate and no filesystem/path resolution is called.
- [ ] 2.3 Implement literal ordinary-node contributions for `MeshInstance3D`, the three supported 3D light types, and explicit `shadow_enabled=true`; verify other types, absent/false shadows, mounts, inherited roots, and stubs do not change those counters.

## 3. Implement Local Paths and Depth

- [ ] 3.1 Implement canonical serialized-parent validation and the first-pass local path index, supporting exact `.`, multi-segment relative parents, forward declarations, and duplicate-path detection; verify the package never cleans `..`, accepts absolute NodePaths, or chooses one ambiguous record by declaration order.
- [ ] 3.2 Implement second-pass optional depth propagation from the depth-1 root anchor, source-aware deterministic findings for invalid/missing/ambiguous parents, and partial-depth state; verify unknown descendants stay unknown while independently known classifications and counters remain intact.
- [ ] 3.3 Compute local `TreeDepth` as the maximum known depth across ordinary nodes and non-inherited mounts and preserve each mount depth for later expansion; verify instance headers never add an ordinary root contribution that later resolved aggregation would double-count.

## 4. Commit the Production Feature

- [ ] 4.1 Format all new production Go files, run `go build ./...`, `go test ./...`, and `go vet ./...`, then commit only non-test implementation files as `feat: add local scene summaries`; verify `git show --stat --oneline HEAD` contains no `*_test.go` files or unrelated changes.

## 5. Add Focused Contract Tests

- [ ] 5.1 Add table-driven tests for ordinary roots, `.`, multi-segment and forward-declared parents, and exact local node/tree-depth values; verify the focused analysis-package test reports depths 1 through 4 and deterministic repeated output.
- [ ] 5.2 Add table-driven tests for missing and ambiguous parents, `..` segments, repeated separators, non-root `.` segments, absolute NodePaths, and unknown-depth descendants; verify findings retain raw parent/source position, partial state is set, and no guessed depth enters the maximum.
- [ ] 5.3 Add tests for external `PackedScene` candidates, missing/wrong-type external IDs, `SubResource`, placeholders, inherited roots, and override stubs; verify mount depths and scene-instance counts are exact and no instance or inherited root is counted as an ordinary node.
- [ ] 5.4 Add tests for literal mesh/light/shadow classifications and ID-sorted external-resource cloning with duplicate raw paths; verify unique graph metrics remain zero and mutation of input maps or references after summarization cannot alter the result.

## 6. Verify and Deliver the Test Commit

- [ ] 6.1 Run the focused analysis tests plus `go test ./...`, `go test -race ./...`, `go vet ./...`, and the CI-pinned `golangci-lint`; verify every gate exits successfully without Godot, network, filesystem fixtures, Node.js, or OpenSpec in the shipped runtime graph.
- [ ] 6.2 Commit only `*_test.go` files and OpenSpec task-progress updates as `test: cover local scene summaries`, keeping the production implementation in the preceding feature commit; verify the two commits are separately reviewable and the test commit contains no production Go changes.
- [ ] 6.3 Run `openspec validate issue-8-local-scene-summary --strict` and `git diff --check`, inspect the Draft PR #33 changed-file list, and verify the final change remains limited to issue #8 planning, production, and tests with resolved-child aggregation still absent.
