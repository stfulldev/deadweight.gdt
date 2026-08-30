## 1. Baseline decoding and compatibility

- [x] 1.1 Add owned schema-v1 baseline input models and a bounded single-document decoder in `internal/reportdiff`, and verify focused tests accept current inspect/tree/check reports plus legacy reports without per-metric confidence.
- [x] 1.2 Validate tool identity, schema, kind, portable scene identity, frozen metrics, confidence, coverage, diagnostics, dependencies, and check evaluation, and verify malformed or incompatible fixtures fail actionably without partial results.

## 2. Semantic comparison and enforcement

- [x] 2.1 Implement deterministic metric, reliability, coverage, diagnostic, dependency, and check-evaluation comparison with owned sorted collections, and verify equal and fully changed fixtures produce stable expected results.
- [x] 2.2 Implement proof-oriented regression/improvement/uncertain assessment including conservative legacy confidence fallback, and verify lower-bound and approximate combinations cannot claim unsupported improvements.
- [x] 2.3 Implement normalized opt-in metric and reliability enforcement with incomplete priority, and verify passed, failed, and incomplete outcomes map to the established statuses.

## 3. Application and CLI flow

- [x] 3.1 Add the injected two-file application diff flow with the 16 MiB input bound, and verify it reads only the requested files and performs no project, scene, configuration, analysis, Godot, or network work.
- [x] 3.2 Add `diff`, `--format`, repeatable `--fail-on-increase`, and `--fail-on-reliability` to the CLI, and verify invalid arguments, formats, duplicates, and metric names fail before file reads while valid outcomes preserve exits `0`, `1`, `2`, and `3`.

## 4. Deterministic presentation

- [x] 4.1 Add deterministic text rendering for empty and changed semantic diffs, confidence qualification, evidence changes, and report-first enforcement summaries, and verify checked-in goldens have exactly one trailing LF.
- [x] 4.2 Add schema-v1 kind `diff`, deterministic JSON rendering, and existing JSON error framing, and verify empty, changed, failed, incomplete, and malformed cases validate against the updated schema.

## 5. Documentation and integration

- [x] 5.1 Document reproducible baseline capture, comparison, CI enforcement, exit handling, and intentional baseline updates in the README, and verify every documented command matches CLI help.
- [x] 5.2 Add cross-platform CLI integration coverage for portable paths, semantic equality, deterministic changes, opt-in enforcement, and invalid input with no partial stdout, and verify the focused integration suite passes.

## 6. Quality gates

- [x] 6.1 Run `openspec validate issue-53-add-report-baselines-and-deterministic-diffs --strict` and verify every artifact and delta specification is valid.
- [x] 6.2 Run `go build ./cmd/deadweight.gdt`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and the repository-pinned golangci-lint command, and verify all repository gates pass.
