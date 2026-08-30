## MODIFIED Requirements

### Requirement: Every instance mount receives a target classification
The recursive analyzer SHALL classify every local instance mount without silently discarding any branch. A matching external-resource declaration whose UID or path evidence securely resolves to an existing canonical in-project target with the exact extension `.tscn` SHALL be attempted as a text scene even when its declared type is not `PackedScene`. Missing or wrong external-resource IDs, path- or UID-resolution failures, `SubResource` references, `instance_placeholder` values, imported or binary scene extensions (`.glb`, `.gltf`, `.blend`, `.scn`), and other unsupported targets SHALL produce structured unresolved evidence that preserves the declaring scene, resource ID when present, raw UID/path target, mount identity, mount depth when known, source position, and classification reason. A successfully parsed inherited-root child SHALL use the limited inherited-scene analysis contract and retain approximation evidence instead of being downgraded to an unresolved one-known-root child.

#### Scenario: Existing text scene candidate
- **WHEN** a mount's external-resource declaration resolves by UID or path to an existing canonical `.tscn` file inside the project
- **THEN** that file is attempted as a recursive text-scene target
- **AND** the target is not rejected solely because its declared resource type is incompatible

#### Scenario: UID and path select the same child
- **WHEN** UID-only and path-backed declarations resolve to the same canonical child scene
- **THEN** both occurrences use the same graph and invocation-cache identity while retaining their original reference evidence

#### Scenario: Unsupported instance forms
- **WHEN** mounts use a missing external-resource ID, an unresolvable UID, a `SubResource`, an instance placeholder, an imported or binary extension, or another unresolved secure target
- **THEN** every mount produces explicit unresolved evidence with its original source and target context
- **AND** none is silently omitted from occurrence or coverage accounting

#### Scenario: Inherited target is deferred honestly
- **WHEN** a resolved nested `.tscn` parses successfully and its local summary identifies an inherited root
- **THEN** the child applies its supported base and explicit local contributions through the inherited-scene contract
- **AND** its occurrence retains approximate inheritance evidence rather than claiming exact expansion
