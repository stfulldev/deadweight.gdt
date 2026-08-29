# Why

Issue [#5](https://github.com/stfulldev/deadweight.gdt/issues/5) is the remaining foundation gate for the MVP implementation. The repository already has a thin executable, Cobra wiring, parser contracts, and the eight metric identifiers, but the metric catalog is not exhaustively protected by tests and the diagnostic package does not yet expose the stable typed code catalog required by the MVP specification. Freezing these contracts now keeps downstream project-resolution, analysis, configuration, and reporting work from inventing incompatible domain primitives.

Planning and implementation progress are tracked in linked Draft PR [#30](https://github.com/stfulldev/deadweight.gdt/pull/30).

# What Changes

- Freeze the eight MVP metric IDs, display labels, and canonical order as one reusable catalog, with table-driven invariant tests that cannot depend on Go map iteration.
- Define the diagnostic domain contract around typed severities and typed stable codes `SB1001` through `SB2004`, including validity and category invariants suitable for downstream errors and reports.
- Preserve typed TSCN parse errors while aligning them with the shared diagnostic code type rather than duplicating a raw code string.
- Freeze the repository dependency boundaries: the executable remains a thin process adapter, CLI owns command/process orchestration, domain packages remain free of scene filesystem reads and console output, and future packages are introduced only with concrete behavior.
- Verify the Milestone 0 build, test, vet, help, version, placeholder-command, determinism, and standalone-runtime acceptance criteria across the existing Go-first skeleton.

Goals:

- Establish stable domain vocabulary that later MVP slices can safely import.
- Make catalog completeness and deterministic ordering executable through tests.
- Keep filesystem, parser, policy, application orchestration, and presentation responsibilities separated.

Non-goals:

- Implement project discovery, path resolution, recursive scene analysis, configuration loading, reports, or final command behavior.
- Change metric definitions, preset values, CLI command names, or the MVP exit-code contract.
- Add empty placeholder packages solely to mirror the proposed final repository tree.

Compatibility impact:

- Existing metric IDs, labels, ordering, JSON fields, CLI commands, and parser error text remain stable.
- The diagnostic and parser APIs gain stronger named types; internal callers may require mechanical type updates, but no user-visible output change is intended.
- The application remains a standalone Go binary with no Godot runtime, Node.js, network, or OpenSpec runtime dependency.

Affected MVP acceptance criteria:

- Sections 9–11: package direction, core metric and diagnostic domain models, typed errors, and `int64` counters.
- Section 20 and Section 28: exactly eight metrics in deterministic canonical order.
- Section 31, Milestone 0 and Section 32, Step 1: buildable CLI skeleton, domain invariants, and table-driven tests.
- Section 36: frozen metric IDs and standalone Go-first constraints.

# Capabilities

## New Capabilities

- `repository-foundation`: Defines the stable metric and diagnostic catalogs, typed parser-error interoperability, package responsibility boundaries, and repository-skeleton verification gates for MVP 0.1.

## Modified Capabilities

- None.

# Impact

- Affected code: `internal/metrics`, `internal/diagnostic`, `internal/tscn`, and their tests; the executable and CLI are verification targets rather than feature-expansion targets.
- Affected CI and developer checks: `go build ./...`, `go test ./...`, `go vet ./...`, and cross-platform GitHub Actions coverage.
- No schema, configuration, preset-data, network, Godot-runtime, or public release-format changes are expected.
