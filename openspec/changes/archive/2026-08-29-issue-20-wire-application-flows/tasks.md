## 1. Application Orchestration

- [x] 1.1 Add owned inspect/check request and result models plus preset results in `internal/app`, and verify package compilation preserves all required project, config, analysis, policy, and evaluation evidence
- [x] 1.2 Implement injectable application dependencies with production defaults for project discovery, secure resolution, config loading, recursive analysis, policy resolution, partial policy, budget evaluation, and presets, and verify focused application tests can replace every external effect
- [x] 1.3 Implement inspect and check sequencing with fatal short-circuiting, optional implicit config, CLI-over-config policy precedence, and report-ready results, and verify focused tests cover complete, partial, overridden, empty-budget, and fatal paths
- [x] 1.4 Implement project-independent preset list/show application calls and verify tests prove project, config, and analysis effects are not invoked

## 2. CLI Wiring

- [x] 2.1 Replace placeholder commands with an injected four-flow application interface and verify `main.go` remains a thin caller of the default executor
- [x] 2.2 Add global project/config/no-color flags and all frozen check flags with exact positional counts, repeated budget ordering, and mutual exclusions, and verify invalid syntax never invokes the application
- [x] 2.3 Add compact deterministic inspect/check presentation and route existing preset presentation through application results, and verify successful reports use injected output streams without ANSI
- [x] 2.4 Centralize fatal, failed-budget, rejected-partial, and success exit mapping and verify reports are emitted before non-fatal codes `1` and `3`

## 3. Test Coverage

- [x] 3.1 Add application unit tests for absolute/relative/`res://` requests, optional configuration, analysis/policy/evaluation sequencing, ownership, and Godot-free injected effects, and verify `go test ./internal/app` passes
- [x] 3.2 Add CLI command tests for request forwarding, global/check flags, exact arguments, mutual exclusions, preset independence, diagnostic rendering, and codes `0/1/2/3`, and verify `go test ./internal/cli` passes
- [x] 3.3 Add a production-composition integration test over temporary Godot text fixtures for inspect and check, and verify it runs without a Godot executable

## 4. Verification

- [x] 4.1 Run `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...`, the repository-pinned `golangci-lint`, `openspec validate issue-20-wire-application-flows --strict`, and `openspec status --change issue-20-wire-application-flows --json`, and verify every gate passes with all tasks complete
