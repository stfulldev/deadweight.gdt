## MODIFIED Requirements

### Requirement: Pinned official corpus gates supported text formats
The repository SHALL run the standalone binary against every configured main scene in the pinned official Godot demo-project corpus without installing or invoking Godot. The gate MUST pass each format-3, format-4, `res://`, filesystem, or `uid://` main-scene value through ordinary analysis, classify it by its actual `COMPLETE`, `PARTIAL`, or unexpected-fatal outcome, and compare the complete deterministic category summary with committed expectations. It MUST fail when any UID root is preclassified as unsupported, when a previously resolved UID becomes missing or ambiguous, or when any category count drifts without review.

#### Scenario: UID corpus roots
- **WHEN** the pinned corpus contains main scenes configured with `uid://` values that uniquely match project-local scene headers
- **THEN** every such root enters ordinary standalone analysis and contributes to `COMPLETE` or `PARTIAL` rather than a separate unsupported-UID category

#### Scenario: Format-4 corpus roots
- **WHEN** the pinned corpus contains main scenes whose text headers declare format 4
- **THEN** every such root enters ordinary standalone analysis and none is categorized as an unsupported format-4 input

#### Scenario: Deterministic corpus drift
- **WHEN** UID resolution, parser behavior, or corpus content alters any committed category count or produces an unexpected fatal outcome
- **THEN** the hosted corpus job fails and prints the actual deterministic summary for review
