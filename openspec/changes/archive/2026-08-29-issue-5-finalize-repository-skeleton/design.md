## Context

See [proposal.md](proposal.md#why) for motivation and [repository-foundation/spec.md](specs/repository-foundation/spec.md) for the behavior contract.

The current repository already has a thin `cmd/deadweight.gdt` entry point, Cobra orchestration, domain packages for metrics, diagnostics, budgets and presets, plus a streaming TSCN parser. Metric constants and order are correct but IDs, labels, validity, and value access are represented by separate structures. Diagnostics use raw code strings and expose no catalog or validation. `tscn.ParseError` is typed and source-aware, but its `Code` field duplicates `SB2001` as a private string constant. Cross-platform CI exists and tests compile the module, although the build gate and several skeleton CLI/domain invariants are not explicit.

Constraints come from MVP sections 9–11, 20, 28, 31, 32, and 36: deterministic output, signed non-negative 64-bit metrics, stable diagnostic codes, no stack traces, no Godot/runtime service dependency, and separation among filesystem, parser, policy, orchestration, and presentation concerns.

## Goals / Non-Goals

**Goals:**

- Make each metric and diagnostic catalog internally single-sourced and externally immutable.
- Let callers classify typed failures with `errors.As` semantics rather than message parsing.
- Add focused domain and CLI invariant coverage without introducing architecture-test dependencies.
- Preserve current command output and parser error messages except for the specified coded CLI rendering path.

**Non-Goals:**

- Create `app`, `config`, `project`, `report`, or `scene` packages before their behavior is implemented.
- Generalize all package errors into one universal error struct.
- Implement project filesystem access, recursive analysis, diagnostic sorting in reports, or complete exit-code handling for future commands.

## Decisions

### 1. Use ordered definition slices as catalog sources of truth

`metrics` will retain its public `Name` constants and `Values` fields for compatibility, but one private ordered collection of immutable definitions will pair every ID with its console label. Public catalog/order accessors will return defensive copies. `Label`, `Valid`, value access, and value validation will resolve against the same catalog. With only eight entries, linear lookup is clearer than maintaining a second map whose completeness could drift.

`diagnostic` will follow the same pattern with `Code`, `Severity`, and immutable code definitions ordered `SB1001` through `SB2004`. Severity will be part of each definition, making code/severity consistency a catalog property rather than a naming convention inferred independently at each call site.

Alternative considered: keep maps plus hand-maintained order slices. Rejected because the issue is specifically freezing completeness and order, and duplicate structures make partial updates easy.

### 2. Validate complete domain records at their boundaries

`metrics.Values` will expose validation that visits every canonical metric and returns a typed value error containing the invalid metric ID and value. Zero is valid. This validation does not replace checked arithmetic in the later analyzer slice; it prevents invalid aggregate values from silently crossing a domain boundary.

`diagnostic.Diagnostic` will use the typed `Code` field and validate known code, known severity, code/severity consistency, and non-negative source coordinates and occurrence count. Empty optional location/resource fields remain valid. Validation errors will retain field context and be testable with `errors.As` where useful.

Alternative considered: constructors that make invalid values impossible. Rejected for this slice because the structs are intentionally JSON-friendly value models and later loaders/reporters need zero-value composition; explicit validation is a better compatibility fit.

### 3. Share codes through a narrow coded-error protocol

The diagnostic package will define a narrow error protocol whose method returns a `diagnostic.Code`, plus a helper that finds that protocol through wrapped error chains. It will not own parser-specific location or configuration-specific context.

`tscn.ParseError.Code` will change from `string` to `diagnostic.Code`, retain its existing source position and message fields, and implement the protocol. The parser will use the shared `SB2001` constant. Existing `Error()` formatting remains unchanged because the named code type formats as its underlying string. An optional code-free diagnostic-message method lets presentation code preserve source context without repeating the stable code already supplied by the typed protocol.

Alternative considered: replace every typed domain error with a single `diagnostic.Error`. Rejected because parser positions, future cycle chains, configuration paths, and overflow operands have different structured context; a narrow protocol preserves those useful types.

Alternative considered: leave `ParseError.Code` as a string and add conversion only in CLI. Rejected because it would preserve duplicated constants and allow unknown codes inside typed errors.

### 4. Keep process policy centralized in CLI

CLI orchestration will recognize code-bearing errors through the shared protocol, obtain code-free text through the optional diagnostic-message contract when available, render `ERROR <code>: <message>` without a stack trace, and retain exit code `2` for fatal failures. Unknown/untyped errors keep the existing `ERROR: <message>` fallback until later slices add their domain mappings. Domain packages continue returning values or errors and receive in-memory data/readers; they do not write process streams or exit.

Alternative considered: make domain errors render themselves. Rejected because presentation, color, path display, and stream policy belong to the CLI/report layers.

### 5. Verify boundaries with focused tests and an explicit audit

Table-driven tests will exhaustively compare catalog IDs, labels, order, validity, defensive-copy behavior, code severity, record validation, negative metric handling, and wrapped typed-error discovery. Existing parser tests will assert the shared code type. CLI tests will cover help/version/placeholders and coded error rendering through the orchestration boundary.

The implementation will audit imports and forbidden process/filesystem calls in domain packages as a review task instead of adding a brittle source-scanning architecture test. GitHub Actions will add an explicit `go build ./...` gate while retaining multi-platform test/vet and Linux race/lint gates.

Alternative considered: add AST/import-rule tests. Rejected for now because the package set is small, no forbidden dependency currently exists, and source-pattern tests would either miss indirect behavior or over-constrain legitimate uses such as formatting errors.

### 6. Do not materialize the final directory tree ahead of behavior

The MVP repository diagram is a target architecture, not a requirement for empty packages. This change will document/audit the dependency direction and modify only packages with current foundation behavior. Project, analysis, configuration, and reporting packages remain owned by their later focused issues.

Alternative considered: add empty packages and placeholder types for the entire target tree. Rejected because they provide no executable contract, make ownership ambiguous, and increase churn for downstream slices.

## Risks / Trade-offs

- [Linear catalog lookup adds repeated small scans] → Each catalog has at most eleven entries, so deterministic single-source behavior is more valuable than a map optimization; benchmarks are unnecessary at this scale.
- [Changing `ParseError.Code` to a named type can break internal comparisons or composite literals] → The underlying representation remains string-compatible for constants and formatting; update all repository call sites and compile every package.
- [Validation APIs can be skipped by future callers] → Exercise them at domain boundaries as those boundaries are introduced and keep invariant tests adjacent to the value types.
- [A generic coded-error renderer may expose a code twice if legacy `Error()` text already includes it] → Let such errors provide code-free diagnostic text while preserving their existing `Error()` compatibility; the CLI renderer remains the only output-prefix owner.
- [Manual boundary audits can regress later] → Re-evaluate import/process rules in each downstream issue; add automated architecture enforcement only if package growth makes review unreliable.

## Migration Plan

1. Add exhaustive tests for the metric and diagnostic contracts, including typed wrapped-error discovery.
2. Introduce the single-source metric definitions and validation while preserving existing identifiers and accessors.
3. Introduce typed diagnostic definitions, record validation, and the narrow coded-error protocol.
4. Migrate `tscn.ParseError` to the shared `SB2001` code and update parser tests without changing its rendered message.
5. Centralize coded fatal rendering in CLI and expand skeleton command tests.
6. Add the explicit build gate, audit package boundaries/runtime dependencies, then run build, test, race, vet, and lint verification.

Rollback is a normal Git revert of the focused implementation commits. No persisted data, config schema, or external API migration is involved.
