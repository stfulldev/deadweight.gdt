## 1. Format-3 Parser Compatibility

- [x] 1.1 Add lexer regression tests for literal LF, CRLF, and CR multiline strings, exact token content, following-token positions, multiline EOF, and post-line unsupported escapes; verify `go test ./internal/tscn -run 'TestLexer'` fails only for the missing behavior before implementation.
- [x] 1.2 Update the streaming string lexer to retain physical line endings and preserve typed malformed-string failures; verify the targeted lexer tests and existing fuzz seeds pass.
- [x] 1.3 Add parser regression tests for balanced unknown quoted properties, following sections, quoted names without assignments, and unchanged `shadow_enabled` semantics; update only section-body property-name acceptance and verify `go test ./internal/tscn` passes.

## 2. Build Version Provenance

- [x] 2.1 Add composition-root tests for explicit linker precedence, semantic module versions with one leading `v`, pseudo-versions, empty/`(devel)` metadata, and the `dev` fallback; verify the new tests fail only for the missing resolver before implementation.
- [x] 2.2 Resolve the command version from explicit linker input and embedded Go build metadata without changing CLI/report packages or adding dependencies; verify `go test ./cmd/deadweight.gdt` and a local `go run ./cmd/deadweight.gdt --version` pass.
- [x] 2.3 Build with an explicit linker version and build/install the fixed source through versioned Go module metadata after its commit is available remotely; verify the explicit value wins, the module-derived value drops one leading `v`, and the immutable `v0.1.0` ref is unchanged.

## 3. Real-Project E2E Coverage and Documentation

- [x] 3.1 Add an opt-in PowerShell E2E runner that accepts explicit binary/demo paths, inspects every declared main scene, classifies COMPLETE, PARTIAL, format-4, and UID-root outcomes, and exits nonzero for other fatals; verify help/argument failures and a run against the recorded corpus commit are deterministic.
- [x] 3.2 Build the fixed CLI and run the E2E runner against `godot-demo-projects` commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`; verify supported format-3 closures have no unexpected fatal result and record the complete category counts in Draft PR #48.
- [x] 3.3 Replace the empty Unreleased changelog entry with the parser and version-provenance fixes, explicitly preserving the format-4, UID-root, project-scan, and immutable-v0.1.0 boundaries; verify the text matches the implemented and validated behavior.

## 4. Quality Gates and Delivery

- [x] 4.1 Run `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and golangci-lint v2.12.0; verify every repository-controlled gate passes on the implementation commit.
- [x] 4.2 Run strict OpenSpec validation, verify every issue #47 acceptance criterion and task, sync the delta specs, and archive `issue-47-real-world-format3-compatibility`; verify the archived capability specs describe shipped behavior.
- [x] 4.3 Commit the archive result separately, push every commit, update Draft PR #48 with exact unit/E2E/gate evidence, and mark it ready only after the remote head and local verified head match.
