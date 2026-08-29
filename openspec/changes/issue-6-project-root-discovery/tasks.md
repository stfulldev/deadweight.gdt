## 1. Project Discovery Domain Contract

- [ ] 1.1 Add table-driven tests for discovery request/result values, stable error reasons, actionable messages, and wrapped causes; verify targeted project-package tests inspect failures with `errors.As` and `errors.Is` rather than message-only matching.
- [ ] 1.2 Implement the project root, request, error reason, and typed error models; verify the domain-contract tests compile and pass without importing CLI, parser, report, or diagnostic packages.

## 2. Explicit Project Discovery

- [ ] 2.1 Add tests for relative and absolute explicit directories, explicit `project.godot` files, missing markers, wrong filenames, non-regular entries, inaccessible paths, and no-fallback precedence; verify each explicit-project acceptance and rejection class is covered.
- [ ] 2.2 Implement explicit-project normalization and validation using only metadata inspection; verify targeted tests return absolute cleaned roots and never search ancestors after an explicit-project failure.

## 3. Filesystem and Resource-Path Discovery

- [ ] 3.1 Add tests for relative and absolute filesystem scenes, exact `.tscn` validation, `res://` cwd starts, nested projects, non-regular markers, root termination, and recorded stat traversal; verify the nearest valid marker and start directory are exact in every case.
- [ ] 3.2 Implement filesystem scene validation and `res://` classification without opening scene contents; verify missing, directory, and wrong-extension inputs produce typed invalid-scene errors while `res://` does not require the target scene to exist.
- [ ] 3.3 Implement parent traversal with `filepath.Dir` root identity and injected/production stat functions; verify nearest-root, filesystem-root termination, non-regular-marker continuation, and non-not-found inspection errors pass targeted tests.

## 4. Application Boundary and Architecture

- [ ] 4.1 Add a focused CLI boundary test that returns a typed project discovery error through command execution; verify exit code `2`, actionable stderr, and absence of a Go stack trace without wiring unfinished analyzer commands.
- [ ] 4.2 Audit package imports and filesystem/process calls; verify only `internal/project` owns root-discovery filesystem metadata, no project code prints or exits, no dormant `--project` flag or empty future package is added, and no Godot/network/runtime dependency appears.

## 5. Verification and Delivery

- [ ] 5.1 Run `gofmt` and verify `go build ./...`, targeted project/CLI tests, `go test ./...`, `go test -race ./...`, and `go vet ./...` all pass.
- [ ] 5.2 Run `golangci-lint` at the CI-pinned version and `openspec validate issue-6-project-root-discovery --strict`; verify zero lint issues and accurate task completion.
- [ ] 5.3 Keep production code and tests in separate focused commits, push them plus OpenSpec progress to Draft PR #31, and verify the PR diff remains limited to issue #6 and its planning artifacts.
