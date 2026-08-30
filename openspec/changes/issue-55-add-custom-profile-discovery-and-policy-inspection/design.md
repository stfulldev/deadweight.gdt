## Context

See `proposal.md` for motivation. The existing version-one decoder owns strict JSON shape validation, while `internal/policy` owns dynamic graph validation and the four-layer budget merge. Application flows currently discover projects only through a scene request, text presets have no JSON mode, and report schema version one uses a discriminated document envelope. The new commands must cross these boundaries without creating a second resolver or making the shipped binary depend on OpenSpec, Node.js, Godot, or network access.

## Goals / Non-Goals

**Goals:**

- Add a scene-free project-context path that reuses existing explicit-project validation and ancestor discovery.
- Extend policy resolution with owned explanation data produced during the actual merge operations.
- Keep application models presentation-independent while supporting deterministic text and JSON renderers.
- Preserve exact check-policy semantics and schema-version-one compatibility.

**Non-Goals:**

- Expose mutable resolver internals or configuration editing APIs.
- Make profile listing project-independent or blend it with the built-in preset catalog.
- Persist explanation caches or optimize for unusually large configuration graphs beyond the existing bounded depth.

## Decisions

### Add a dedicated project-context finder entry point

`project.Finder` will gain a scene-free method that normalizes and validates the working directory, validates an explicit project through the existing path, and otherwise invokes the existing nearest-marker traversal from the working directory. This avoids fake scene inputs and preserves the typed discovery errors. The application dependencies will expose this operation independently so command-flow tests can prove that profiles never invoke scene resolution.

Alternative considered: call scene discovery with `res://` as a sentinel. Rejected because an invented scene obscures intent, couples project-only commands to scene semantics, and makes invocation assertions misleading.

### Produce provenance inside the shared graph resolver

The internal resolved-profile representation will carry effective values, one source per metadata and budget field, and a root-to-child layer chain. Root custom defaults and built-in preset conversion initialize sources; the existing field-by-field custom merge replaces value and source together; the existing top-level budget merge does the same for project values. Ordinary `Resolve` projects the effective values exactly as today, while new list/explain operations project owned summaries and explanations from the same resolved graph.

Alternative considered: re-run merge logic after ordinary resolution to infer sources. Rejected because inference cannot distinguish equal overrides from inheritance and would duplicate the precedence contract.

### Retain explicit `fail_on_partial` presence in the decoded model

The configuration model will preserve whether `fail_on_partial` was present while keeping its existing effective boolean. A read-only query exposes that evidence to policy inspection, allowing absent false to report the default layer and explicit false to report the project layer. The JSON schema and accepted configuration syntax do not change.

Alternative considered: label all false values as project-sourced whenever a configuration exists. Rejected because it reports false provenance for omitted fields.

### Use typed explanation and source models

Policy will define stable layer kinds for `default`, `preset`, `profile`, and `project`; optional IDs are carried only by preset/profile layers. Metadata sources are statically typed fields, budget sources mirror the eight optional limits, and the chain contains only inheritance nodes rather than the implicit default layer. Application results clone all collections and pointer-backed values before returning them.

Alternative considered: a generic map of field names to strings. Rejected because it weakens completeness checks, encourages non-canonical iteration, and makes schema/report drift easier.

### Add two discriminated JSON kinds

The existing version-one envelope will add optional `profiles` and `profile` payloads and schema branches for kinds `profiles` and `profile`. Lists use arrays for canonical order. Effective budgets use an ordered array of `{metric, limit, source}` records, while metadata uses a fixed object whose value/source pairs retain empty and zero values. Configuration paths are normalized relative to the project root using the same portable-path discipline as scene reports.

Alternative considered: encode both commands under one kind with an action field. Rejected because separate payloads yield stricter schema exclusion, clearer consumers, and match existing command-specific report kinds.

### Keep command format selection at presentation boundaries

The application exposes project/config request models and list/show results only. Cobra validates `--format` before calling the service, then selects the text or JSON renderer. Fatal errors use the existing selected-format error presentation and centralized exit code `2`.

Alternative considered: pass format into application methods. Rejected because it would couple discovery and policy logic to encoding and streams.

## Risks / Trade-offs

- [Adding provenance fields can drift from merge fields] → Mirror metadata and budget structures with exhaustive tests that exercise every field and compare explained effective output to ordinary `Resolve`.
- [Configuration object order is not meaningful after decoding] → Define canonical ascending profile-ID order rather than claiming source-text order.
- [Schema-version-one grows additional branches] → Keep all established branches unchanged and validate both new documents and earlier golden reports against the evolved schema.
- [Project-relative configuration paths can be unavailable for explicit files outside the project] → Reuse the existing portable configuration selection representation: project-relative paths when contained, otherwise the cleaned user selection without canonical checkout expansion.

## Migration Plan

1. Add internal evidence and project-context APIs without changing existing callers.
2. Add application, CLI, report, and schema support with focused tests.
3. Run OpenSpec validation and all Go quality gates, then archive the change before marking PR #64 ready.
4. Roll back by reverting the additive commands, report kinds, and internal evidence fields; existing config and report consumers require no data migration.
