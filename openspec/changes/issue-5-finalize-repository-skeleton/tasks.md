## 1. Metric Domain Contracts

- [ ] 1.1 Add table-driven metric tests for all eight IDs, labels, canonical order, validity, defensive copies, value access, and negative-value errors; verify the targeted metric test command demonstrates every invariant.
- [ ] 1.2 Replace duplicated metric catalog state with one ordered definition source and add typed value validation while preserving existing IDs and JSON fields; verify the targeted metric tests and `go test ./internal/budget ./internal/preset` pass.

## 2. Diagnostic Domain Contracts

- [ ] 2.1 Add table-driven diagnostic tests for the complete `SB1001`–`SB2004` catalog, severity validity, defensive order, record consistency, invalid fields, and wrapped coded-error discovery; verify the targeted diagnostic test command covers each catalog entry.
- [ ] 2.2 Implement typed diagnostic codes, ordered definitions, record validation, and the narrow code-bearing error protocol; verify all diagnostic tests pass and unknown codes/severities are rejected.

## 3. Parser and CLI Integration

- [ ] 3.1 Extend parser and CLI tests to cover shared `SB2001` typing, wrapped-error lookup, preserved parse location/text, coded stderr rendering, root help, version, and placeholder failures; verify the targeted `tscn` and `cli` test commands exercise these cases.
- [ ] 3.2 Migrate `tscn.ParseError` to the shared diagnostic code type and centralize typed fatal rendering in CLI while retaining exit code `2` and the untyped fallback; verify `go test ./internal/tscn ./internal/cli` passes with no stack trace or duplicate code prefix.

## 4. Repository Skeleton Gates

- [ ] 4.1 Add an explicit `go build ./...` step to the cross-platform CI job while retaining test, vet, Linux race, and lint coverage; verify the workflow contains the intended gates for Linux, macOS, and Windows.
- [ ] 4.2 Audit `cmd/deadweight.gdt` and current domain-package imports/calls for the documented layering, scene-filesystem, console, process-exit, and runtime-dependency boundaries; record or fix any in-scope violation and verify no empty future architecture package is added.

## 5. Verification and Delivery

- [ ] 5.1 Run `gofmt` on changed Go files and verify `go build ./...`, `go test ./...`, `go test -race ./...`, and `go vet ./...` all pass.
- [ ] 5.2 Run the repository lint gate and `openspec validate issue-5-finalize-repository-skeleton --strict`; verify both complete successfully and all implemented task checkboxes accurately reflect the change.
- [ ] 5.3 Keep production/CI changes and test changes in separate focused commits as requested, push both commits to Draft PR #30, and verify the PR diff remains limited to issue #5 plus its OpenSpec artifacts.
