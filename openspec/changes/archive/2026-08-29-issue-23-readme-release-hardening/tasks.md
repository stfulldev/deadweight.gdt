## 1. Public Release Documentation

- [x] 1.1 Rewrite `README.md` into the twelve §33.5 sections using only shipped commands, flags, metrics, configuration fields, support boundaries, and roadmap scope; verify every required section is present and no early-skeleton or not-implemented claim remains.
- [x] 1.2 Include the heuristic-preset, Steam Deck/Valve, target-hardware profiling, static-analysis, runtime-created node, and imported/binary scene disclaimers; verify forbidden FPS, compatibility, certification, and profiler promises do not appear as product claims.
- [x] 1.3 Document `go install ...@v0.1.0`, future release binaries, quick-start flows, all eight metric definitions, frozen preset values, custom profiles/overrides, strict JSON v1 config, reliability, exit codes, CI, roadmap, contribution workflow, and MIT license; verify examples match CLI help, embedded preset JSON, and the canonical schema.

## 2. Release Metadata and Evidence

- [x] 2.1 Replace the early Unreleased changelog with a dated `0.1.0` entry covering the shipped MVP, frozen experimental preset values, and known limitations; verify all three preset tables match embedded JSON exactly.
- [x] 2.2 Add `docs/RELEASE_0.1.0_CHECKLIST.md`, link dependency #19 and the complete acceptance matrix, record local/CI evidence rules and exact post-merge tag steps, and update criterion 25 to verified; verify the matrix still contains exactly 25 numbered rows.
- [x] 2.3 Verify local and remote `v0.1.0` refs are absent before release and document the annotated-tag target/verification commands; do not create the tag until PR #46 is merged to the verified `main` commit.

## 3. Benchmark Baselines

- [x] 3.1 Add `BenchmarkParseRepresentativeScene` with representative in-memory parser input, semantic validation outside the timed loop, and allocation reporting; verify `go test ./internal/tscn -run '^$' -bench '^BenchmarkParseRepresentativeScene$' -benchtime=100ms` passes.
- [x] 3.2 Add `BenchmarkRecursiveRepeatedScene100` with one child mounted 100 times and one normal per-invocation cache; verify exact metrics/cache semantics and `go test ./internal/analysis -run '^$' -bench '^BenchmarkRecursiveRepeatedScene100$' -benchtime=100ms` passes.

## 4. Release Verification

- [x] 4.1 Run `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, golangci-lint v2.12.0, both benchmark smoke commands, and README/changelog consistency checks; verify every repository-controlled gate is green.
- [x] 4.2 Run strict OpenSpec validation and confirm `openspec instructions apply --change issue-23-readme-release-hardening --json` reports every task complete before archive, exact-head PR merge, release tag, and tracker closure.
