## 1. Finalize Root Metrics

- [x] 1.1 Add a pure finalizer that preserves the six recursive values, checked-counts unique resource identities, copies checked graph dependencies, and validates all eight non-negative values; verify unit tests cover valid, cardinality-overflow, and negative-input failure paths.
- [x] 1.2 Call finalization only after root dependency/resource identity merging and before constructing `RecursiveResult`; verify fatal errors return zero results and `Expand` projects the same final values.
- [x] 1.3 Keep local and cached one-occurrence summaries' unique fields zero; verify repeated and diamond applications do not multiply `external_resources` or `scene_dependencies`.
- [x] 1.4 Keep production changes inside `internal/analysis` with no inherited aggregation, reliability, CLI, report, config, budget, preset, dependency, process, Godot, or network changes; verify the production diff boundary.

## 2. Commit the Production Feature

- [x] 2.1 Format production files and run `go build ./...`, `go test ./...`, and `go vet ./...`; verify existing recursive/graph behavior remains green with final unique metrics.
- [x] 2.2 Commit only non-test implementation files as `feat: finalize eight scene metrics`; verify no `*_test.go`, planning progress, or unrelated files enter the commit.

## 3. Cover Frozen Metric Semantics

- [x] 3.1 Add table-driven finalizer tests covering every field, canonical ordered lookup, ownership, validation, and root-only zero unique values; verify all eight exact values and order.
- [x] 3.2 Add literal-type fixtures covering exact mesh/three 3D lights, explicit/default shadows, custom subclasses, `MultiMeshInstance3D`, 2D lights, imported contents, mounts, and inherited roots; verify no false counts.
- [x] 3.3 Add repeated ×100 and diamond fixtures proving occurrence multiplication, maximum depth, canonical resource union, unresolved tuple distinction, cache reuse, and unique dependency count.
- [x] 3.4 Add resolved/transitive inheritance and unresolved/imported fixtures proving graph-wide unique fields while inherited effective-tree and unavailable contents remain unclaimed.
- [x] 3.5 Add the exact §20.7 City/Building/Lamp integration fixture; verify 62 nodes, seven scene instances, one shared texture, and the expected unique dependencies.

## 4. Verify and Deliver Tests

- [x] 4.1 Run focused metric/local/recursive/graph tests plus `go test ./...`, `go test -race ./...`, `go vet ./...`, and CI-pinned golangci-lint v2.12.0; verify all gates pass deterministically.
- [x] 4.2 Commit only `*_test.go` files and OpenSpec task progress as `test: cover eight scene metrics`; verify the feature and test commits remain independently reviewable.
- [x] 4.3 Run strict OpenSpec validation and `git diff --check`, inspect Draft PR #37 through the GitHub connector, and verify the diff stays within #12 planning, analysis finalization, and metric tests before archive/merge.
