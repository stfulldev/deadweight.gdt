## Context

See `proposal.md` for motivation and `specs/effective-policy-resolution/spec.md` for the behavioral contract. Issue #16 now provides owned, statically validated config declarations, issue #15 provides an owned embedded preset catalog, and `budget.Limits` preserves absent versus explicit-zero fields. The missing layer is a pure resolver between those inputs and the later checker/application flow.

The resolver must report profile/config failures as deterministic `SB2003` errors, must not read files or scene data, and must keep Cobra, reporting, OpenSpec, Godot, and network concerns out of its runtime dependency graph.

## Goals / Non-Goals

**Goals:**

- Introduce one small policy package with an explicit input/output boundary and no side effects.
- Validate every custom declaration before selecting and merging a policy.
- Keep optional-budget and metadata ownership intact across every merge layer.
- Make graph traversal, error selection, and cycle/depth evidence reproducible regardless of Go map iteration.
- Leave a composition-friendly API for issue #20 without importing CLI framework types.

**Non-Goals:**

- Re-decoding or mutating configuration, changing the embedded preset catalog, evaluating scene metrics, or resolving partial-analysis exit policy.
- Producing console strings beyond typed deterministic error messages.
- Adding generic graph or schema libraries for a three-parent-domain, single-inheritance problem.

## Decisions

### Add a pure `internal/policy` package

The package will accept a config source string, an owned `config.Config`, a CLI selector value, and ordered raw CLI budget values. It will load the existing validated embedded preset catalog and return either an owned `Effective` policy or an error. The effective value records selected kind (`none`, `preset`, or `profile`), selected ID, merged metadata, and `budget.Limits`.

This keeps JSON shape/static validation in `internal/config`, built-in data in `internal/preset`, and comparisons in `internal/budget`. The package will return `*config.Error` with `ReasonValidation`, `SB2003`, source/field evidence, and deterministic detail for all policy failures.

Alternatives considered:

- Put resolution in `internal/config`: rejected because profile/preset/CLI precedence is a policy concern after config decoding and requires the preset catalog.
- Put resolution in Cobra handlers: rejected because it would couple domain rules to command parsing and make issue #20 difficult to test.
- Add a configurable resolver object: unnecessary for the immutable embedded v0.1 catalog; a function boundary is smaller and the catalog already returns defensive copies.

### Validate the whole profile map before resolving selection

The resolver will copy its inputs, obtain built-in IDs, sort custom IDs, reject collisions, and validate every parent reference. It will then run one deterministic tri-color traversal over all sorted custom IDs, maintaining a stack and stack-index map. A gray parent produces the exact closed cycle slice; completed profiles are memoized.

Depth is the number of custom-profile nodes in the active path. Depth 32 is accepted; attempting to enter a 33rd custom node fails. Reaching a built-in ends the custom path without incrementing depth. This definition makes a root custom profile depth 1 and matches the bounded configuration-inheritance intent.

Alternatives considered:

- Resolve only the selected branch: rejected because malformed unselected profiles would make a strict config conditionally valid depending on CLI choice.
- Recurse directly over map iteration: rejected because the first reported invalid profile and cycle anchor would be nondeterministic.
- Use a general graph dependency: rejected because single inheritance needs only linear parent edges and a small local traversal.

### Represent metadata as a complete effective value

An internal metadata value will hold name, description, platform, renderer, target FPS, quality, status, and stability. Built-ins seed every field. A root custom profile starts from the frozen custom defaults, with empty name/description/stability and status `custom`; declared fields then replace those defaults. A child starts from its resolved parent and replaces only fields represented by non-nil config pointers. The selected kind/ID remains the requested base even when its metadata derives from an ancestor.

Alternatives considered:

- Keep effective metadata optional: rejected because downstream reporting would have to repeat inheritance/default rules.
- Reset all lifecycle fields whenever any custom child is present: rejected because the frozen contract says omitted metadata inherits and specifically defines custom defaults only for profiles without `extends`.

### Merge optional budgets by explicit metric order

A local merge helper will clone the lower layer, visit the eight frozen metric IDs in canonical order, and replace only configured higher-layer fields, including zero. Profile traversal memoizes fully merged budgets. Resolution then overlays top-level config budgets and parsed CLI budgets.

CLI values will be parsed left-to-right as exactly one `metric=limit` pair. The parser will reject empty pieces, whitespace changes, unknown metric IDs, signs/negative values, non-decimal content, and signed-64-bit overflow. Setting the same metric again replaces its pointer, so the final occurrence wins.

Alternatives considered:

- Convert limits to a generic map: rejected because it would weaken the existing fixed-eight typed contract and make unknown metrics easier to leak through.
- Sort CLI overrides: rejected because duplicate last-wins semantics require preserving caller order.

### Publish only a complete owned result

Resolution will build into local values and return the zero `Effective` value on every error. Before success it will require `Budgets.Count() > 0`; no-base policies remain valid when project or CLI overrides supply at least one limit. Every returned `budget.Limits` value will be cloned, and metadata consists only of value strings/integers, so callers cannot mutate config, catalog, memoized, or later results.

Alternatives considered:

- Return a partial policy alongside an error: rejected because issue #20 needs a clear fatal boundary and the spec forbids publishing a usable policy after semantic failure.
- Enforce non-empty budgets in config decoding: rejected because built-in/profile selection and CLI overrides do not exist at that phase.

## Risks / Trade-offs

- [Risk] A recursive DFS could approach Go stack limits if hostile configs contain very large maps. → Mitigation: traversal stops at 33 active custom nodes, and completed nodes are memoized; the maximum recursion depth is bounded.
- [Risk] Returning `config.Error` for CLI override validation blurs config and command provenance. → Mitigation: use explicit `cli.budgets[index]` fields and preserve the source separately; issue #20 remains responsible for mapping domain errors to exit 2.
- [Risk] Embedded catalog load failure is not caused by user config. → Mitigation: propagate the catalog error without manufacturing a partial policy; embedded-data validation already makes this an internal build defect.
- [Risk] Full-map validation is stricter than selected-branch-only evaluation. → Mitigation: this is intentional for strict version-one configuration and prevents dormant invalid declarations from changing validity by CLI selection.

## Migration Plan

1. Add the isolated resolver and production model without changing any existing command path.
2. Add exhaustive unit coverage against the real embedded catalog and constructed config declarations.
3. Run all Go, race, vet, lint, and strict OpenSpec gates, then archive the capability spec.
4. Issue #20 can adopt the resolver as an additive composition step; rollback consists of reverting this isolated package and its archived spec before that wiring occurs.
