## Context

See `proposal.md` for motivation. `internal/app` now returns owned report-ready models and `internal/cli` already centralizes non-fatal outcome mapping, but formatting is split between transitional scene renderers and legacy preset helpers. The final console contract depends on domain reliability, coverage, unresolved, diagnostic, effective-policy, and budget-evaluation evidence and must remain deterministic across operating systems and output destinations.

## Goals / Non-Goals

**Goals:**

- Make report formatting a pure, independently tested projection over application results.
- Encode reliability, ordering, whitespace, number grouping, disclaimers, and verdict language in one boundary.
- Decide color eligibility from injectable environment and terminal capabilities without changing application requests.
- Preserve report-first non-fatal outcomes and deterministic fatal stderr formatting.

**Non-Goals:**

- Add report formats other than console text or change any domain status.
- Build the full cross-project fixture matrix or acceptance traceability owned by Issue #22.
- Add a terminal library, locale dependency, Godot dependency, or runtime configuration cache.

## Decisions

### Add a pure `internal/report` package

The report package will expose focused render functions for inspect, check, preset list, preset show, and fatal errors. Inputs are existing application/domain values plus immutable render options such as version and color eligibility; outputs are complete strings or validation errors. The package performs no filesystem, environment, command parsing, or application calls.

Alternative considered: continue formatting in Cobra handlers. Rejected because command syntax, terminal policy, sorting, and report contract would remain coupled and golden tests would require fake command execution for every formatting case.

### Normalize into local projections before rendering

Renderers will use explicit metric group lists, clone diagnostic and unresolved slices, group equivalent unresolved occurrences, and sort local projections with full deterministic tie-breakers. Number formatting will be a small decimal grouping function independent of locale. This prevents map or caller slice order from affecting bytes and lets tests assert the input models remain unchanged.

All eight metrics will use one reliability marker for partial results: `+` for lower bounds and `~` for approximate inheritance. This is the conservative uniform option permitted by §26.2 and avoids claiming per-metric precision the domain model does not carry.

Alternative considered: selectively mark only some metrics. Rejected because the current result has report-level reliability rather than per-metric confidence and selective markers would encode unsupported assumptions.

### Keep layout assembly explicit

Inspect output will use the frozen Structure, Rendering, Resources, and Coverage blocks. Optional unresolved and diagnostics blocks appear only when evidence exists, followed by a fixed reliability warning. Check output will use a fixed comparison table, policy metadata block, explicit summary, optional reliability explanation, and built-in preset disclaimer. Preset renderers reuse the same integer and metadata display helpers.

Alternative considered: a generic table/layout framework. Rejected because the MVP has four small fixed layouts and a generic abstraction would obscure exact spacing while adding no reuse outside reports.

### Inject color environment at CLI composition

CLI composition will own a small runtime policy containing `LookupEnv` and `IsTerminal` functions. Production defaults use `os.LookupEnv` and character-device inspection of an `*os.File`; tests can inject deterministic answers. A report receives `Color=true` only when the output is terminal and both explicit and environment suppression are absent. Non-terminal buffers are always plain.

ANSI styling will wrap only status tokens and warning labels. Plain text remains identical after stripping ANSI, so color cannot change meaning. Fatal stderr remains uncolored for stable diagnostics and redirection safety.

Alternative considered: use a third-party terminal package. Rejected because character-device detection is sufficient for the frozen MVP and a new runtime dependency is unnecessary.

### Retain centralized exit signals and centralize error text in reports

Cobra handlers will render successful application results, then return the existing private exit signal for failed/incomplete checks. The executor will continue to consume signals without treating them as fatal errors. All other errors go through the report error renderer and return `2`. Multiline coded errors will split the code-free message into a heading and consistently indented evidence lines.

Alternative considered: return exit codes from report functions. Rejected because presentation does not own application outcome policy and should remain usable without Cobra.

## Risks / Trade-offs

- [Uniform reliability markers are conservative] → Document and golden-test the permitted all-metrics marker rule until per-metric confidence exists.
- [Fixed spacing can regress during innocent refactors] → Store byte-for-byte golden files and use one intentional update path in tests.
- [Terminal detection differs across platforms] → Keep detection behind an injected function and test policy independently; non-terminal behavior remains the safe default.
- [Unresolved evidence can duplicate diagnostics] → Use compact unresolved rows for occurrence/path evidence and a separate sorted diagnostics block for code/message evidence, making both headings explicit.
- [Color sequences complicate width calculations] → Apply ANSI only after each already-formatted status token is complete, never while padding columns.

## Migration Plan

1. Add the report package and replace transitional/legacy CLI formatters without changing application models.
2. Add injectable terminal/environment policy and route every command renderer through shared options.
3. Add golden fixtures and focused mutation, ordering, color, error, and outcome tests.
4. Run all quality gates and archive the capability; rollback is a branch revert because there is no persisted-data or config migration.
