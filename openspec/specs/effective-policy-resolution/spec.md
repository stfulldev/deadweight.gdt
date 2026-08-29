## Purpose

Defines deterministic resolution of configured selectors, custom-profile inheritance, project overrides, and ordered CLI budgets into one owned effective check policy.

## Requirements

### Requirement: Selector precedence preserves preset and profile domains
The resolver SHALL treat a preset selector as a built-in-only reference and a profile selector as a custom-config-only reference. CLI selection MUST replace config selection as a whole, and the CLI preset and profile selectors MUST NOT coexist. With no CLI selector, the resolver SHALL use the config selector; with neither, it SHALL resolve without a base policy. An unknown or cross-domain selector MUST fail instead of falling back to the other domain.

#### Scenario: CLI selector replaces config selector
- **WHEN** config selects custom profile `shipping` and CLI selects built-in preset `mobile`
- **THEN** the base policy is the built-in `mobile` preset
- **AND** the config profile selector is not also merged

#### Scenario: No selected base
- **WHEN** neither CLI nor config selects a preset or profile
- **THEN** resolution starts with no base metadata or budgets
- **AND** project and CLI budget overrides remain eligible to form the effective policy

#### Scenario: Invalid CLI selector pair
- **WHEN** CLI supplies both a preset and a profile selector
- **THEN** resolution fails with an actionable selector-conflict error

#### Scenario: Selector domains do not fall through
- **WHEN** a preset selector names only a custom profile or a profile selector names only a built-in preset
- **THEN** resolution fails and identifies the unresolved selector

### Requirement: Every custom profile graph is semantically valid
Before publishing a policy, the resolver MUST reject any custom profile ID that collides with a built-in preset ID and any `extends` reference that identifies neither a built-in nor a declared custom profile. Dynamic validation SHALL cover the complete custom-profile map, including profiles not reachable from the selected base, and equivalent maps MUST yield errors in deterministic profile-ID order.

#### Scenario: Built-in ID collision
- **WHEN** the custom profile map declares `steam-deck`
- **THEN** resolution fails and identifies `profiles.steam-deck` as colliding with a built-in preset

#### Scenario: Missing parent in an unselected profile
- **WHEN** an unselected custom profile extends an unknown ID
- **THEN** resolution fails and identifies that profile's `extends` field and missing parent

### Requirement: Inheritance cycles and depth are bounded and actionable
Custom profiles SHALL use single inheritance. Resolution MUST reject every custom-profile cycle with the full closed chain in traversal order and MUST reject an inheritance path containing more than 32 custom-profile nodes. A built-in terminal parent SHALL not add a custom-profile level. Cycle and depth failures MUST be deterministic and identify the affected profile path.

#### Scenario: Full cycle chain
- **WHEN** `a` extends `b`, `b` extends `c`, and `c` extends `a`
- **THEN** resolution fails with the closed chain `a -> b -> c -> a`

#### Scenario: Maximum accepted depth
- **WHEN** a selected inheritance path contains exactly 32 custom-profile nodes and terminates at no parent or a built-in
- **THEN** profile resolution succeeds

#### Scenario: Depth overflow
- **WHEN** an inheritance path contains 33 custom-profile nodes
- **THEN** resolution fails with the maximum depth and affected chain identified

### Requirement: Profile metadata merges field by field
For a custom profile with a parent, each declared `name`, `description`, `platform`, `renderer`, `target_fps`, and `quality` value SHALL replace that field from the effective parent metadata, while each omitted field SHALL remain inherited. A custom profile without `extends` MUST begin with platform `custom`, renderer `unspecified`, target FPS `0`, quality `custom`, and status `custom`; its name and description SHALL remain empty unless declared. Built-in lifecycle metadata SHALL remain available when inherited and not replaced by a custom default.

#### Scenario: Child overrides selected metadata
- **WHEN** a custom profile extends `steam-deck`, overrides only name and quality, and omits the remaining metadata
- **THEN** its effective name and quality come from the child
- **AND** platform, renderer, target FPS, status, and stability remain inherited from `steam-deck`

#### Scenario: Root custom defaults
- **WHEN** a custom profile has no `extends` and declares no metadata
- **THEN** its effective platform is `custom`, renderer is `unspecified`, target FPS is `0`, quality is `custom`, and status is `custom`

### Requirement: Budgets merge in the frozen four-layer order
All eight optional budgets SHALL merge field by field from lowest to highest priority: built-in or ancestor profile, descendant profile, top-level config budgets, then ordered CLI budget overrides. An absent field MUST leave the lower layer unchanged, while a present zero MUST replace it as a valid hard limit. The four-layer example based on `steam-deck` SHALL resolve CLI `nodes=4000`, profile `mesh_instances=1600`, project `shadow_lights=6`, and all other limits from the built-in.

#### Scenario: Four-layer policy
- **WHEN** `shipping` extends `steam-deck`, sets `mesh_instances=1600`, config sets `shadow_lights=6`, and CLI sets `nodes=4000`
- **THEN** effective nodes equal 4000, mesh instances equal 1600, and shadow lights equal 6
- **AND** the other five limits equal the frozen `steam-deck` values

#### Scenario: Explicit zero override
- **WHEN** a higher-priority layer sets a budget to zero
- **THEN** the effective policy retains zero instead of treating it as absent

### Requirement: Repeated CLI budget overrides are strict and ordered
Each CLI budget override SHALL use the exact `metric=limit` form, where metric is one of the frozen eight IDs and limit is a non-negative signed 64-bit base-10 integer. Missing or extra separators, empty components, whitespace-altered metric IDs or limits, unknown metrics, negative values, non-integers, and overflow MUST fail. Repeated metrics SHALL be accepted and applied in input order so the last value wins.

#### Scenario: Duplicate CLI metric
- **WHEN** CLI overrides contain `nodes=1000`, then `lights=8`, then `nodes=2000`
- **THEN** effective nodes equal 2000 and effective lights equal 8

#### Scenario: Invalid CLI budget
- **WHEN** an override contains an unknown metric, negative or non-integer limit, overflow, whitespace, or malformed separator count
- **THEN** resolution fails and identifies the override position and value

### Requirement: Effective policies are non-empty, owned, and deterministic
Successful resolution SHALL return a policy containing at least one configured budget, the selected base kind and ID when present, merged metadata, and owned optional budget values that do not alias the config, catalog, or caller inputs. A no-base policy with only project or CLI budgets SHALL be valid. If every layer contains no budget, resolution MUST fail with `SB2003` and an actionable suggestion to select a base or provide a budget. Every semantic resolution failure SHALL expose `SB2003`, preserve the config source when available, identify the relevant selector/profile/override field, and return no usable policy.

#### Scenario: Top-level-only policy
- **WHEN** no base is selected and top-level config supplies one or more budgets
- **THEN** resolution succeeds with those effective budgets and no selected base

#### Scenario: Empty policy
- **WHEN** no base is selected and neither config nor CLI supplies a budget
- **THEN** resolution fails with `SB2003` and explains how to provide an effective budget

#### Scenario: Caller mutation is isolated
- **WHEN** a caller mutates a returned metadata or budget value and resolves the same inputs again
- **THEN** the later result is unchanged
