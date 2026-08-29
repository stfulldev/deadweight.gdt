## Why

Issue #7 and linked Draft PR #32 complete the project/path milestone after root discovery by defining one secure conversion boundary for root-scene inputs and declared resources. Centralizing canonical I/O paths, deterministic `res://` display paths, and unresolved classifications prevents traversal and symlink escapes while giving later graph and completeness layers stable evidence instead of ambiguous strings.

## What Changes

- Add a resolver that consumes the canonical project root established by project discovery and resolves root-scene inputs from absolute, cwd-relative, and `res://` values.
- Resolve resource references from `res://`, declaring-scene-relative, parent-relative, and host-absolute values without opening targets outside the canonical project root.
- Canonicalize existing paths, enforce segment-aware containment, and reject lexical or symlink escapes without case-insensitive recovery.
- Preserve canonical absolute paths for I/O/cache identity, forward-slash `res://` paths for display, and original raw values for diagnostics.
- Classify empty, `uid://`, `user://`, missing, unsupported, outside-project, and inaccessible targets as stable typed unresolved results suitable for later diagnostics and completeness tracking.
- Keep resolution standalone, deterministic, and testable without Godot, network access, parser coupling, or process output.

Goals:

- Provide the stable `project.Resolver` and `ResolvedPath` contracts required by scene-graph traversal and resource accounting.
- Guarantee that paths returned for downstream I/O remain inside the canonical project root after lexical cleaning and symlink evaluation.
- Make the complete project-path matrix executable through platform-neutral unit tests and real temporary filesystem fixtures.

Non-goals:

- Discover the project root; issue #6 already provides that capability.
- Open or parse TSCN/resource contents, build the dependency graph, aggregate metrics, or decide final analysis reliability.
- Resolve `uid://` through Godot metadata, support `user://`, run the Godot import pipeline, or recover path casing.
- Wire root-scene and resource resolution into placeholder CLI analyzer commands; orchestration belongs to later issues.

Compatibility impact:

- This is a new internal capability; existing CLI output, commands, parser APIs, project discovery, presets, and diagnostic codes remain unchanged.
- Host-specific path syntax is accepted through Go filesystem semantics, while stored display paths are normalized to Godot-style forward slashes.
- The shipped binary gains no external dependency and retains standalone operation without Godot or network access.

Affected MVP acceptance criteria:

- Section 13.2–13.3: resolver API, canonical/cache/display/original values, resource resolution, containment, and symlink policy.
- Section 14.1 and 19.4: explicit unresolved evidence needed for later missing/imported/unsupported classification and partial analysis.
- Section 29.2: relative, parent-relative, absolute, scheme, escape, symlink, display, and cross-platform path tests.
- Section 30 criteria 2 and 4: root scene input forms and declaring-scene-relative `ExtResource.path` resolution.
- Section 31, Milestone 2 and Section 35: centralized cross-platform project/path resolution with the full path matrix green.
- Section 36: preserve the standalone no-Godot boundary and all frozen MVP semantics.

## Capabilities

### New Capabilities

- `secure-path-resolution`: Defines canonical root-scene and resource resolution, project containment, symlink-escape prevention, deterministic display paths, and typed unresolved classifications.

### Modified Capabilities

- None.

## Impact

- New implementation is expected to extend `internal/project` with resolver models, resolution reasons, filesystem seams, and focused unit tests.
- Later graph, diagnostic, completeness, and CLI layers can consume the resolver contract without duplicating path policy; those integrations remain outside this change.
- Verification covers build, targeted/full tests, race, vet, pinned lint, OpenSpec strict validation, and Linux/macOS/Windows CI.
- No new config/schema, report format, persistent state, Godot integration, network behavior, or third-party runtime dependency.
