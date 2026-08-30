## Context

See `proposal.md` for motivation. Schema-v1 reports are portable producer output, but their Go projection types are intentionally private to `internal/report`, whose presenters already depend on application models. The new flow must decode reports without creating an application↔presentation import cycle, accept older compatible v1 documents that predate per-metric confidence, and preserve the centralized `0/1/2/3` outcome taxonomy.

## Goals / Non-Goals

**Goals:**

- Keep baseline decoding, semantic validation, comparison, assessment, and enforcement independent of console/JSON rendering.
- Make two file reads injectable and keep diff independent of project/scene effects.
- Represent only semantic changes with deterministic portable identities and owned collections.
- Reuse schema version one and existing error/presentation framing without weakening older v1 compatibility.

**Non-Goals:**

- Reconstruct analysis domain graphs or contribution arithmetic from a report.
- Compare presentation metadata, raw JSON bytes, configuration provenance, tool versions, or arbitrary unknown fields.
- Add configuration-file policy, percentages, tolerance bands, remote storage, or source-control operations.

## Decisions

### Add an independent reportdiff domain package

`internal/reportdiff` will own minimal schema-v1 input models, streaming-safe single-document decode, semantic validation, comparison models, confidence assessment, and enforcement. The application package will inject `ReadFile`, decode both paths, compare them, and return an owned result. The report package will depend on the result only for presentation. Reusing private JSON projection structs was rejected because exporting them would couple input compatibility to producer implementation and importing report from application would create a cycle.

### Decode known semantics and ignore unknown version-one fields

The reader will use `encoding/json` with one-document/trailing-content checks, retain exact `int64` types, and intentionally not disallow unknown fields. It will explicitly validate schema version, tool name, supported kind, portable root identity, canonical metric order and uniqueness, reliability/confidence, coverage, diagnostics, dependency identities, and check evaluation. This preserves the version-one consumer rule while rejecting incomplete or incompatible evidence. Loading the repository JSON Schema at runtime was rejected because installed binaries cannot depend on a checkout-relative schema file.

### Compare normalized typed collections

Decode normalizes metrics into frozen order, diagnostics by a stable key excluding occurrence count, dependency identities as a unique sorted set, and check comparisons by metric. Result collections contain only changes. Diagnostic occurrence changes are distinct from additions/removals; raw absolute deltas are signed `after - before`; tool version and JSON formatting are excluded. Maps may be used internally, but returned slices are sorted and cloned.

### Use proof-oriented confidence assessment

Metric confidence comes from each metric when present. For older v1 inputs it falls back to report reliability with source `report_summary`. An increase is proven only when an exact baseline is compared with an exact or lower-bound candidate. A decrease is proven only when an exact candidate is compared with an exact or lower-bound baseline. Approximate evidence and all other directional combinations are uncertain. This asymmetric rule uses the mathematical meaning of a lower bound and prevents a smaller partial candidate from being called an improvement.

### Keep enforcement separate from semantic change

The result always contains the same semantic diff regardless of policy. A normalized policy contains a unique canonical metric set and a reliability-degradation boolean. Proven selected increases produce `FAILED`; selected uncertain increases or enforced reliability degradation produce `INCOMPLETE`; incomplete wins over failed. With no flags or no triggers the outcome is `PASSED`. The CLI maps these existing statuses to `0`, `1`, and `3` after presentation; decoding/compatibility failures remain fatal `2`.

### Add one additive schema-v1 diff kind

Text and JSON presenters project the same owned result. JSON kind `diff` gets a dedicated payload and schema branch; established kinds remain unchanged. The input baseline paths are not serialized, avoiding host-specific leakage. Existing kind `error` handles JSON-mode failures. Checked-in text/JSON goldens and schema tests cover empty, changed, failed, incomplete, and malformed cases.

## Risks / Trade-offs

- [Hand-written semantic validation can drift from producer schema] → Validate every current golden through the reader, retain focused invalid fixtures, and keep the reader limited to fields comparison actually needs.
- [Older v1 reports lack metric confidence] → Fall back conservatively to report reliability and expose the fallback source in changes.
- [Diagnostics with changed prose may look removed plus added] → Use all stable fields except occurrences as identity; prose or source changes intentionally remain remove/add evidence.
- [Large baseline files consume memory] → Bound each input file to 16 MiB before decode, far above expected single-scene reports, and fail actionably when exceeded.
- [Exit `3` can surprise CI users] → Enforcement is opt-in, text/JSON name every incomplete trigger, and documentation includes exact shell behavior.

## Migration Plan

1. Add and test typed baseline decoding, normalization, comparison, confidence assessment, and policy evaluation.
2. Add injectable application flow and CLI validation without touching scene flows.
3. Add deterministic text/JSON presenters, schema-v1 diff kind, goldens, and end-to-end baseline fixtures.
4. Document baseline capture/update and CI enforcement, then run strict OpenSpec and complete repository gates.

Rollback is a normal PR revert. Existing reports and commands are not migrated or rewritten, and the new flow stores no external state.
