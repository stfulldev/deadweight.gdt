## Purpose

Defines deterministic discovery and auditable effective-policy inspection for version-one custom profiles selected from a Godot project configuration.

## ADDED Requirements

### Requirement: Custom profiles are discovered from the selected project configuration
The system SHALL list only custom profiles declared by the selected strict version-one project configuration and SHALL order them by ascending case-sensitive profile ID. Listing MUST validate the complete custom-profile graph with the same collision, parent-existence, cycle, and depth rules used by policy resolution. A configuration with no profiles SHALL produce a successful empty list, while an absent configuration MUST fail with actionable guidance to create the implicit configuration or pass `--config`.

#### Scenario: Canonical profile order
- **WHEN** a selected configuration declares profiles `shipping`, `ci`, and `portable` in any JSON member order
- **THEN** custom-profile discovery returns `ci`, `portable`, and `shipping` in that order

#### Scenario: Empty selected configuration
- **WHEN** the selected configuration is valid but declares no custom profiles
- **THEN** custom-profile discovery succeeds with an empty collection

#### Scenario: Missing configuration
- **WHEN** no explicit configuration is supplied and the project has no implicit `.deadweight.gdt.json`
- **THEN** custom-profile discovery fails with guidance that identifies both supported ways to select a configuration

#### Scenario: Invalid unselected profile
- **WHEN** any declared profile has an unknown parent, colliding ID, inheritance cycle, or excessive inheritance depth
- **THEN** listing fails deterministically even if that profile would not otherwise be displayed in detail

### Requirement: Effective profile inspection is complete and policy-equivalent
For one declared custom profile ID, the system SHALL expose its effective metadata fields `name`, `description`, `platform`, `renderer`, `target_fps`, `quality`, `status`, and `stability`; every effective budget; effective `fail_on_partial`; and its complete inheritance chain. Metadata and budgets MUST exactly match the effective policy used by `check --profile <id>` against the same selected configuration with no CLI budget overrides, and effective `fail_on_partial` MUST match that check without a partial-policy flag.

#### Scenario: Built-in terminal parent with project overrides
- **WHEN** a custom profile extends a built-in preset and the selected configuration overrides one inherited budget at the project layer
- **THEN** inspection shows the built-in and custom inheritance chain, inherited metadata and budgets, and the project-overridden budget
- **AND** the values equal those resolved for `check --profile` under the same configuration

#### Scenario: Custom parent chain
- **WHEN** a profile extends another custom profile that extends a built-in preset
- **THEN** inspection returns the chain from the built-in terminal parent through each custom profile to the selected child

#### Scenario: Unknown custom profile
- **WHEN** inspection requests an ID that is not declared in the selected configuration
- **THEN** it fails without falling through to a same-named or unrelated built-in preset

### Requirement: Every effective value has stable provenance
Inspection SHALL associate every effective metadata value, effective budget, and `fail_on_partial` value with exactly one source layer. Source layers MUST distinguish built-in defaults, a built-in preset ID, a custom profile ID, and the project configuration layer; an explicit child or project value SHALL replace both the inherited value and its source. Provenance MUST be deterministic and MUST NOT depend on JSON member order or map iteration.

#### Scenario: Inherited and overridden sources
- **WHEN** a child profile inherits platform from a built-in preset, overrides quality itself, and receives a top-level nodes budget override
- **THEN** platform identifies the built-in preset, quality identifies the child profile, and nodes identifies the project layer as their respective sources

#### Scenario: Root custom defaults
- **WHEN** a root custom profile omits defaulted metadata and the configuration omits `fail_on_partial`
- **THEN** the root metadata defaults and false partial policy identify the built-in default layer as their source

#### Scenario: Explicit false partial policy
- **WHEN** the selected configuration explicitly declares `fail_on_partial: false`
- **THEN** inspection reports false with the project layer as its source

### Requirement: Built-in and custom namespaces remain distinct
Custom-profile discovery SHALL NOT include built-in presets, and custom-profile inspection SHALL accept only IDs declared in the configuration. Existing built-in preset listing and inspection MUST remain project-independent and MUST NOT include custom profiles.

#### Scenario: Separate list namespaces
- **WHEN** a project declares custom profile `shipping`
- **THEN** `profiles` lists `shipping` without adding built-in presets
- **AND** `presets` retains its frozen built-in-only catalog

