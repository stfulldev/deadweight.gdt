## MODIFIED Requirements

### Requirement: Every partial branch produces grouped deterministic diagnostics
Every unresolved scene mount, imported or binary scene, inherited target, unavailable declaration, placeholder, unresolved UID or `user://` path, `SubResource` scene source, unsupported scene target, and unsupported parent finding SHALL produce validated warning evidence. Diagnostics SHALL use `SB1001` for otherwise unresolved scene instances, `SB1002` for imported or binary scenes, `SB1003` for inheritance, `SB1004` for missing, unreadable, empty, or outside-project resource paths, `SB1005` for placeholders, `SB1006` for missing, malformed, ambiguous, stale, unsupported, or unsafe UID evidence and `user://` paths, and `SB1008` for unsupported parent semantics. A successfully UID-resolved target MUST NOT become partial merely because its original reference used `uid://`. Equivalent evidence SHALL be grouped by stable code, display path, classification or reason, and target/resource identity with checked occurrence totals; returned diagnostics SHALL be owned and sorted by severity, code, display path, line, column, resource, and message independently of map order.

#### Scenario: Repeated imported target
- **WHEN** one imported scene target occurs repeatedly through cached scene summaries
- **THEN** one `SB1002` warning retains the complete checked occurrence count

#### Scenario: Distinct unresolved reasons
- **WHEN** equivalent raw targets fail for different resolution reasons or declaring scenes
- **THEN** they remain distinguishable diagnostic groups in deterministic order

#### Scenario: Resolved UID target
- **WHEN** a root or nested resource UID uniquely and securely resolves to an otherwise supported project file
- **THEN** UID use alone emits no `SB1006` warning and does not reduce analysis or metric confidence

#### Scenario: Ambiguous UID target
- **WHEN** a requested UID has conflicting project mappings and cannot be selected exactly
- **THEN** analysis retains grouped `SB1006` evidence with the UID, ambiguity reason, and complete checked occurrence count

#### Scenario: Unsupported parent
- **WHEN** a parsed local node has an invalid, missing, ambiguous, absolute, or otherwise unsupported parent path
- **THEN** analysis is `partial lower_bound` and emits grouped `SB1008` evidence without inventing tree depth
