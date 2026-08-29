## 1. Report Primitives

- [x] 1.1 Add pure report options, locale-independent integer formatting, explicit metric groups, reliability markers, and ANSI token styling, and verify focused helper tests cover exact/lower-bound/approximate/plain/color cases
- [x] 1.2 Add cloned diagnostic and unresolved projections with full deterministic sorting/grouping tie-breakers, and verify tests prove repeated rendering is byte-stable and does not mutate inputs
- [x] 1.3 Add deterministic fatal error rendering for coded multiline and uncoded failures, and verify stderr fixtures contain one prefix/code, stable indentation, and no stack trace or ANSI

## 2. Command Reports

- [x] 2.1 Implement complete, lower-bound, and approximate inspect layouts with metric groups, coverage, optional unresolved/diagnostic sections, and reliability warnings, and verify focused golden files match byte-for-byte
- [x] 2.2 Implement passed, failed, and incomplete check layouts with policy metadata, canonical comparisons, deltas, summaries, reliability explanations, and preset disclaimer, and verify focused golden files match byte-for-byte
- [x] 2.3 Implement preset list/show reports using shared metadata and integer formatting while preserving product order, and verify existing and golden preset expectations pass

## 3. CLI Integration

- [x] 3.1 Add injectable environment/terminal detection and compute color eligibility from terminal status, `--no-color`, and `NO_COLOR`, and verify all four policy combinations without host-terminal assumptions
- [x] 3.2 Route inspect, check, preset, and fatal error output through `internal/report` while preserving report-first non-fatal exit signals and thin `main.go`, and verify command tests cover exit codes `0/1/2/3`

## 4. Verification

- [x] 4.1 Run focused report/CLI tests and golden comparisons, and verify complete, lower-bound, approximate, pass, fail, incomplete, preset, coded-error, and no-color states are covered
- [ ] 4.2 Run `gofmt`, `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, pinned golangci-lint v2.12.0, and strict OpenSpec validation, and verify all tasks and gates pass
