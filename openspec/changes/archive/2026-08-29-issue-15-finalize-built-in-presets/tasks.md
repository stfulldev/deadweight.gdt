## 1. Freeze the Built-in Catalog Contract

- [x] 1.1 Audit all three embedded JSON records against MVP section 22.2, correct only genuine discrepancies, and add one explicit expected-catalog test covering product order, every metadata field, lifecycle label, and all eight limits; verify with `go test ./internal/preset -run TestBuiltinsAreFrozenAndOrdered`.
- [x] 1.2 Confirm the production catalog is loaded solely from version-controlled `go:embed` data and add a focused embedded-loading assertion if coverage is missing; verify the preset package test passes without reading repository-relative runtime files.

## 2. Make Embedded-data Validation Strict and Testable

- [x] 2.1 Refactor cached loading into deterministic decode/validation helpers that reject unknown fields, trailing JSON, missing metadata, mismatched or duplicate IDs, non-positive target FPS, and invalid lifecycle labels with preset-and-field context; verify table-driven negative cases with `go test ./internal/preset -run 'Test.*Invalid|Test.*Reject'`.
- [x] 2.2 Enforce the MVP renderer and quality allowlists and add cases for every allowed value plus representative unsupported values; verify the focused validation tests accept allowed IDs and return field-specific errors for invalid IDs.
- [x] 2.3 Enforce that all eight metric limits are present non-negative integers and add table-driven missing/negative cases across the metric set; verify zero is accepted and each invalid metric is named by the focused preset validation tests.

## 3. Isolate Returned Values and Centralize Lookup Errors

- [x] 3.1 Introduce one preset deep-copy path and use it for both `Builtins` and successful catalog lookup results; verify tests can mutate returned metadata and budget pointers without changing package state or the source catalog.
- [x] 3.2 Change preset lookup to return an actionable domain error, derive available IDs from catalog order, and update all CLI callers; verify known lookup succeeds and unknown lookup/CLI tests include the requested ID plus `mobile, steam-deck, desktop` with exit code `2`.

## 4. Preserve Experimental Positioning and Release Governance

- [x] 4.1 Audit README, preset list/show output, tests, and `CHANGELOG.md` so built-ins remain labeled heuristic and experimental, carry a performance-guarantee disclaimer, make no certification or Valve-endorsement claim, and document any frozen-data correction made by this change; verify the relevant CLI tests and a review of the resulting text and diff.

## 5. Run Final Quality Gates

- [x] 5.1 Format changed Go files and run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `golangci-lint run`; verify every command exits successfully.
- [x] 5.2 Run `openspec validate issue-15-finalize-built-in-presets --strict` and `git diff --check`, then inspect the final diff to confirm it stays within issue #15 and introduces no runtime dependency on OpenSpec, Node.js, Godot, or the network.
