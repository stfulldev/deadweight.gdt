## 1. Establish the Evaluation Domain

- [x] 1.1 Add frozen `PASSED`, `FAILED`, and `INCOMPLETE` status values plus validation; verify the zero/unknown values are rejected and every catalog value is accepted.
- [x] 1.2 Add owned evaluation output containing reliability, effective partial policy, exceeded count, and ordered comparison results; verify caller mutation cannot affect later evaluation.
- [x] 1.3 Add typed canonical-order validation for configured limits; verify all eight absent/non-negative fields pass and the first negative field is identified deterministically.

## 2. Resolve Partial Policy

- [x] 2.1 Add zero-value inherit plus explicit fail/allow domain overrides; verify override validation does not import or expose Cobra flag types.
- [x] 2.2 Resolve config `fail_on_partial` with inherit/fail/allow precedence; verify default false, config true, forced fail, forced allow, and unknown override behavior.

## 3. Evaluate Budgets and Verdicts

- [x] 3.1 Validate metrics, limits, and reliability before comparison and return zero evaluation on failure; verify negative actuals/limits and unknown reliability never publish partial results.
- [x] 3.2 Reuse canonical inclusive comparisons to count exceedances and retain exact/lower-bound/approximate evidence; verify absence, zero, equality, limit+1, all-eight order, and empty-limit behavior.
- [x] 3.3 Apply verdict priority `INCOMPLETE > FAILED > PASSED` for non-fatal outcomes; verify partial rejection retains every comparison and exceeded count while partial allowed uses the budget verdict.

## 4. Commit the Production Feature

- [x] 4.1 Format production files and run focused budget tests plus `go build ./...`, `go test ./...`, and `go vet ./...`; verify existing checker, policy, analyzer, and CLI behavior remains compatible.
- [x] 4.2 Commit only non-test `internal/budget` production files as `feat: add budget evaluation and partial policy`; verify no `*_test.go`, OpenSpec progress, or unrelated files enter the commit.

## 5. Cover the Frozen Matrix

- [x] 5.1 Add table-driven comparison tests for all eight metrics, absent limits, zero, boundary equality, limit+1, multiple failures, canonical order, and result ownership.
- [x] 5.2 Add exact/lower-bound/approximate verdict tests with partial allowed/rejected and simultaneous exceedance; verify reliability evidence and final priority remain deterministic.
- [x] 5.3 Add partial override tests for every config/override pair and invalid values; verify default false and fail/allow precedence exactly.
- [x] 5.4 Add invalid-input tests for every negative metric/limit and unknown reliability; verify typed evidence and zero `Evaluation` on every error.

## 6. Verify and Deliver Tests

- [x] 6.1 Run focused budget tests, `go test ./...`, `go test -race ./...`, `go vet ./...`, CI-pinned golangci-lint, strict OpenSpec validation, and `git diff --check`; verify deterministic green gates.
- [x] 6.2 Mark completed OpenSpec tasks and commit only `*_test.go` plus task progress as `test: cover budget evaluation and partial policy`; verify feature and test commits remain independently reviewable.
- [x] 6.3 Inspect Draft PR #42 through the GitHub connector and verify the final diff stays within #18 planning and `internal/budget` before sync, archive, and merge.
