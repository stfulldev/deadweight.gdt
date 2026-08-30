## ADDED Requirements

### Requirement: Diff text presents semantic changes and uncertainty
Diff text SHALL render tool version, compatible report kind, portable root scene, before/after analysis reliability, changed metrics in canonical order with before, after, signed delta, confidence-aware assessment, and confidence source where needed. It SHALL then render only changed coverage, diagnostic, dependency, and evaluation evidence in deterministic order, followed by explicit enforcement status and triggers. An empty semantic diff SHALL state that no semantic changes exist. Regression, improvement, uncertain, failed, and incomplete meaning MUST remain understandable without color, and output SHALL use exactly one trailing LF.

#### Scenario: Empty text diff
- **WHEN** the inputs have equal semantic evidence
- **THEN** text states `No semantic changes.` and reports enforcement passed or disabled without empty tables

#### Scenario: Qualified decrease
- **WHEN** a numerically smaller candidate is not a proven improvement because its confidence is non-exact
- **THEN** text labels it `UNCERTAIN` and names the before/after reliability instead of using improvement language

#### Scenario: Report-first enforcement failure
- **WHEN** selected regression enforcement produces exit `1` or `3`
- **THEN** stdout contains the complete deterministic diff and enforcement summary before the centralized outcome is returned

