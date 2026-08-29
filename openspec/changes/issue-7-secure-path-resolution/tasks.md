## 1. Resolver Domain Contract

- [x] 1.1 Add table-driven tests for `ResolvedPath`, every stable `ResolutionReason`, resolved/unresolved status, typed fatal errors, wrapped causes, and original/candidate retention; verify targeted project-package tests branch with values plus `errors.As`/`errors.Is` rather than message parsing.
- [x] 1.2 Implement the resolved-path, resolution, reason, and fatal-error models without changing discovery error contracts; verify the domain-contract tests compile and pass with no CLI, parser, graph, report, or diagnostic imports.

## 2. Canonical Root and Display Paths

- [x] 2.1 Add tests for resolver construction from valid, relative, missing, non-directory, inaccessible, and symlinked project roots plus display conversion for root, nested, non-clean, relative, and outside paths; verify canonical identity and exact forward-slash `res://` output.
- [x] 2.2 Implement canonical-root construction with production/injected stat and symlink-evaluation functions plus segment-aware display conversion; verify targeted tests return no display value for unsafe inputs and do not re-run project discovery.

## 3. Containment and Symlink Security

- [x] 3.1 Add tests for in-project parents, lexical `..` escape, absolute outside paths, sibling-prefix collisions, existing symlinks inside/outside the root, and missing targets below safe/escaping ancestor symlinks; verify no rejected candidate produces a resolved canonical path.
- [x] 3.2 Implement the shared lexical/canonical containment pipeline using `filepath.Rel`, nearest-existing-ancestor traversal, and symlink evaluation; verify targeted tests distinguish outside-project, missing, unsupported non-regular, and filesystem failures with paths and wrapped causes.
- [x] 3.3 Add injected failure/recording tests and a wrong-case fixture that runs only when the host proves case-sensitive behavior; verify containment never scans directories, repairs case, opens target contents, or walks above the filesystem root.

## 4. Root Scene Resolution

- [x] 4.1 Add tests for absolute, cwd-relative, and `res://` root scenes plus empty, missing, non-regular, wrong-extension, relative-cwd, inaccessible, lexical-escape, and symlink-escape inputs; verify successes preserve all three identities and every failure is a typed fatal error.
- [x] 4.2 Implement root-scene input classification and resolution on the shared secure pipeline; verify only existing regular exact-lowercase `.tscn` files inside the canonical project root resolve.

## 5. Declared Resource Resolution

- [x] 5.1 Add tests for project-relative, declaring-scene-relative, parent-relative, and host-absolute resource targets of multiple extensions; verify each successful base, canonical path, display path, and raw value is exact.
- [x] 5.2 Add table-driven tests for empty, `uid://`, `user://`, unknown schemes, missing, non-regular, outside, inaccessible, and invalid-declaring-scene inputs; verify each returns the specified nonfatal unresolved reason without a Go error control path.
- [x] 5.3 Implement resource scheme classification, canonical declaring-scene validation, and nonfatal resolution assembly; verify ordinary existing resources resolve regardless of extension while unsupported and unsafe references retain actionable evidence.

## 6. Verification and Delivery

- [x] 6.1 Audit package imports and filesystem/process operations; verify resolution remains in `internal/project`, no code opens contents, prints, exits, scans for case recovery, or depends on Godot, network, parser, graph, report, CLI, OpenSpec, or Node.js runtime code.
- [x] 6.2 Run `gofmt` and verify `go build ./...`, targeted project tests, `go test ./...`, `go test -race ./...`, and `go vet ./...` all pass.
- [x] 6.3 Run `golangci-lint` at the CI-pinned version and `openspec validate issue-7-secure-path-resolution --strict`; verify zero lint issues and accurate task completion.
- [x] 6.4 Keep production code and tests in separate focused commits, push them plus OpenSpec progress to Draft PR #32, and verify the PR diff remains limited to issue #7 and its planning artifacts.
