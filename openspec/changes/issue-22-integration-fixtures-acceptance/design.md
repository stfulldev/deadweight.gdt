## Context

See `proposal.md` for motivation. The repository already has focused unit coverage for parser states, secure paths, recursion/cache behavior, policy resolution, budget outcomes, and report rendering. What is missing is a shared set of filesystem fixtures and production-composition tests proving those layers work together. Existing report goldens use constructed domain values rather than the real application pipeline, and the CI matrix runs race testing only on Linux.

The fixture suite must be portable across Linux, macOS, and Windows, must not invoke Godot, and must not make output depend on an absolute checkout path. The shipped packages and frozen MVP behavior are not design targets for this change.

## Goals / Non-Goals

**Goals:**

- Exercise production `app.NewDefault()` composition and the real CLI command tree against committed Godot projects.
- Make fixture intent, exact metrics, reliability, diagnostics, and exit outcomes reviewable without reverse-engineering scene files.
- Keep golden snapshots byte-stable by replacing the canonical fixture root with `<PROJECT>` before comparison.
- Close test-corpus gaps without copying large scenarios already proven by focused unit tests.
- Run build, unit/integration tests, race tests, and vet on every supported GitHub-hosted operating system.

**Non-Goals:**

- Adding a public testing API, changing dependency injection boundaries, or exporting internal helpers.
- Making snapshots hide nondeterminism other than the unavoidable absolute project-root prefix and host path separator.
- Treating an external GitHub-hosted runner startup failure as repository test evidence.
- Adding Godot-generated binary/import artifacts to fixtures.

## Decisions

### 1. Store one shared fixture corpus at the repository root

Create the seven groups from §29.5 under `testdata/projects`. Each group owns a `README.md` that names its entry scenes and expected result. Scene files remain minimal hand-authored `format=3` text so the repository neither installs nor depends on Godot.

Alternatives considered:

- Package-local duplicate fixtures would make `go test` lookup convenient, but would fragment the acceptance corpus and allow application and CLI tests to drift.
- Generating every project in `t.TempDir()` would cover behavior but leave no reviewable reusable artifacts.

### 2. Test real application and CLI composition in two layers

Expand `internal/app/default_integration_test.go` with table-driven exact assertions for fixture metrics, reliability, cache counts, relative resolution, repeated occurrences, unresolved evidence, inheritance, cycles, and malformed roots. Add production-path CLI golden tests in `internal/cli` that call `ExecuteWithApplicationAndRuntime` with `app.NewDefault()` and explicit project paths.

The CLI helper normalizes only canonical occurrences of the selected fixture root to `<PROJECT>` and converts remaining host separators in path evidence to `/`. It compares stdout, stderr, and exit code as one snapshot contract. Tests fail if an unexpected temporary or checkout path remains.

Alternatives considered:

- Extending only `internal/report` goldens would not prove project discovery, configuration, analysis, policy, or exit mapping.
- Spawning the compiled binary would test `os.Exit`, but slows and complicates cross-platform tests while adding little over the real Cobra/application composition root.

### 3. Use focused fixtures for end-to-end states and retain unit tests as narrow evidence

The end-to-end golden matrix covers complete, partial lower-bound, inherited approximate, pass, fail, partial rejected, cycle, invalid config, missing project, and terminal output suppressed by `NO_COLOR`. Existing unit tests remain the primary evidence for exhaustive parser permutations, path boundary/symlink cases, overflow, four-layer policy merge, and immutable built-in values.

This avoids inflating every golden whenever an internal diagnostic gains unrelated evidence while still demonstrating every user-visible outcome required by #22.

### 4. Strengthen fuzz regression inputs without adding production limits

Add seeds for every parser shape named in §29.1, especially malformed delimiters, comments in strings, unknown multiline values, inheritance, duplicate IDs, and large packed arrays. The Go fuzz engine supplies arbitrary bytes and detects panics; normal test timeouts detect hangs. A deterministic large-blob regression test verifies that skipped unknown data is not retained in the parsed model.

Alternatives considered:

- Imposing a new parser input-size cap would change the runtime contract and belongs in a separate behavior change.
- Measuring heap bytes inside the fuzz callback would be nondeterministic under parallel GC and unsuitable for the cross-platform gate.

### 5. Keep the acceptance matrix human-readable and mechanically referential

Add `docs/MVP_0.1_ACCEPTANCE.md` with one row per §30 criterion, stable test symbol or document links, and verification commands. Criterion 25 points to the release documentation work in #23 until that change lands; the row remains traceable rather than being silently marked complete.

Alternatives considered:

- Encoding the matrix only as Go comments would be difficult for release review and would not provide documented evidence for non-code criteria.
- A custom matrix parser/linter would add maintenance cost without increasing behavioral confidence.

### 6. Put the complete Go quality gate in the OS matrix

Each Linux, macOS, and Windows job runs `go build ./...`, `go test ./...`, `go test -race ./...`, and `go vet ./...`; lint stays a separate Linux job. The official action major refs already resolve through GitHub, so this change does not speculate about unrelated hosted-runner failures. Jobs receive explicit timeouts and no Godot setup step.

## Risks / Trade-offs

- [Golden snapshots become noisy when intentional report text changes] → Keep production-path goldens limited to frozen MVP states and preserve smaller report unit tests for formatting internals.
- [Cross-platform canonical paths differ] → Normalize only the known fixture root and separators, and assert that no absolute fixture root leaks into snapshots.
- [Symlink creation is unavailable on some Windows runners] → Keep exhaustive symlink behavior in existing platform-aware resolver tests; committed acceptance fixtures use portable regular files.
- [Race testing all three operating systems increases CI time] → Use one compact matrix, fail independently, and set bounded job timeouts.
- [A GitHub-hosted runner can fail before repository steps begin] → Record zero-step runs separately; do not weaken or remove local gates to manufacture a green result.

## Migration Plan

1. Commit fixtures and fixture documentation without changing production code.
2. Add integration/golden/fuzz coverage and the acceptance matrix.
3. Update the CI matrix and run the complete local quality gate.
4. Roll back by reverting test/tooling commits; no runtime data or user migration is required.
