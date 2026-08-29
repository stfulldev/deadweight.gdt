## Context

See `proposal.md` for motivation and `specs/strict-configuration-v1/spec.md` for behavior. The repository has an eight-field pointer-based `budget.Limits`, validated embedded built-ins, project-root discovery, and `SB2003`, but no config package or canonical schema. Standard `encoding/json` decoding directly into pointers would incorrectly make explicit `null` indistinguishable from absence, while issue #17 needs metadata presence and unresolved inheritance references preserved.

## Goals / Non-Goals

**Goals:**

- Give config discovery, strict decoding, static validation, and typed errors one focused package boundary.
- Preserve omitted versus present-zero budget values and omitted versus present profile metadata.
- Keep the checked-in JSON Schema aligned with the frozen model without adding a runtime schema engine.
- Leave declarations in a form issue #17 can resolve without reparsing JSON.

**Non-Goals:**

- Resolving preset/profile references, detecting collisions or cycles, enforcing depth 32, merging profiles, or constructing effective policies.
- Wiring Cobra flags, loading config during `inspect` or `check`, evaluating partial policy, or rendering errors.
- Changing the existing project finder, built-in preset values, budget checker, or diagnostic taxonomy.

## Decisions

### Split discovery, decoding, and validation APIs

Add `internal/config` with a small discovery result that distinguishes absent config from a selected path, a loader that reads the selected regular file, a decoder for exactly one JSON document, and a static validator callable on constructed models. Composition keeps CLI concerns out of the package and lets tests target filesystem priority, structural errors, and semantic rules independently. A single CLI-owned loader function was rejected because it would couple issue #16 to issue #20.

### Decode through strict wire objects with raw optional fields

Use `encoding/json.Decoder` with `DisallowUnknownFields` for the top-level object and each profile/budget object, but retain optional fields as `json.RawMessage` until typed helpers decode them. The helpers reject literal `null`, enforce one exact JSON value, and attach stable dotted field paths. This preserves absence separately from zero and avoids the standard pointer behavior that accepts `null` as nil. Decoding through `map[string]any` was rejected because it loses typed ownership and makes integer/range handling less direct.

### Reuse budget limits and preserve profile metadata presence

The public config model carries `budget.Limits` for both top-level and per-profile budgets. Optional profile metadata uses pointers so issue #17 can distinguish inheritance from an explicit zero target FPS or explicit empty human-readable text. Returned maps and pointers are owned by the decoded value; no package-level cache or mutable singleton is introduced.

### Keep static validation narrower than profile resolution

Static validation owns version, selector exclusion, ID syntax, non-negative numeric domains, non-empty platform, and renderer/quality enums. Pattern-valid unknown selectors and parents are retained. A later resolver will own catalog lookup, custom collision, missing parents, cycles, depth, and merges. Combining these phases here was rejected because it would duplicate issue #17 and make the decoder require a preset catalog.

### Use one typed SB2003 error family

Define a config error with a stable reason, source path, dotted field path, detail, and optional wrapped filesystem/JSON cause. It implements the diagnostic coded/message interfaces and always returns `SB2003`. Discovery and decoding return a zero value on error. This provides actionable CLI material without deciding final exit rendering in this slice.

### Check in the frozen schema without runtime validation

Add the Draft 2020-12 schema under `schema/` using the exact field set and constraints from the MVP specification. Tests parse the schema and assert parity-critical nodes while the Go decoder matrix exercises every accepted and rejected class. Adding a third-party JSON Schema engine to the shipped dependency graph was rejected because runtime behavior is already authoritative and the product must remain a small standalone binary.

## Risks / Trade-offs

- [Risk] Schema and Go validation could drift later. → Centralize frozen constants in Go and add schema-parity tests for fields, metrics, patterns, enums, and object strictness.
- [Risk] Raw-message decoding adds code compared with direct struct decoding. → Keep helpers small, deterministic, and table-driven; the explicit layer is required to reject `null` correctly.
- [Risk] Users may expect unknown parent references to fail during decode. → Document and test the phase boundary; issue #17 performs dynamic validation before any effective policy is used.
- [Risk] Filesystem messages can vary by operating system. → Wrap causes but make the diagnostic message lead with stable reason, source, and action text rather than OS-specific wording.

## Migration Plan

1. Add the config model, error, discovery, decoder, and static validator without wiring existing commands.
2. Add the canonical schema and parity/behavior tests across the frozen rejection matrix.
3. Issue #17 consumes the declarations for dynamic resolution; issue #20 later composes discovery and loading into CLI flows.
4. Revert the focused feature and test commits to roll back; no existing stored format or runtime behavior is replaced in this slice.
