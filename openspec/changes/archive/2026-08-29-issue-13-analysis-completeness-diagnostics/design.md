## Context

See `proposal.md` for motivation. The recursive analyzer already retains occurrence coverage, unresolved mount evidence, inherited-target evidence, unsupported parent findings, closure-wide resource identities, and a checked unique parse-cache count. It does not yet publish a root completeness contract or convert that evidence into validated diagnostics. The diagnostic package owns stable code/severity validation, while the analysis package owns semantic classification and occurrence arithmetic.

## Goals / Non-Goals

**Goals:**

- Finalize one owned root-level status, reliability, coverage, and diagnostic collection after graph, metric, and parse-cache finalization succeeds.
- Preserve enough resolution evidence on unresolved resource identities to distinguish missing, filesystem, root-escape, UID, user-data, empty, and unsupported paths.
- Group repeated evidence transactionally with checked arithmetic and deterministic ordering.
- Keep all fatal paths zero-result and make normal resolved non-scene resources neutral to completeness.

**Non-Goals:**

- Aggregating inherited base metrics or applying overrides; issue #14 owns that behavior.
- Rendering diagnostics, coverage, or reliability in CLI/report output; later tracker issues own presentation.
- Configuration, budgets, `fail_on_partial`, presets, dependency changes, runtime Godot inspection, or custom script class inference.

## Decisions

### Finalize completeness once at the root boundary

Add a pure completeness finalizer in `internal/analysis` and call it after metrics and the checked parsed-file count are available but before constructing `RecursiveResult`. It receives the completed summary and graph evidence, derives the public coverage, groups warnings, validates all values, and returns an all-or-error value. This avoids storing root-only status, diagnostics, or parsed-file counts inside reusable one-occurrence summaries. The rejected alternative was mutating cached summaries, which would multiply unique coverage and duplicate diagnostics when children are reused.

### Extend the analyzer result without replacing retained lower-level evidence

Add stable analysis status and reliability types plus a public coverage value and diagnostics to `RecursiveResult`. Keep `ExpandedSummary.Coverage`, unresolved entries, inherited targets, and parent findings as the recursive aggregation substrate because their multiplicity is already checked and tested. The public coverage copies resolved/unresolved totals, the cache count, and a checked sum of inherited occurrences. This keeps issue #13 focused and leaves issue #14 free to change inherited aggregation without redesigning the root contract.

### Preserve resolution reasons on unresolved resource identities

Extend unresolved `ResourceIdentity` values with `project.ResolutionReason`; resolved identities remain canonical-only for uniqueness. The unresolved uniqueness contract remains the tuple `(declaring scene, resource ID, raw path)` because one declaration has one deterministic resolution result. Carrying the reason alongside that tuple lets the completeness layer classify ordinary missing resources without resolving paths a second time. The rejected alternative was rerunning filesystem resolution during finalization, which would duplicate effects and could observe different state.

### Build diagnostics from semantic evidence with explicit mapping

Use one internal grouping key containing code, file, source location, resource/target identity, classification, resolution reason, and stable message. Map imported, inherited, placeholder, UID/user, unavailable-resource, generic unresolved-scene, and unsupported-parent evidence to the codes defined by the capability spec; add warning code `SB1008` to the diagnostic catalog. Identical keys use checked occurrence addition, then validate and sort the resulting `diagnostic.Diagnostic` values by the frozen order. Resource declarations already represented by a mount warning are deduplicated by their declaring scene/resource identity so one underlying cause is not reported twice.

### Derive reliability from causes, not rendered text

The finalizer tracks whether any lower-bound cause and any approximation cause exist while it processes structured evidence. Inheritance sets approximation; unresolved mounts, unresolved declarations, and parent findings set lower-bound. Approximation has explicit precedence. A result with no cause is complete/exact. The rejected alternative was inferring status from the final diagnostic messages, which would couple correctness to presentation strings.

### Treat validation failure as fatal and transactional

Negative occurrences, invalid resolution reasons, unknown diagnostic mappings, grouping overflow, invalid diagnostics, or invalid coverage return an error and cause `Analyze` to return `RecursiveResult{}`. The finalizer operates on copies and allocates its own maps/slices, so callers cannot mutate cached analyzer state through the returned diagnostics.

## Risks / Trade-offs

- [Risk] A resource declaration used by both graph and expansion evidence could create duplicate warnings. → Track covered declaring-scene/resource-ID pairs and let mount/inheritance evidence own occurrence multiplicity before adding declaration-only warnings.
- [Risk] Grouping source lines too strictly can prevent repeated mounts from collapsing. → Group scene-target evidence by stable semantic target/reason and declaring display path, retaining the first deterministic source position rather than mount name.
- [Risk] Adding `SB1008` expands the diagnostic catalog beyond the MVP's listed minimum. → Keep it warning-only, document it in the capability, and test exact catalog order and validation.
- [Risk] Issue #14 will change inherited occurrence metrics. → Base status/reliability only on retained inherited evidence and avoid assumptions about future effective-tree aggregation.

## Migration Plan

1. Add domain types and resolution-reason retention without changing public CLI behavior.
2. Add the pure finalizer and wire it transactionally into `Analyze`.
3. Update exact result tests and add completeness, resource, grouping, ownership, and fatal-matrix fixtures.
4. Roll back by reverting the focused feature and test commits; no persisted data or external format migration is involved.
