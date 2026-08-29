## 1. Establish the Configuration Domain

- [x] 1.1 Add owned version-one config/profile models with optional selectors, metadata, and `budget.Limits`; verify constructed values distinguish omission from explicit zero.
- [x] 1.2 Add stable config error reasons with source/field evidence and `SB2003`; verify wrapped filesystem and JSON causes remain discoverable while user messages stay deterministic.
- [x] 1.3 Add shared ID, renderer, and quality constants/validators; verify every frozen accepted value and boundary-length ID is represented without importing CLI concerns.

## 2. Discover and Decode Configuration

- [x] 2.1 Implement explicit-first and implicit-default discovery with an absent result for a missing implicit file; verify missing explicit, directory, unreadable, and filesystem failures are typed.
- [x] 2.2 Implement exactly-one-document decoding with `DisallowUnknownFields` at top-level, profile, and budget boundaries; verify malformed, trailing, unknown, wrong-type, and explicit-null fields return zero config plus dotted paths.
- [x] 2.3 Convert all eight raw budget properties into owned optional `int64` limits; verify absence, zero, maximum signed value, negative, fractional, string, boolean, null, and overflow behavior.
- [x] 2.4 Implement separately callable static validation for version, selector exclusion, IDs, platform, target FPS, renderer, quality, and budget domains; verify pattern-valid unknown selectors/parents remain available to issue #17.

## 3. Deliver the Canonical Schema

- [x] 3.1 Add `schema/deadweight.gdt.schema.json` from the frozen Draft 2020-12 contract; verify it parses as JSON and every object rejects additional properties.
- [x] 3.2 Add schema parity checks for required version, selectors, eight metrics, ID pattern, profile map, metadata types, and enums; verify dynamic graph/merge rules are absent.

## 4. Commit the Production Feature

- [x] 4.1 Format production Go files and run focused config tests plus `go build ./...`, `go test ./...`, and `go vet ./...`; verify no existing project, preset, budget, or CLI behavior regresses.
- [x] 4.2 Commit only `internal/config` production files and the canonical schema as `feat: add strict configuration v1`; verify no `*_test.go`, OpenSpec progress, or unrelated files enter the commit.

## 5. Cover the Frozen Matrix

- [x] 5.1 Add discovery fixtures for explicit priority, implicit present/absent, missing explicit, directories, unreadable files, and source ownership; verify behavior on platform-neutral temporary paths.
- [x] 5.2 Add minimal/full document and all eight optional-limit fixtures; verify selectors, profiles, metadata presence, `fail_on_partial`, empty objects, and zero survive exactly.
- [x] 5.3 Add table-driven rejection fixtures for unknown keys/metrics, nulls, negative/float/string/boolean/overflow limits, unsupported versions, trailing JSON, selector conflict, invalid IDs/platform/enums/FPS, and malformed input; verify every error exposes `SB2003`, source, field, and zero result.
- [x] 5.4 Add ownership and phase-boundary fixtures; verify caller mutation cannot alias input storage and valid unknown preset/profile/extends IDs are not dynamically rejected.

## 6. Verify and Deliver Tests

- [x] 6.1 Run focused config/schema tests, `go test ./...`, `go test -race ./...`, `go vet ./...`, CI-pinned golangci-lint v2.12.0, strict OpenSpec validation, and `git diff --check`; verify deterministic green gates.
- [x] 6.2 Commit only `*_test.go` files and OpenSpec task progress as `test: cover strict configuration v1`; verify feature and test commits remain independently reviewable.
- [x] 6.3 Inspect Draft PR #40 through the GitHub connector and verify the final diff stays within #16 planning, `internal/config`, canonical schema, and tests before sync/archive/merge.
