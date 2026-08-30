## Why

MVP 0.2 users can declare reusable custom profiles, but they cannot discover those declarations or explain the policy that inheritance and project overrides actually produce without running a scene check. GitHub issue #55 and its linked Draft PR #64 add deterministic, offline inspection so configuration authors can verify both effective values and their provenance before analysis.

## What Changes

- Add project-aware `profiles` and `profiles show <id>` commands for version-one custom profiles while keeping the built-in `presets` namespace distinct.
- Resolve and display effective metadata, all effective budgets, `fail_on_partial`, the complete parent chain, and the source layer of every resolved value through the existing strict configuration and policy-resolution rules.
- Add deterministic text output and compatible schema-version-one JSON document kinds for custom-profile listing and inspection.
- Add project-root discovery from the working directory for commands that need project context but do not consume a scene.
- Reject missing projects, absent or invalid configuration, unknown profiles, invalid parents, collisions, cycles, and excessive inheritance depth with actionable deterministic errors.
- Preserve version-one configuration and preset compatibility; this change does not alter existing `check`, `presets`, configuration-schema, selector, or merge semantics.

### Goals

- Make declared custom profiles discoverable in canonical profile-ID order.
- Make every inherited, defaulted, and overridden effective value auditable.
- Guarantee that inspected effective values match `check --profile <id>` for the same project configuration.
- Keep text and JSON output deterministic and checkout-independent.

### Non-goals

- Editing or interactively creating configuration.
- Automatic hardware detection or platform recommendation.
- Changing built-in preset contents, configuration schema version, or policy precedence.

### Compatibility and acceptance impact

The change is additive. Existing version-one configuration remains valid, existing built-in preset commands and report kinds retain their meanings, and consumers may ignore the new JSON kinds. It satisfies the issue #55 acceptance criteria for deterministic ordering, exact effective-policy parity, complete provenance, actionable missing-context errors, preset/config-v1 compatibility, automated coverage, and the repository quality gates.

## Capabilities

### New Capabilities

- `custom-profile-inspection`: Discover custom profiles and inspect effective values, inheritance chains, and value provenance.

### Modified Capabilities

- `application-command-flows`: Add project-context custom-profile list/show command flows and syntax.
- `project-root-discovery`: Support deterministic discovery for project-context commands without a scene input.
- `effective-policy-resolution`: Publish explained profile resolution from the same merge implementation used by checks.
- `deterministic-console-reports`: Add canonical custom-profile list/show text presentations.
- `versioned-json-reports`: Add compatible schema-version-one custom-profile list/show document kinds.

## Impact

- Affected packages: `internal/project`, `internal/config`, `internal/policy`, application orchestration, `internal/report`, `internal/cli`, and the composition root.
- Affected public CLI surface: additive `profiles`, `profiles show <id>`, and their `--format text|json` options; existing global `--project` and `--config` flags apply.
- Affected schemas: additive schema-version-one report kinds and payload definitions; `schema/deadweight.gdt.schema.json` is unchanged.
- Dependencies and runtime: no new runtime dependency, network access, Godot process, Node.js, or OpenSpec requirement in the shipped binary.
