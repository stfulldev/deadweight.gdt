## Context

See `proposal.md` for motivation. Recursive analysis currently retains enough typed evidence to classify the whole report and individual contribution rows, but `RecursiveResult` and `SceneContribution` each expose only one reliability value. Text rendering applies that aggregate value to metric formatting, and schema-v1 JSON metric objects contain no confidence metadata. The implementation must keep metric arithmetic, exit behavior, portability, and the standalone Go runtime unchanged.

## Goals / Non-Goals

**Goals:**

- Make the analysis domain the single source of truth for complete, validated per-metric confidence.
- Preserve existing aggregate reliability fields as derived compatibility summaries.
- Keep reason ordering, projections, and report bytes deterministic.
- Make mixed uncertainty visible without flooding uniform text reports.

**Non-Goals:**

- Recalculate metric values, reinterpret budget policy, or parse ordinary external resources deeply.
- Infer probabilities or attach confidence to tree edges, diagnostics, coverage, or budget limits.
- Introduce a new report schema version or runtime dependency.

## Decisions

### Use one fixed confidence value per frozen metric

The analysis package will own a `Confidence` value containing reliability and reason codes plus a fixed `MetricConfidence` aggregate with one field per frozen metric. Fixed fields make omission impossible after construction, align validation with the frozen eight-metric contract, and avoid accepting duplicate metric IDs in domain state. Accessors will expose canonical metric iteration to report projections. A dynamic slice was rejected because every call site would need repeated completeness, uniqueness, and ordering checks.

### Use stable semantic reason codes and canonical set operations

Reason codes will distinguish unresolved scene instances, imported scenes, unsupported scenes, subresource scenes, placeholders, unavailable scenes, unavailable resources, unsupported resource paths, inherited scenes, and unsupported parents. Constructors will normalize reasons into a fixed canonical order and remove duplicates; validation will reject non-canonical externally assembled state. This preserves concrete evidence without leaking diagnostic prose or filesystem paths. Reusing only diagnostic codes was rejected because several resolution paths share a diagnostic family while affecting metrics differently.

### Derive confidence from an explicit evidence impact matrix

Completeness finalization will start all root metrics exact, then merge evidence impacts with `approximate > lower_bound > exact` precedence. Unexpandable scene closure can hide any frozen metric and therefore affects all eight. An unavailable ordinary resource affects only `external_resources`. Unsupported parent composition affects only `tree_depth`. Inheritance affects all metrics that the current approximation cannot prove exact. Each contribution is initialized from its own kind and depth evidence, then merged during compaction and inherited propagation. Unknown values remain unavailable; confidence never manufactures a zero.

An analysis-wide blanket classification was rejected because it is the behavior issue #54 replaces. Optimistically declaring unresolved resources exact was rejected because the existing completeness contract intentionally treats their resource identity as unavailable evidence.

### Derive compatibility summaries from metric confidence

Report-wide and row-wide reliability will be recomputed as the conservative maximum of their metric entries and validated against the stored compatibility field. Status remains complete exactly when every metric is exact. Existing budget evaluation continues to consume the report-wide summary, preserving `fail_on_partial` semantics and exit priority.

### Project additive JSON fields within schema version one

Root and contribution metric objects will gain a required `confidence` object in producer output and the committed schema. The object contains required reliability plus an owned reason array. Existing fields retain their names and meanings, so tolerant version-one consumers remain compatible. Checked-in goldens and schema validation tests will cover inspect, check, and tree kinds.

### Keep text concise for mixed confidence

Metric values and check actuals will use their own confidence marker. A shared formatter will append a deterministic `Metric confidence` section only if an entry differs from the report-wide summary, listing differing metrics in frozen order with normalized reasons. Uniform exact, lower-bound, and approximate reports retain their established structure apart from metric-correct markers. Selected top-contribution values use the selected entry's confidence while the row summary remains visible.

## Risks / Trade-offs

- [Adding required nested JSON metadata can expose strict consumer assumptions] → Keep schema version and all existing fields stable, document the additive contract, and validate every emitted kind against the checked-in schema.
- [Impact rules can accidentally overstate exactness] → Encode the matrix centrally, default new partial evidence to conservative handling, and test every frozen metric across ordinary-resource, parent, scene-closure, and inheritance cases.
- [Owned reason slices can alias cached summaries] → Deep-clone metric confidence in recursive results and contribution collections and add mutation tests.
- [Text qualification can become noisy] → Render it only for classifications that differ from the aggregate summary and keep canonical one-line entries.

## Migration Plan

1. Introduce and test the analysis-domain confidence model and impact matrix while retaining aggregate fields.
2. Wire recursive and contribution construction, validation, compaction, scaling, and cloning.
3. Add JSON projection/schema fields and metric-aware text formatting.
4. Regenerate deterministic fixtures and document the compatible schema-v1 extension.
5. Run strict OpenSpec validation and the complete build, unit, race, vet, lint, schema, and CLI golden gates before archive and merge.

Rollback is a normal PR revert: existing metric values and reliability fields remain intact, and no data migration or external service state is involved.
