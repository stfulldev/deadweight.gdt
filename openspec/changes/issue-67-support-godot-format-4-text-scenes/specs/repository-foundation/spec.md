## ADDED Requirements

### Requirement: Pinned official corpus gates supported text formats
The repository SHALL run the standalone binary against every configured main scene in the pinned official Godot demo-project corpus without installing or invoking Godot. The gate MUST classify format-3 and format-4 roots by their actual `COMPLETE`, `PARTIAL`, or otherwise supported analysis outcome, MUST fail for any root rejected only because it declares format 4, and MUST compare the complete deterministic category summary with committed expectations. UID-only roots and other separately specified unsupported boundaries SHALL remain independently visible rather than being folded into an unexpected-fatal count.

#### Scenario: Format-4 corpus roots
- **WHEN** the pinned corpus contains main scenes whose text headers declare format 4
- **THEN** every such root enters ordinary standalone analysis and none is categorized as an unsupported format-4 input

#### Scenario: Deterministic corpus drift
- **WHEN** a parser or corpus change alters any committed category count or produces an unexpected fatal outcome
- **THEN** the hosted corpus job fails and prints the actual deterministic summary for review

