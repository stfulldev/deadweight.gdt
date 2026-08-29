# Built-in Heuristic Presets Specification

## Purpose

Defines the reproducible built-in heuristic budget catalog and the safe retrieval contract that v0.1 consumers can rely on without treating presets as performance certification.

## Requirements

### Requirement: Frozen v0.1 preset catalog
The system SHALL expose exactly three built-in presets in the product order `mobile`, `steam-deck`, `desktop`. Each preset SHALL contain the exact metadata below, where every status is `heuristic` and every stability is `experimental`.

| ID | Name | Description | Platform | Renderer | Target FPS | Quality |
|---|---|---|---|---|---:|---|
| `mobile` | `Mobile` | `Mobile-class 3D hardware` | `mobile` | `mobile` | 30 | `low` |
| `steam-deck` | `Steam Deck` | `Steam Deck-class hardware` | `steam_deck` | `forward_plus` | 60 | `balanced` |
| `desktop` | `Desktop` | `Mid-range desktop hardware` | `desktop` | `forward_plus` | 60 | `high` |

Each preset SHALL contain all eight limits with the exact frozen v0.1.0 values below.

| Metric | `mobile` | `steam-deck` | `desktop` |
|---|---:|---:|---:|
| `nodes` | 1500 | 3000 | 6000 |
| `tree_depth` | 15 | 20 | 30 |
| `scene_instances` | 100 | 250 | 500 |
| `mesh_instances` | 500 | 1000 | 2500 |
| `lights` | 16 | 32 | 64 |
| `shadow_lights` | 4 | 8 | 16 |
| `external_resources` | 150 | 300 | 600 |
| `scene_dependencies` | 40 | 80 | 160 |

#### Scenario: Retrieve the complete catalog
- **WHEN** a consumer requests the built-in preset catalog
- **THEN** the system returns exactly `mobile`, `steam-deck`, and `desktop` in that order
- **AND** every returned metadata field and limit exactly matches the frozen tables

#### Scenario: Lifecycle metadata is uniform
- **WHEN** any built-in preset is retrieved
- **THEN** its status is `heuristic`
- **AND** its stability is `experimental`

### Requirement: Self-contained versioned preset data
The built-in catalog SHALL be sourced from version-controlled JSON embedded in the shipped binary. Reading built-ins MUST NOT require loose runtime data files, network access, Godot, or an OpenSpec installation.

#### Scenario: Load built-ins in an isolated installation
- **WHEN** the shipped binary runs without the source tree, network access, Godot, or OpenSpec
- **THEN** the complete built-in catalog remains available

### Requirement: Strict embedded-data validation
The system MUST reject an embedded preset record that has an unknown field, a mismatched or duplicate ID, missing metadata, a non-positive target FPS, a status other than `heuristic`, a stability other than `experimental`, a renderer outside `forward_plus`, `mobile`, `compatibility`, and `unspecified`, or a quality outside `low`, `balanced`, `high`, and `custom`. The system MUST also reject a record unless it defines all eight frozen metrics as non-negative integer limits. A validation failure SHALL identify the preset and the invalid field or condition.

#### Scenario: Reject unsupported renderer or quality
- **WHEN** an embedded preset declares an unsupported renderer or quality ID
- **THEN** catalog loading fails with an error that identifies the preset and invalid field

#### Scenario: Reject a negative budget
- **WHEN** any embedded budget limit is negative
- **THEN** catalog loading fails with an error that identifies the preset and invalid metric

#### Scenario: Reject an incomplete budget set
- **WHEN** an embedded preset omits any of the eight required metric limits
- **THEN** catalog loading fails with an error that identifies the incomplete preset

#### Scenario: Reject malformed identity or lifecycle metadata
- **WHEN** an embedded record has a missing or mismatched ID, duplicate ID, unknown field, missing metadata, non-positive target FPS, or unsupported status or stability
- **THEN** catalog loading fails with an actionable error identifying that record and condition

### Requirement: Copy-isolated catalog and lookup results
The system SHALL return built-in catalog and lookup values that do not expose mutable package state. A preset returned from a lookup MUST also be independent of its source catalog, including all optional budget values.

#### Scenario: Mutate a returned catalog
- **WHEN** a caller mutates metadata or a budget in a returned catalog and then requests the built-ins again
- **THEN** the newly returned catalog still contains the frozen metadata and limits

#### Scenario: Mutate a lookup result
- **WHEN** a caller mutates metadata or a budget in a successfully looked-up preset and then repeats the lookup against the same catalog
- **THEN** the repeated lookup still returns the unmodified preset

### Requirement: Actionable preset lookup
Lookup by built-in ID SHALL either return an independent preset value or an error. An unknown-ID error MUST include the requested ID and the available IDs in product order.

#### Scenario: Look up a known preset
- **WHEN** a consumer looks up `steam-deck`
- **THEN** the system returns the complete frozen `steam-deck` value without an error

#### Scenario: Look up an unknown preset
- **WHEN** a consumer looks up `unknown`
- **THEN** no preset is returned
- **AND** the error identifies `unknown` and lists `mobile`, `steam-deck`, and `desktop` in that order

### Requirement: Experimental positioning and change governance
Every product surface that describes a built-in preset SHALL identify it as a heuristic, experimental starting guardrail and MUST NOT imply certified performance, guaranteed FPS, or official Valve endorsement. Frozen IDs, metadata, and limits MUST NOT change in a v0.1 patch release unless the correction is explicitly documented in `CHANGELOG.md`.

#### Scenario: Present a built-in preset
- **WHEN** a user views or applies a built-in preset
- **THEN** the product describes it as a heuristic experimental guardrail rather than a performance guarantee

#### Scenario: Correct frozen data in a patch release
- **WHEN** a necessary correction changes a frozen built-in ID, metadata field, or limit in a v0.1 patch release
- **THEN** the same release includes a `CHANGELOG.md` entry that identifies the correction
