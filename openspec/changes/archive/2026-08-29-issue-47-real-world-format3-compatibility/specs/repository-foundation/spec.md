## ADDED Requirements

### Requirement: Binary version output has truthful build provenance
The executable SHALL choose one user-visible version by the following precedence: a non-development value explicitly injected at link time MUST win; otherwise a semantic version recorded in Go module build metadata SHALL be used with one optional leading `v` removed; otherwise the version SHALL remain `dev`. Tagged module installation, version selection, and `--version` output MUST remain self-contained and MUST NOT require network access, Godot, Node.js, OpenSpec, an MCP server, or loose runtime metadata files after the binary is built.

#### Scenario: Explicit linker version wins
- **WHEN** a source build injects a non-development version through the documented linker variable
- **THEN** `--version` and reports use that exact injected value even if Go module metadata contains another version

#### Scenario: Tagged Go module install
- **WHEN** Go builds the command from module version `v0.1.0` without an explicit linker override
- **THEN** `--version` and reports identify the binary as `0.1.0` rather than `dev`

#### Scenario: Untagged development build
- **WHEN** neither an explicit linker version nor a semantic Go module version is present
- **THEN** `--version` and reports identify the binary as `dev`
