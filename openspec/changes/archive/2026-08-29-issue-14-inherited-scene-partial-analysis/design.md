## Context

See `proposal.md` for motivation. Local summaries already separate inherited roots, typeless override stubs, explicit typed nodes, and mounts; the parser already records editable syntax. Graph discovery already traverses inheritance edges, canonicalizes dependencies, caches parsed documents, and rejects cycles. Recursive expansion currently stops at an inherited summary and substitutes one known root at the parent, so resolved base metrics and root-inherited inputs cannot be analyzed.

## Goals / Non-Goals

**Goals:**

- Build an inherited scene's reusable one-occurrence summary from one base summary plus supported local additions.
- Preserve approximation evidence for root, nested, repeated, transitive, missing, imported, unreadable, and editable/override cases.
- Reuse the existing graph validation, invocation caches, checked arithmetic, completeness finalizer, and deterministic identity unions.

**Non-Goals:**

- Applying override properties, removals, type changes, reorder/owner semantics, or editable-child mutations.
- Custom script inheritance, imported-scene contents, runtime Godot, CLI/report output, config, budgets, or presets.

## Decisions

### Expand inheritance inside the canonical one-occurrence summary

Replace the inherited-scene sentinel error with an inheritance step inside `expandScene`. The step resolves the root base reference, recursively obtains a cached base summary when supported, applies it once without a mount occurrence, then continues through the derived scene's ordinary mount loop. Root and nested inherited inputs therefore share one implementation, and repeated derived mounts multiply the completed derived summary normally. Keeping inheritance at the parent would duplicate logic and cannot support an inherited root input.

### Add a dedicated base-application operation

Add a transactional summary-builder operation that checked-adds the base's nodes, nested scene instances, meshes, lights, and shadows; takes the maximum known tree depth; carries base coverage/evidence; unions resources/dependencies; and deliberately does not add the `1 + child.scene_instances` mount formula. Local summary metrics already exclude the inherited root and override stubs, so base plus explicit local additions has no duplicate root.

### Represent unsupported bases as inherited evidence, not scene-instance evidence

When the base cannot be expanded, add one known node for the inherited root and keep the inheritance edge outside resolved/unresolved scene-instance coverage. Extend inheritance evidence with base classification/reason and override/editable flags. This matches the frozen definition that inheritance is not a scene instance and lets completeness count inherited occurrences independently. Remove the implementation-only constraint that inherited coverage must be a subset of unresolved scene-instance coverage.

### Retain editable syntax in local summaries

Copy the parser's `HasEditable` feature into the local summary and then into the inheritance record. Override presence is derived from the already owned `OverrideStubs` slice. No parser changes are required, and cached local summaries continue to own all evidence.

### Let graph validation remain authoritative

Graph discovery still runs before occurrence expansion. Resolved inheritance cycles and malformed supported text bases therefore fail before any metric result is published. Expansion reuses graph-populated cached documents; opener failures are classified as unsupported-base approximation, while typed parser failures remain fatal.

### Keep `SB1003` as the inheritance warning

Completeness groups inheritance records by declaring scene and base target, multiplies their occurrence counts through cached summary application, and emits `SB1003`. Base classification/reason remains structured evidence for tests and later rendering; no new diagnostic code is introduced in this slice.

## Risks / Trade-offs

- [Risk] Base/local depth paths may be changed by unsupported overrides. → Publish the known maximum only under `approximate` reliability and never claim an exact merged tree.
- [Risk] Missing-base root counting could be mistaken for resolved base aggregation. → Retain classification/reason and exactly one fallback root with no inner metrics.
- [Risk] Repeated or transitive inheritance could double-apply a base. → Cache one-occurrence summaries by canonical scene and use a separate base operation without mount addition.
- [Risk] Completeness coverage validation currently assumes inherited occurrences are unresolved instances. → Remove that non-spec subset constraint and test root/repeated inherited coverage explicitly.

## Migration Plan

1. Retain editable/base evidence and relax the internal coverage subset constraint.
2. Implement transactional base application and replace the inherited sentinel path.
3. Add root, nested, repeated, transitive, unsupported-base, override/editable, and cycle fixtures.
4. Revert the focused feature/test commits to roll back; no stored data or external format migration exists.
