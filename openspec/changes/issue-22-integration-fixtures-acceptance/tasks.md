## 1. Shared Godot Fixture Corpus

- [ ] 1.1 Create the seven §29.5 project groups under `testdata/projects`, each with `project.godot`, minimal entry/dependency scenes, and an adjacent README that states expected behavior; verify a repository file listing contains all named groups and required entry scenes.
- [ ] 1.2 Encode complete, repeated, relative, unresolved, inherited, cyclic, and malformed semantics in the fixture scenes; verify targeted production application tests assert exact metrics, reliability, evidence, cycle chains, and parse-error codes for every group.

## 2. Production-Path Acceptance Tests

- [ ] 2.1 Expand `internal/app/default_integration_test.go` into a table-driven fixture suite covering absolute, relative, and `res://` roots, repeated-cache instrumentation, unique evidence, inherited approximation, unresolved lower bounds, cycles, and malformed inputs; verify `go test ./internal/app -run 'TestDefaultApplication|TestFixture'` passes.
- [ ] 2.2 Add real CLI golden snapshots for complete, partial, approximate, pass, fail, partial rejection, cycle, invalid config, missing project, and `NO_COLOR`, normalizing only the fixture root to `<PROJECT>`; verify `go test ./internal/cli -run 'TestAcceptanceGoldens'` passes without ANSI or host paths.
- [ ] 2.3 Expand the `FuzzParse` seed corpus for every §29.1 parser shape and retain a deterministic large-blob regression; verify parser tests pass and a bounded smoke fuzz run such as `go test ./internal/tscn -run '^$' -fuzz '^FuzzParse$' -fuzztime=5s` completes without panic or hang.

## 3. Traceability and Cross-Platform Gate

- [ ] 3.1 Add `docs/MVP_0.1_ACCEPTANCE.md` with one traceable row for each of the 25 §30 criteria, referencing stable test symbols or documented release evidence and explicitly tracking criterion 25 in issue #23; verify all criterion numbers 1–25 occur exactly once.
- [ ] 3.2 Update `.github/workflows/ci.yml` so Linux, macOS, and Windows each run build, test, race, and vet with bounded timeouts while lint remains independent and no Godot setup exists; verify the workflow text contains every required command under the OS matrix.

## 4. Repository Verification

- [ ] 4.1 Run `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and golangci-lint v2.12.0; verify every local gate is green.
- [ ] 4.2 Run strict OpenSpec validation and confirm `openspec status --change issue-22-integration-fixtures-acceptance` reports every task complete before sync/archive and PR merge.
