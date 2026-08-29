# Local Scene Summary Specification

## Purpose

Defines deterministic, non-recursive extraction of local node contributions, tree depth, external-resource evidence, and scene-instance mount points from one parsed Godot text scene.

## Requirements

### Requirement: Local summaries are isolated and deterministic
The system SHALL convert one valid parsed TSCN document into a local summary without opening files, resolving resource paths, parsing child scenes, or aggregating child summaries. Repeated conversion of equivalent documents SHALL produce equivalent counters and records in deterministic order independent of map iteration.

#### Scenario: Equivalent document is summarized repeatedly
- **WHEN** the same parsed scene document is summarized more than once
- **THEN** every local counter, mount, resource record, classification record, and finding is identical and ordered deterministically
- **AND** no filesystem, network, Godot, child-parser, or process-output operation occurs

### Requirement: Node headers receive one local classification
Every parsed node SHALL be classified exactly once as an ordinary node, nested-scene mount, inherited root, instance placeholder mount, or override stub. A root node carrying an instance reference SHALL be an inherited root rather than a nested-scene occurrence. A non-root node carrying an instance reference or placeholder SHALL be a mount rather than an ordinary node. A node with no literal type, instance, or placeholder SHALL be an override stub and MUST NOT contribute a new node or literal-type count.

#### Scenario: Ordinary local tree
- **WHEN** a scene has a typed ordinary root and typed ordinary descendants without instance references
- **THEN** every header contributes exactly one ordinary local node
- **AND** no mount, inherited-root, placeholder, or override-stub record is produced

#### Scenario: Inherited root and override stub
- **WHEN** the first node carries an instance reference and a later node has no literal type, instance reference, or placeholder
- **THEN** the first node is recorded as an inherited root and the later node as an override stub
- **AND** neither header contributes an ordinary local node or nested-scene occurrence
- **AND** the summary preserves evidence that later whole-scene analysis must handle inheritance and overrides as partial approximate semantics

### Requirement: Supported local parent paths determine exact depth
The system SHALL assign the ordinary non-inherited root depth `1`. The exact parent value `.` SHALL refer to that root, and a canonical slash-separated relative parent path SHALL refer to an already represented local node path below the root. A supported child depth SHALL equal its referenced parent's depth plus one, so `parent="."`, `parent="Arm"`, and `parent="Arm/Hand"` produce depths `2`, `3`, and `4` respectively when those parent nodes exist. Depth calculation SHALL be independent of node declaration order.

#### Scenario: Root and multi-segment descendants
- **WHEN** a scene contains a root, `Arm` with `parent="."`, `Hand` with `parent="Arm"`, and `Finger` with `parent="Arm/Hand"`
- **THEN** their depths are exactly `1`, `2`, `3`, and `4`
- **AND** local tree depth is `4`

#### Scenario: Parent appears after child
- **WHEN** a child uses a supported parent path whose corresponding local node is declared later in the document
- **THEN** the same exact depth is produced as when the parent is declared first

### Requirement: Unsupported parent semantics are explicit partial evidence
The system MUST NOT infer a depth when a non-root parent is missing, ambiguous, absolute, contains `..`, contains empty or `.` path segments outside the exact root marker, or otherwise is not a canonical serialized scene-tree path. The summary SHALL retain the affected node, raw parent value, and source position in a deterministic structured finding, SHALL mark local depth knowledge partial, and SHALL exclude the unknown depth from the known maximum rather than inventing a parent. Unsupported parent semantics SHALL NOT erase independently known local counts or classifications.

#### Scenario: Unknown parent
- **WHEN** an ordinary node or mount names a canonical relative parent path that has no matching local node
- **THEN** its depth is unknown and a source-aware missing-parent finding is returned
- **AND** the summary is partial while independent ordinary and literal-type counts remain available

#### Scenario: Parent traversal
- **WHEN** a node parent is `..`, contains a `..` segment, or otherwise traverses above its serialized local parent
- **THEN** no normalized or guessed depth is assigned
- **AND** a source-aware unsupported-parent finding marks the summary partial

#### Scenario: Absolute NodePath
- **WHEN** a node parent begins with `/` or otherwise uses absolute NodePath semantics
- **THEN** no depth is assigned and a source-aware unsupported-parent finding marks the summary partial

### Requirement: Instance mounts preserve expansion inputs without double-counting
Each non-root instance or placeholder header SHALL produce one mount occurrence containing its local node path when known, depth when known, source position, node name, and original reference evidence. An external instance SHALL be marked as a resolved candidate only when its exact resource ID identifies a local `PackedScene` declaration; this classification MUST NOT claim that the target path exists or is a supported TSCN. Mount occurrences SHALL contribute one local scene-instance occurrence but MUST NOT contribute an ordinary local node, mesh, light, or shadow-light occurrence. Local tree depth SHALL include each known mount depth so a later unresolved mount still represents its known root position.

#### Scenario: PackedScene candidate mount
- **WHEN** a non-root node references an existing local external resource whose literal type is `PackedScene`
- **THEN** one resolved-candidate mount records that resource ID, raw path evidence, and exact mount depth
- **AND** local scene-instance occurrences increase by one while ordinary local nodes do not

#### Scenario: Resolved child is aggregated later
- **WHEN** a later analysis layer replaces a candidate mount with a resolved child summary
- **THEN** the local representation supplies only the mount occurrence and does not supply an extra node for the child's root

#### Scenario: Placeholder mount
- **WHEN** a non-root node has `instance_placeholder` instead of an instance reference
- **THEN** one placeholder mount preserves its raw placeholder target and known depth
- **AND** the summary preserves partial evidence for later analysis without treating it as a resolved candidate

### Requirement: Literal node metrics use the frozen MVP definitions
Ordinary nodes SHALL contribute only from their literal parsed type. Literal `MeshInstance3D` SHALL increment the local mesh count. Literal `DirectionalLight3D`, `OmniLight3D`, and `SpotLight3D` SHALL increment the local light count, and one of those nodes SHALL increment the local shadow-light count only when its parsed `shadow_enabled` value is explicitly `true`. An absent or explicit `false` shadow property MUST contribute zero shadow lights. Other literal types, custom classes, mounts, inherited roots, and override stubs MUST NOT be inferred as one of these metric types.

#### Scenario: Literal supported node types
- **WHEN** ordinary nodes include `MeshInstance3D`, each of the three supported 3D light types, a 2D light, and a custom type
- **THEN** only the literal mesh node contributes to mesh instances
- **AND** only the three supported literal 3D light nodes contribute to lights

#### Scenario: Shadow property values
- **WHEN** supported ordinary 3D lights respectively have `shadow_enabled=true`, `shadow_enabled=false`, and no parsed shadow property
- **THEN** exactly the first light contributes to shadow lights

### Requirement: Local external-resource evidence is complete
The summary SHALL retain every parsed external-resource declaration exactly once, keyed by its document-local resource ID and preserving its literal type, UID, raw path, and source position. Resource records SHALL be deterministically ordered by ID and MUST NOT be collapsed merely because two declarations share a raw path. The local layer SHALL NOT claim final unique `external_resources` or `scene_dependencies` values before path resolution and graph closure.

#### Scenario: Repeated raw target under distinct IDs
- **WHEN** two external-resource declarations use distinct IDs and the same raw path
- **THEN** both local records are preserved in deterministic ID order
- **AND** no final cross-scene uniqueness decision is made
