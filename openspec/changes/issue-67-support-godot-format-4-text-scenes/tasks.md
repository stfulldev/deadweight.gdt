## 1. Parser Compatibility

- [x] 1.1 Add focused parser fixtures/tests for format-4 headers, base64 `PackedByteArray`, nested `PackedVector4Array`, node `unique_id`, malformed packed values, format-3 equivalence, and rejected formats 2 and 5; verify `go test ./internal/tscn` exercises every format4-scene-parsing scenario.
- [x] 1.2 Implement the explicit supported-version set `{3,4}` and deterministic missing/unsupported-format diagnostics without adding a second parser or AST path; verify all `internal/tscn` tests pass and parsed headers retain the declared version.
- [x] 1.3 Add a representative large base64 packed-value regression/benchmark and verify parsing reaches later nodes without decoding the payload or retaining a Variant AST; record allocations if they expose a concrete need for an in-scope discard optimization.

## 2. Recursive Analysis Integration

- [x] 2.1 Add analyzer and CLI integration fixtures for a complete format-4 root, a mixed format-3/format-4 nested chain, a format-4 inherited base under the existing approximate contract, and an unknown future nested format; verify focused analysis and CLI tests cover exact, approximate, and fatal outcomes.
- [x] 2.2 Route successfully parsed format-4 documents through the existing local-summary, graph, cache, recursive aggregation, completeness, contribution, tree, check, and JSON/text presentation paths with no version-specific metric branch; verify paired format-3/format-4 fixtures produce equivalent frozen metrics and confidence when their supported content is equivalent.
- [x] 2.3 Run the frozen default-text acceptance goldens and JSON schema-v1 validation tests and verify issue #67 changes neither existing format-3 bytes nor report, exit-code, preset, budget, or configuration semantics.

## 3. Official Corpus and Documentation

- [x] 3.1 Remove `UNSUPPORTED_FORMAT_4` preclassification from `scripts/e2e-godot-demo-projects.ps1`, run the standalone binary against pinned corpus commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`, and verify all nine former format-4 roots enter ordinary analysis with zero unexpected fatal outcomes.
- [x] 3.2 Commit the newly measured deterministic corpus categories in `.github/workflows/ci.yml`, keep UID-only roots independently visible, and verify the PowerShell runner fails on any count drift or format-4 rejection; confirm the hosted `official-demo-e2e` job passes on the final PR head.
- [x] 3.3 Update README and CHANGELOG current-source compatibility wording for format-3/format-4 text scenes and the remaining UID/import/inheritance boundaries; verify `docs/MVP_0.1_SPEC.md`, MVP 0.2 acceptance evidence, metric definitions, and preset values remain historically unchanged.

## 4. Verification and Delivery

- [x] 4.1 Run targeted parser, analysis, CLI, corpus-runner, acceptance-golden, and JSON-schema tests plus `git diff --check`; verify every new scenario and unchanged compatibility contract passes deterministically.
- [x] 4.2 Run `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint run`, and `openspec validate --all --strict`; verify every repository-controlled gate passes without installing or invoking Godot.
- [x] 4.3 Commit production feature changes separately from test/fixture changes, keep documentation/tooling evidence separately reviewable where practical, push every commit, and update Draft PR #73 with exact local and hosted corpus/gate evidence.
- [ ] 4.4 Sync and archive `issue-67-support-godot-format-4-text-scenes`, commit the archive separately, verify strict OpenSpec status and local/remote head equality, and mark PR #73 ready only after every task and hosted check passes.
