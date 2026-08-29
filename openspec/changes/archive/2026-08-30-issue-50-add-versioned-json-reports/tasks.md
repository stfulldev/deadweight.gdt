## 1. Version-One Wire Contract

- [x] 1.1 Add private version-one inspect, check, shared-analysis, policy/evaluation, diagnostic, and error wire models in `internal/report`; verify focused projection tests cover all required fields, canonical metric/comparison order, checked integer invariants, and caller-owned slices.
- [x] 1.2 Implement portable scene/configuration/diagnostic projection that omits canonical checkout paths and normalizes in-project display identities; verify equivalent temporary checkouts and Windows-style source separators produce identical portable model values.
- [x] 1.3 Add `schema/deadweight.gdt.report-v1.schema.json` plus a test-only Draft 2020-12 validation harness; verify representative inspect, check, and error documents validate while missing discriminators, mixed kind payloads, invalid enums, and unsupported schema versions are rejected.
- [x] 1.4 Implement buffered standard-library JSON encoding with disabled HTML escaping, deterministic two-space indentation, LF line endings, and one trailing LF; verify repeated renders are byte-identical, contain no ANSI/stack text, and do not mutate input results.

## 2. Scene Command Format Selection

- [x] 2.1 Add a validated `text|json` presentation format and local `--format` flags to `inspect` and `check`, defaulting to text; verify unknown values fail with exit `2` before the injected application is called and preset commands do not silently accept the scene-only flag.
- [x] 2.2 Route successful and non-fatal inspect/check outcomes through the selected renderer without changing application request types or results; verify paired text/JSON command tests make equivalent requests and retain exit codes `0`, `1`, and `3`.
- [x] 2.3 Route every fatal after reliable JSON selection through one kind `error` document on stderr while retaining text for pre-selection parse failures; verify coded cycle/parser errors, uncoded errors, empty stdout, no duplicate prose, and exit code `2`.
- [x] 2.4 Preserve existing color/runtime behavior only for text and make JSON color-independent; verify terminal, `--no-color`, `NO_COLOR`, and redirected JSON executions produce identical bytes.

## 3. Golden, Compatibility, and Documentation Evidence

- [x] 3.1 Add schema-valid JSON golden fixtures for complete inspect, lower-bound inspect, approximate inspect, passing check, failed check, partial-rejected check, coded fatal, and uncoded fatal outcomes; verify every fixture is deterministic and semantically matches its existing text counterpart.
- [x] 3.2 Add integration coverage for absolute, relative, and `res://` scene inputs across different checkout roots, maximum-domain integers, grouped diagnostics, built-in heuristic metadata, and explicit/implicit/absent configuration provenance; verify no successful document contains a canonical temp root or OS-specific separator.
- [x] 3.3 Document `--format json`, stream/exit behavior, schema versioning, portable identity, signed 64-bit integers, and the committed schema in README/roadmap material; verify examples are produced by current fixtures and do not claim baseline, tree, SARIF, project-wide, or runtime-profiler behavior.

## 4. Quality Gates and Delivery

- [x] 4.1 Run focused `internal/report` and `internal/cli` tests plus `git diff --check`; verify all new schema, encoder, command, stream, immutability, and compatibility cases pass on the implementation head.
- [x] 4.2 Run `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, golangci-lint v2.12.0, and `openspec validate --all --strict`; verify every repository-controlled gate passes without Godot or a runtime schema service.
- [x] 4.3 Verify every issue #50 acceptance criterion and unchanged default text golden, sync the application-command and new JSON-report capability specs, and archive `issue-50-add-versioned-json-reports`; verify the archived specs describe shipped behavior.
- [x] 4.4 Commit the archive separately, push all commits, update Draft PR #58 with exact schema/golden/gate evidence, and mark it ready only after the verified local head equals the remote head.
