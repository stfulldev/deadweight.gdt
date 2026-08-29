## ADDED Requirements

### Requirement: Inspect validates top-contributor presentation selectors
The `inspect` command SHALL accept `--metric METRIC` together with `--top LIMIT` to request an opt-in top-contributors projection. Both flags MUST be supplied together, `LIMIT` MUST be a positive signed 64-bit integer, and `METRIC` MUST be one of `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, or `shadow_lights`. Missing pairs, non-positive or overflowing limits, unknown metrics, and unique-union metric selections MUST fail as actionable usage errors before project discovery or application analysis. The selectors SHALL remain presentation inputs and MUST NOT change the application request, analysis result, or exit outcome.

#### Scenario: Valid explicit selection
- **WHEN** inspect receives `--metric nodes --top 5`
- **THEN** it performs the ordinary inspect application flow once and supplies the resulting contribution evidence to the selected presentation

#### Scenario: Missing selector pair
- **WHEN** inspect receives only `--metric` or only `--top`
- **THEN** it exits `2` with actionable usage guidance before invoking the application service

#### Scenario: Unsupported unique metric
- **WHEN** inspect receives `--metric external_resources --top 5`
- **THEN** it exits `2` before analysis and explains that unique-union metrics have no additive top-owner ranking

#### Scenario: Check remains unchanged
- **WHEN** a user invokes `check`
- **THEN** top-contributor selectors are unavailable and check request, report, and exit semantics remain unchanged

