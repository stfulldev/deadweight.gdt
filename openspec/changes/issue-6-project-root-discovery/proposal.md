## Why

Issue [#6](https://github.com/stfulldev/deadweight.gdt/issues/6) is the next unblocked MVP path-layer slice. Filesystem and `res://` inputs cannot be resolved safely until the analyzer can deterministically identify the nearest valid Godot project root without relying on Godot or machine-specific state.

Planning and implementation progress are tracked in linked Draft PR [#31](https://github.com/stfulldev/deadweight.gdt/pull/31).

## What Changes

- Add an isolated project-root discovery capability based on a regular `project.godot` marker.
- Discover from the root scene directory for absolute or cwd-relative filesystem inputs, and from the supplied current working directory for `res://` inputs.
- Give an explicit project directory or `project.godot` file precedence over automatic discovery and validate it before returning a root.
- Validate filesystem root-scene existence, regular-file status, and the case-sensitive `.tscn` extension before automatic discovery.
- Return typed fatal errors that distinguish invalid scene input, invalid explicit project, inaccessible filesystem state, and missing project root while retaining actionable paths and underlying causes.
- Keep discovery deterministic and testable through explicit cwd and filesystem-stat seams; preserve standalone operation and centralized CLI fatal-exit handling.

Goals:

- Provide the stable root-discovery contract required by path resolution, configuration discovery, and application orchestration.
- Make nearest-parent behavior and all failure modes executable through platform-neutral table-driven tests.
- Keep filesystem discovery independent of TSCN parsing, resource resolution, reporting, and Cobra handlers.

Non-goals:

- Resolve a scene into canonical/cache/display paths, enforce project containment, or evaluate symlinks; those behaviors belong to issue #7.
- Resolve `ExtResource.path`, configuration files, or nested scene dependencies.
- Wire the final global `--project` flag or replace analyzer placeholder commands; application wiring belongs to issue #20.
- Read or interpret `project.godot` contents or launch Godot.

Compatibility impact:

- This is a new internal capability; existing CLI commands, parser APIs, metric catalogs, presets, and output remain unchanged.
- The finder accepts an explicit project value shaped like the future `--project` argument, so later CLI wiring does not need to redefine validation semantics.
- The shipped binary retains only Go and existing Cobra runtime dependencies and performs no network or Godot lookup.

Affected MVP acceptance criteria:

- Section 13.1: filesystem, `res://`, and explicit-project discovery rules.
- Section 29.2: nearest-root, explicit directory/file, relative input, cwd, and failure tests.
- Section 30 criteria 2–3: absolute/relative/`res://` entry semantics and nearest `project.godot` selection.
- Section 31, Milestone 2 and Section 32, Step 3: isolated project/path layer foundations.

## Capabilities

### New Capabilities

- `project-root-discovery`: Defines deterministic Godot project-root discovery, explicit-project validation, filesystem-scene preconditions, and typed actionable failures.

### Modified Capabilities

- None.

## Impact

- New code: `internal/project` root model, finder, typed errors, and unit tests.
- Verification targets: existing CLI fatal error mapping plus repository build, test, race, vet, lint, and cross-platform CI gates.
- No new external dependency, config/schema change, report change, Godot integration, network behavior, or persistent state.
