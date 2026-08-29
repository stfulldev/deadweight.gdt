## Context

See `proposal.md` for motivation and `specs/built-in-heuristic-presets/spec.md` for the behavior contract. The repository already embeds three JSON files and loads them once in `internal/preset`, but the loader validates only identity, lifecycle labels, positive target FPS, and budget count. The existing tests pin only product order and Steam Deck limits, while `Catalog.Find` returns budget pointers that alias the caller's catalog and reports misses as a boolean that forces CLI code to construct its own error.

The implementation must refine this existing package rather than create a second catalog. It must preserve the standalone Go binary, the documented CLI output, and the `internal/preset` to `internal/budget` dependency direction.

## Goals / Non-Goals

**Goals:**

- Keep embedded JSON as the single runtime source for built-in records while making decoding and validation independently testable.
- Centralize deep-copy behavior so catalog retrieval and lookup share the same isolation guarantee.
- Centralize actionable lookup failures in the preset domain and remove hard-coded available-ID text from CLI callers.
- Make any catalog value change visible through an exact, readable test failure.

**Non-Goals:**

- Introduce a generic configuration validator, schema generator, runtime JSON files, or a new package dependency.
- Redesign the budget model or CLI formatting.
- Enforce release-note policy through a source-tree-reading runtime or unit test.

## Decisions

### 1. Retain JSON plus `go:embed` as the only runtime catalog source

The three existing files under `internal/preset/data` remain the authoritative runtime data and the fixed product-order ID slice controls loading order. The loader will continue to cache either the validated catalog or its error with `sync.Once`, so a corrupt build fails consistently rather than serving a partial catalog.

Alternative considered: duplicate the frozen values in Go constants and compare embedded JSON against them at runtime. Rejected because two runtime sources could drift and would weaken the purpose of version-controlled embedded data. Exact independent expectations belong in tests.

### 2. Separate decoding and validation from the cached built-in entry point

Refactor the loader into testable helpers that accept record bytes or an `fs.FS`, decode exactly one JSON object with unknown fields rejected, and validate the resulting preset. `Builtins` remains the cached production entry point over the embedded filesystem.

Validation will require non-empty identity and descriptive metadata, the expected file ID, unique IDs, positive target FPS, the fixed lifecycle labels, allowed renderer and quality IDs from MVP section 22.1, and all eight configured limits with values greater than or equal to zero. Errors will wrap the preset ID and name the invalid field or condition. Validation will iterate `metrics.OrderedNames()` so the metric set stays aligned with the budget model without adding a reverse package dependency.

Alternative considered: test only the committed production files. Rejected because malformed renderer, quality, and limit cases cannot be exercised without either mutating source fixtures or exposing deterministic validation helpers to package-level tests.

### 3. Use one deep-copy path for catalog and lookup values

Add a single preset clone operation that copies the struct and deep-clones `budget.Limits`. `Builtins` will clone every record and the lookup path will clone the matched record again, preventing mutation of package state and of the caller-owned source catalog respectively.

The preset domain lookup will return `(Preset, error)` instead of `(Preset, bool)`. On a miss it will build an error from the requested ID and IDs observed in the catalog's existing order. CLI code will propagate that error rather than duplicate the product list. All in-repository callers will be updated together; the package is under `internal/`, so this is not a public Go compatibility break.

Alternative considered: keep boolean lookup and format errors only in the CLI. Rejected because non-CLI callers would not receive the issue's actionable-error contract and the available-ID list could diverge from catalog order.

### 4. Pin the complete catalog with explicit table-driven expectations

Replace the partial Steam Deck assertion with an expected catalog containing all metadata and all eight limits for all three presets. Compare records in order and retain focused tests for lifecycle labels, embedded loading, catalog-copy isolation, lookup-copy isolation, and exact unknown-ID errors. Invalid JSON cases will use in-memory test files or direct decode helpers so production data remains untouched.

The expectation intentionally duplicates frozen values in test code: that duplication is the change detector required by MVP section 29.4. Updating both data and expectation remains a deliberate review event. The patch-release `CHANGELOG.md` rule stays a contributor/review obligation; automated source-tree inspection would only prove that some changelog text exists, not that it explains a particular correction.

Alternative considered: golden snapshots generated from the same JSON files. Rejected because self-generated snapshots would accept the very catalog drift the tests must detect.

### 5. Audit positioning without expanding CLI scope

Existing README and preset list/show wording will be checked against the non-certification requirement. Tests will preserve `heuristic`, `experimental`, and the performance-guarantee disclaimer on user-visible preset surfaces. No benchmark, certification, or Valve-endorsement language will be added.

## Risks / Trade-offs

- [Changing the internal lookup signature breaks an overlooked caller] → Search all `Catalog.Find` call sites and rely on the Go compiler plus `go test ./...` to catch incomplete migration.
- [Cached `sync.Once` state makes negative cases order-dependent] → Test pure decode/load helpers with isolated in-memory inputs and reserve `Builtins` tests for the production catalog.
- [Exact expected structs are verbose and require intentional maintenance] → Keep a single readable expected-catalog fixture in tests and treat changes as frozen-data review events.
- [Validation errors reveal only the first corrupt record] → Fail fast with preset and field context; the shipped catalog is build-time-controlled and partial results are never returned.

## Migration Plan

1. Add exhaustive tests and testable validation boundaries around the current embedded catalog.
2. Tighten validation and centralize cloning in `internal/preset`.
3. Change lookup to return domain errors and update the CLI caller and tests.
4. Run the repository quality gates and review README/CLI/CHANGELOG wording against the positioning and release rules.

There is no persisted data or external API migration. Rollback is a normal commit revert; the three embedded JSON records and existing CLI behavior remain compatible throughout.
