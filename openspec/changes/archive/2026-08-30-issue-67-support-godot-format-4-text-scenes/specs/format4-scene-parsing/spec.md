## Purpose

Defines the deterministic standalone parsing boundary for Godot format-4 text scenes while preserving the existing minimal AST and format-3 compatibility.

## ADDED Requirements

### Requirement: Supported text-scene versions are explicit
The text-scene parser SHALL accept a scalar integer `format` value of `3` or `4` in the first `[gd_scene]` header and SHALL retain the accepted value in the parsed scene header. A missing, non-integer, older, or unknown future format MUST fail with a source-positioned typed `SB2001` error that identifies formats 3 and 4 as the supported set; the parser MUST NOT optimistically interpret an unrecognized version.

#### Scenario: Format-4 root
- **WHEN** a structurally valid `.tscn` begins with `[gd_scene format=4]`
- **THEN** parsing succeeds and the parsed header retains format 4

#### Scenario: Format-3 compatibility
- **WHEN** an existing valid format-3 fixture is parsed after format-4 support is enabled
- **THEN** its document fields, recognized property behavior, and failures remain unchanged

#### Scenario: Unsupported past or future version
- **WHEN** a `.tscn` header declares format 2 or an unknown version greater than 4
- **THEN** parsing fails with positioned `SB2001` guidance that only formats 3 and 4 are supported

### Requirement: Format-4 packed values remain opaque and balanced
For format-4 section properties outside the minimal recognized metric subset, the streaming parser SHALL consume base64-string `PackedByteArray(...)`, `PackedVector4Array(...)`, and their nested container forms with the existing quote, newline, escape, comment, and balanced-delimiter rules. It MUST NOT decode packed bytes, materialize a Variant AST, or derive metric evidence from an opaque payload. TSCN token or delimiter corruption MUST remain a positioned `SB2001` failure even when the surrounding property is otherwise ignored.

#### Scenario: Base64 packed byte array
- **WHEN** an ignored format-4 property contains a closed `PackedByteArray("...")` string payload followed by recognized scene sections
- **THEN** the payload is consumed without decoding and subsequent nodes and resources are parsed normally

#### Scenario: Packed vector-four array
- **WHEN** an ignored format-4 property contains a balanced `PackedVector4Array(...)` value nested inside an array or dictionary
- **THEN** the entire value is skipped and no vector or metric value is invented

#### Scenario: Malformed packed value
- **WHEN** a format-4 packed value has an unterminated string or mismatched or unclosed delimiter
- **THEN** parsing fails with positioned `SB2001` rather than publishing a truncated document

### Requirement: Format-4 metadata preserves existing analysis identities
Format-4 scenes SHALL use the same external-resource IDs, filesystem paths, node names, parent paths, owners, instances, editable markers, and recognized metric properties as the existing minimal document contract. A scalar node `unique_id` attribute SHALL be accepted as non-metric metadata and MUST NOT replace canonical scene paths or serialized node paths as graph, cache, contribution, or aggregation identities in this change.

#### Scenario: Nodes with unique IDs
- **WHEN** format-4 nodes include integer `unique_id` attributes alongside ordinary names, types, parents, and instances
- **THEN** the ordinary fields drive the same local summary as their equivalent format-3 scene and unique IDs do not multiply or rename nodes

#### Scenario: Format-4 PackedScene dependency
- **WHEN** a format-4 external resource provides an ordinary path and is mounted through its string resource ID
- **THEN** the declaration and mount enter the existing secure-resolution and recursive-analysis contracts

