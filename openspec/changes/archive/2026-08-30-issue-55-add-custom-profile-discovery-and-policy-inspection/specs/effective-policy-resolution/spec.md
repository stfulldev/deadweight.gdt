## ADDED Requirements

### Requirement: Explained custom-profile resolution reuses effective merge semantics
The policy boundary SHALL resolve a requested custom profile into both its effective policy and an explanation without maintaining a second inheritance or override algorithm. Explained resolution MUST validate the complete graph and apply built-in inheritance, custom inheritance, selected profile values, and top-level project budgets in the same field-by-field order as ordinary check policy resolution. It SHALL NOT apply CLI selectors, CLI budget overrides, or CLI partial-policy overrides.

#### Scenario: Effective parity
- **WHEN** explained resolution and ordinary resolution select the same custom profile from the same configuration without CLI overrides
- **THEN** their policy kind, ID, metadata, and budgets are equal

#### Scenario: Shared graph failure
- **WHEN** the configuration contains an unselected invalid profile or a selected inheritance cycle
- **THEN** explained and ordinary resolution fail under the same deterministic dynamic validation rule

### Requirement: Explained resolution tracks value origins through each merge
Each default or built-in base value SHALL begin with its originating layer, and every present custom-profile or project-budget field SHALL replace both the value and its origin. The explanation SHALL retain an ordered root-to-child inheritance chain with distinct built-in-preset and custom-profile layers. Absent budgets SHALL have neither an effective value nor misleading provenance.

#### Scenario: Field-level source replacement
- **WHEN** a parent supplies several metadata fields and a child overrides one of them
- **THEN** only the overridden field identifies the child while the remaining fields retain the parent source

#### Scenario: Project budget replacement
- **WHEN** a selected profile supplies a budget that the top-level project configuration overrides
- **THEN** the effective value and provenance both identify the project override

