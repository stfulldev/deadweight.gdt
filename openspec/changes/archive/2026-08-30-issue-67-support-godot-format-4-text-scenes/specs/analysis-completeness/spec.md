## MODIFIED Requirements

### Requirement: Successful analysis exposes status and reliability
Every successful root result SHALL expose status `complete` or `partial` and reliability `exact`, `lower_bound`, or `approximate`. Status SHALL be `complete` with `exact` reliability only when every reachable scene occurrence and every declared resource is statically accounted for and no inherited or unsupported parent semantics remain. Any non-inheritance partial reason SHALL produce `partial` with `lower_bound`; any inherited-scene or override evidence SHALL produce `partial` with `approximate`, and `approximate` MUST win when both classes occur.

#### Scenario: Fully resolved closure
- **WHEN** all reachable supported format-3 and format-4 text scenes and all declared resources resolve successfully without inheritance or unsupported parent semantics
- **THEN** the successful result is `complete` and `exact`

#### Scenario: Mixed partial reasons
- **WHEN** one closure contains both a missing nested scene and inherited-scene evidence
- **THEN** the successful result is `partial` and `approximate`

