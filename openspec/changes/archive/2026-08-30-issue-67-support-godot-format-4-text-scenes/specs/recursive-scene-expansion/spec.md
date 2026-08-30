## MODIFIED Requirements

### Requirement: Supported text scenes expand recursively
An existing canonical non-inherited `.tscn` target that parses as supported Godot format 3 or format 4 SHALL be converted to its local summary and expanded recursively for its own nested mounts. A syntax or supported-format parse failure in a resolved nested `.tscn` SHALL remain a fatal typed analysis failure rather than being downgraded to an unresolved instance. Expansion MUST use canonical absolute scene identities for loading and memoization while retaining normalized display and original target identities for later presentation.

#### Scenario: Three-scene chain
- **WHEN** root scene A mounts supported scene B and B mounts supported scene C across any supported combination of format 3 and format 4
- **THEN** the expanded root summary includes the per-occurrence contributions and closure evidence of B and C
- **AND** each load uses the canonical identity returned by secure path resolution

#### Scenario: Malformed resolved text scene
- **WHEN** a resolved `.tscn` target declares format 3 or format 4 but cannot be parsed as the supported subset
- **THEN** recursive expansion returns the typed parse failure and does not publish a truncated expanded summary

#### Scenario: Unknown future nested format
- **WHEN** a resolved `.tscn` dependency declares an unknown format greater than 4
- **THEN** recursive expansion returns the typed unsupported-format failure rather than treating the dependency as an unresolved partial mount

