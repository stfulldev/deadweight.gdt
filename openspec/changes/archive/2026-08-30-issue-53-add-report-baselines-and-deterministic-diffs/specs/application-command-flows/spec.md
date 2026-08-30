## ADDED Requirements

### Requirement: Diff executes a project-independent baseline flow
The `diff` command SHALL accept exactly two positional JSON report paths, read and compare them through an injected application boundary, and avoid working-directory discovery, project discovery, configuration loading, scene resolution, analysis, policy resolution, and Godot. It SHALL accept `--format text|json`, repeatable `--fail-on-increase METRIC` for any frozen metric, and `--fail-on-reliability`; reject invalid format, argument count, duplicate selected metrics, and unknown metrics before reading files; render every non-fatal result before centrally returning exit `0`, `1`, or `3`; and map every read, decode, compatibility, or validation failure to exit `2` using the selected error representation.

#### Scenario: Compare two baselines offline
- **WHEN** a user runs `diff before.json after.json`
- **THEN** the application reads exactly those two files, performs no scene flow, and renders deterministic text with exit `0` unless an explicit policy triggers

#### Scenario: JSON diff selection
- **WHEN** the same valid comparison uses `--format json`
- **THEN** only presentation bytes change and the semantic result and exit outcome remain equivalent

#### Scenario: Invalid selected metric
- **WHEN** `--fail-on-increase` names an unknown or duplicate metric
- **THEN** diff returns actionable exit `2` guidance before reading either baseline

