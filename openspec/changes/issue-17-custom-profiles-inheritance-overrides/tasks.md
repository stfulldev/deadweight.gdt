## 1. Establish the Effective Policy Domain

- [x] 1.1 Add owned selector, metadata, and effective-policy value types in `internal/policy`; verify kind validation, zero-value no-base semantics, and defensive budget cloning with focused tests.
- [x] 1.2 Add canonical optional-budget merge/set helpers that preserve absence and explicit zero across all eight metrics; verify every metric can be independently inherited and overridden.
- [x] 1.3 Add one deterministic `SB2003` policy-error boundary using `config.Error`; verify source, field, detail, diagnostic code, and zero-result behavior for semantic failures.

## 2. Validate and Resolve Profile Graphs

- [x] 2.1 Load the owned built-in catalog, index built-in/custom IDs, and reject every custom collision or missing parent in sorted profile order; verify unselected invalid profiles still fail.
- [x] 2.2 Implement bounded tri-color single-inheritance traversal with stack evidence and memoization; verify full closed cycle chains, deterministic anchors, exactly 32 custom levels accepted, and 33 rejected.
- [x] 2.3 Resolve root custom defaults plus inherited/overridden metadata and budgets field by field; verify profile-to-built-in and profile-to-profile chains preserve omitted values.

## 3. Apply Selection and Override Precedence

- [x] 3.1 Implement CLI-over-config selector selection with mutually exclusive preset/profile domains and no-base support; verify cross-domain and unknown references fail without fallback.
- [x] 3.2 Parse ordered CLI `metric=limit` overrides strictly and apply duplicates last-wins; verify all eight IDs, zero, maximum int64, malformed separators, whitespace, unknown metrics, negatives, non-integers, and overflow.
- [x] 3.3 Overlay selected base, descendant profiles, top-level budgets, and CLI budgets in frozen order; verify the documented four-layer example and every layer transition.
- [x] 3.4 Reject policies with no effective budget while allowing no-base top-level-only and CLI-only policies; verify the failure suggests selecting a base or supplying a budget.

## 4. Commit the Production Feature

- [x] 4.1 Format production files and run focused policy tests plus `go build ./...`, `go test ./...`, and `go vet ./...`; verify no config, preset, budget, parser, analyzer, or CLI regression.
- [x] 4.2 Commit only non-test `internal/policy` production files as `feat: add custom profile resolution`; verify no `*_test.go`, OpenSpec progress, or unrelated files enter the commit.

## 5. Cover the Frozen Matrix

- [x] 5.1 Add table-driven selector and graph tests for CLI/config precedence, no base, unknown/cross-domain IDs, collisions, missing parents, unused invalid profiles, cycles, and depth boundaries; verify deterministic fields and messages.
- [x] 5.2 Add metadata inheritance tests for built-in parents, custom parents, root defaults, each optional field, lifecycle retention, and caller ownership; verify repeated resolution remains unchanged after mutation.
- [x] 5.3 Add all-eight budget tests for ancestor/descendant/project/CLI layers, absence, zero, duplicates, top-level-only, CLI-only, and empty policies; verify the four-layer result exactly.
- [x] 5.4 Add malformed CLI override and error-contract tests; verify `SB2003`, source, indexed field, exact offending value, signed-64-bit boundaries, and zero effective results.

## 6. Verify and Deliver Tests

- [x] 6.1 Run focused policy tests, `go test ./...`, `go test -race ./...`, `go vet ./...`, CI-pinned golangci-lint, strict OpenSpec validation, and `git diff --check`; verify deterministic green gates.
- [x] 6.2 Mark completed OpenSpec tasks and commit only `*_test.go` plus task progress as `test: cover custom profile resolution`; verify feature and test commits remain independently reviewable.
- [x] 6.3 Inspect Draft PR #41 through the GitHub connector and verify the final diff stays within #17 planning and `internal/policy` before sync, archive, and merge.
