## MODIFIED Requirements

### Requirement: Every instance mount receives a target classification
The recursive analyzer SHALL classify every local instance mount without silently discarding any branch. A matching external-resource declaration with an existing canonical in-project target whose exact extension is `.tscn` SHALL be attempted as a text scene even when its declared type is not `PackedScene`. Missing or wrong external-resource IDs, path-resolution failures, `SubResource` references, `instance_placeholder` values, imported or binary scene extensions (`.glb`, `.gltf`, `.blend`, `.scn`), and other unsupported targets SHALL produce structured unresolved evidence that preserves the declaring scene, resource ID when present, raw target, mount identity, mount depth when known, source position, and classification reason. A successfully parsed inherited-root child SHALL use the limited inherited-scene analysis contract and retain approximation evidence instead of being downgraded to an unresolved one-known-root child.

#### Scenario: Existing text scene candidate
- **WHEN** a mount's external-resource declaration resolves to an existing canonical `.tscn` file inside the project
- **THEN** that file is attempted as a recursive text-scene target
- **AND** the target is not rejected solely because its declared resource type is incompatible

#### Scenario: Unsupported instance forms
- **WHEN** mounts use a missing external-resource ID, a `SubResource`, an instance placeholder, an imported or binary extension, or an unresolved secure path
- **THEN** every mount produces explicit unresolved evidence with its original source and target context
- **AND** none is silently omitted from occurrence or coverage accounting

#### Scenario: Inherited target is deferred honestly
- **WHEN** a resolved nested `.tscn` parses successfully and its local summary identifies an inherited root
- **THEN** the child applies its supported base and explicit local contributions through the inherited-scene contract
- **AND** its occurrence retains approximate inheritance evidence rather than claiming exact expansion

### Requirement: Canonical resource and dependency identities form unique closure sets
The expanded summary SHALL take its dependency identities from the authoritative scene-dependency graph and preserve every successfully loaded nested `.tscn` canonical path reachable through resolved instance or inheritance edges, excluding the analyzed root identity. Resolved inheritance traversal SHALL contribute topology, unique dependency evidence, and the deliberately approximate base-summary contribution defined by the inherited-scene capability. The summary SHALL preserve one external-resource identity for every declaration in every successfully parsed scene used by recursive or graph traversal: the canonical absolute target for a resolved declaration, or the tuple `(declaring canonical scene, document-local resource ID, raw path)` for an unresolved declaration. Applying a cached child more than once MUST union these identities rather than multiply them. Returned identity collections SHALL be owned and deterministically ordered.

#### Scenario: Diamond dependency
- **WHEN** two child branches reach the same canonical descendant scene and declarations resolve to the same canonical external target
- **THEN** each shared scene and resource identity appears once in the expanded closure sets
- **AND** occurrence metrics still include both branches

#### Scenario: Repeated unresolved declaration
- **WHEN** separate parsed scenes declare unresolved resources with the same raw path
- **THEN** their declaring-scene and resource-ID tuple identities remain distinct

#### Scenario: Inherited topology without inherited metric expansion
- **WHEN** a parsed nested scene inherits a resolved supported base scene
- **THEN** the nested and base canonical paths appear in graph-backed dependency identities
- **AND** the base's one-occurrence summary contributes once while the result remains approximate

#### Scenario: Caller mutation cannot alter cached identity sets
- **WHEN** a caller mutates dependency or resource slices returned from a completed invocation
- **THEN** no cached one-occurrence summary or independently repeated invocation is altered
