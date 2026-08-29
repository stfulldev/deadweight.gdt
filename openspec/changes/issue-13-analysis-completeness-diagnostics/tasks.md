## 1. Model and Finalize Completeness

- [ ] 1.1 Add stable analysis status, reliability, and root coverage domain values with validation; verify unit tests cover every valid value, invalid values, and negative coverage.
- [ ] 1.2 Preserve typed resolution reasons on unresolved resource identities without changing canonical or document-local uniqueness; verify resolved deduplication and unresolved tuple distinction remain exact.
- [ ] 1.3 Add warning code `SB1008` for unsupported parent semantics while preserving deterministic diagnostic catalog order and severity validation; verify diagnostic package tests cover the new code.
- [ ] 1.4 Add a pure transactional completeness finalizer that derives status/reliability precedence, checked coverage, resource completeness, grouped warning diagnostics, deterministic sorting, and owned results; verify invalid evidence and grouping overflow return zero values plus errors.
- [ ] 1.5 Call completeness finalization only after graph, metrics, and parsed-file coverage succeed, then publish it through `RecursiveResult`; verify `Analyze` returns a zero result on finalization failure and cached summaries retain no root-only state.
- [ ] 1.6 Keep production changes within `internal/analysis` and `internal/diagnostic`, with no inherited metric aggregation, CLI, report, config, budget, preset, dependency, parser, resolver, process, Godot, or network changes; verify the production diff boundary.

## 2. Commit the Production Feature

- [ ] 2.1 Format production files and run focused analysis/diagnostic tests plus `go build ./...`, `go test ./...`, and `go vet ./...`; verify existing graph, metrics, and fatal behavior remain green.
- [ ] 2.2 Commit only non-test implementation files as `feat: finalize analysis completeness`; verify no `*_test.go`, OpenSpec progress, or unrelated files enter the commit.

## 3. Cover Honesty and Grouping Semantics

- [ ] 3.1 Add complete/exact and partial/lower-bound fixtures for resolved closures, missing/imported/unsupported targets, all path-resolution reasons, placeholders, `SubResource` sources, unavailable scenes, and unsupported parents; verify no branch silently remains complete.
- [ ] 3.2 Add normal resolved `.tres`, texture, material, script, and audio declarations plus unresolved ordinary resource fixtures; verify existing non-scene resources remain complete while missing declarations become partial.
- [ ] 3.3 Add inherited-only and mixed-cause fixtures; verify inherited occurrence coverage is checked and `approximate` wins over `lower_bound` without implementing inherited metrics.
- [ ] 3.4 Add repeated ×100 and diamond fixtures for resolved/unresolved/parsed/inherited coverage, semantic grouping, diagnostic occurrence counts, deterministic order, deduplication, and result ownership.
- [ ] 3.5 Add finalizer validation, negative evidence, grouping overflow, malformed nested text scene, and cycle fixtures; verify every fatal path returns an exact zero `RecursiveResult`.

## 4. Verify and Deliver Tests

- [ ] 4.1 Run focused completeness/recursive/graph/diagnostic tests plus `go test ./...`, `go test -race ./...`, `go vet ./...`, and CI-pinned golangci-lint v2.12.0; verify all gates pass deterministically.
- [ ] 4.2 Commit only `*_test.go` files and OpenSpec task progress as `test: cover analysis completeness`; verify feature and test commits remain independently reviewable.
- [ ] 4.3 Run strict OpenSpec validation and `git diff --check`, inspect Draft PR #38 through the GitHub connector, and verify the diff stays within #13 planning, analysis/diagnostic finalization, and tests before archive/merge.
