## Why

Root-level scene budgets can say that a scene is heavy, but they do not identify which nested scene occurrences produced the measured weight. Issue [#51](https://github.com/stfulldev/deadweight.gdt/issues/51), delivered through Draft PR [#60](https://github.com/stfulldev/deadweight.gdt/pull/60), adds deterministic attribution so developers and automation can find actionable contributors without weakening the static-analysis reliability contract established by #50.

## What Changes

- Publish owned, deterministic per-scene contribution records from recursive analysis, retaining portable scene identity, immediate mount context, checked occurrence multiplicity, frozen metric order, and contribution reliability.
- Define additive attribution for occurrence metrics so contribution totals reconcile exactly with the root aggregate while repeated and diamond dependencies continue to reuse cached scene work.
- Represent maximum and unique-union metrics explicitly as non-additive evidence; shared resources and dependencies are never assigned misleading additive ownership.
- Add an opt-in `inspect` top-contributors presentation selected by an explicit metric and positive limit, with deterministic ordering and tie-breaking.
- Extend successful version-one inspect and check JSON with compatible contribution evidence while preserving existing required fields and meanings.
- Cover chain, diamond, repeated-instance, shared-resource, partial, inherited, portability, ownership, and overflow behavior.
- Preserve the existing default text report, root metric values, application outcome, exit-code policy, standalone Go runtime, and Godot-free execution.

### Goals

- Make root scene weight explainable through auditable scene-level evidence.
- Keep additive attribution mathematically reconcilable and non-additive attribution honest.
- Make top-contributor and JSON output byte-stable across supported platforms and checkout locations.

### Non-goals

- Rendering the dependency tree tracked by #52.
- Adding per-metric confidence metadata tracked by #54.
- Parsing imported scenes or deep resource contents, resolving UIDs, or implementing full inherited-scene merge semantics.
- Baseline comparison, multi-root project scans, or new metric definitions.

### Compatibility and acceptance impact

- The default `inspect` text bytes remain unchanged when contribution selection is absent.
- Existing version-one JSON fields remain compatible; inspect and check contribution evidence is an additive schema extension.
- Existing root metric totals and exit codes remain authoritative and unchanged.
- MVP 0.2 gains exact reconciliation for additive occurrence metrics, deterministic top ordering, explicit non-additive semantics for shared unique evidence, and honest qualification of partial or approximate contributions.

## Capabilities

### New Capabilities

- `scene-contributions`: Defines contribution identity, occurrence attribution, additive reconciliation, non-additive evidence, reliability, deterministic ownership, and top-contributor selection.

### Modified Capabilities

- `recursive-scene-expansion`: Retains contribution records while applying cached child summaries and checked multiplicity.
- `scene-metric-aggregation`: Defines how contribution evidence relates to additive, maximum, and unique-union root metrics.
- `analysis-completeness`: Qualifies contribution rows affected by unresolved, imported, inherited, or unsupported-parent evidence.
- `application-command-flows`: Adds validated inspect-only contribution presentation selectors without changing analysis or outcome semantics.
- `deterministic-console-reports`: Adds the opt-in top-contributors text section while preserving default output.
- `versioned-json-reports`: Adds portable, ordered contribution evidence to compatible version-one inspect and check payloads.

## Impact

- Analysis models and recursive aggregation gain checked, clone-safe contribution evidence derived from the existing local-summary cache and authoritative graph.
- Inspect CLI parsing and report options gain metric/limit selection; application service interfaces remain presentation-independent.
- Text and JSON report validation/rendering, the committed v1 JSON Schema, goldens, and documentation gain contribution coverage.
- Focused analyzer, report, CLI, schema, portability, and integration tests are added; no shipped runtime dependency or network/Godot requirement is introduced.
