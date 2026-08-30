## ADDED Requirements

### Requirement: Project-context discovery starts at the working directory
For a command that requires project context but has no scene input, the system SHALL validate the supplied working directory and search it and each ancestor for the nearest regular `project.godot`. A valid explicit project path MUST retain precedence and validation semantics identical to scene-based discovery. Missing context MUST return the same typed, actionable project-not-found failure and MUST NOT fabricate or require a scene path.

#### Scenario: Context command inside a nested directory
- **WHEN** a project-context command runs in a project descendant without an explicit project
- **THEN** discovery returns the nearest ancestor containing a regular `project.godot`

#### Scenario: Explicit project from unrelated working directory
- **WHEN** a project-context command supplies a valid explicit project while its working directory is outside that project
- **THEN** discovery returns the explicit project without searching working-directory ancestors

#### Scenario: Invalid working directory
- **WHEN** project-context discovery receives a missing, relative, or non-directory working directory and no usable explicit context
- **THEN** it returns the existing typed invalid-working-directory failure

