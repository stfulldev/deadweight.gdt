## MODIFIED Requirements

### Requirement: Contribution reliability is evidence-based
Each contribution row SHALL expose `exact`, `lower_bound`, or `approximate` row reliability as the conservative summary of confidence entries for all eight frozen metrics. Every metric entry SHALL expose its own reliability and deterministic machine-readable reasons derived from evidence that can affect that metric in that row. Unsupported or unavailable nested content MUST make potentially missing contribution values or candidates lower-bound rather than silently treating their unknown remainder as zero; unknown parent composition MUST make the depth candidate lower-bound without degrading unrelated additive values; and inherited roots or override evidence MUST make applicable metrics approximate. Approximate SHALL win when both approximate and lower-bound reasons affect the same metric. Unique-union entries SHALL expose confidence without inventing a per-row owned value.

#### Scenario: Imported child
- **WHEN** an imported scene mount contributes only a known unresolved root occurrence
- **THEN** potentially missing metrics on that unresolved row are `lower_bound` with an imported-scene reason and cannot appear exact
- **AND** its known nodes and scene-instance occurrences remain present

#### Scenario: Inherited contribution
- **WHEN** a scene contribution includes the supported inherited-base approximation or override evidence
- **THEN** applicable metric entries and the conservative row reliability are `approximate` even when every referenced path resolves

#### Scenario: Unrelated exact row
- **WHEN** one resolved scene's direct contribution is fully supported and another row is partial
- **THEN** the supported row's metrics may remain exact while the affected row and whole analysis retain their conservative classifications

#### Scenario: Unknown parent depth
- **WHEN** a row has exact additive values but its root-relative depth cannot be composed because of unsupported parent evidence
- **THEN** only its depth confidence is lower-bound and no exact depth candidate is invented

