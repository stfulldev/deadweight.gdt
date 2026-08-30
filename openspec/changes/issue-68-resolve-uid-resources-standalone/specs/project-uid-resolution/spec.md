## Purpose

Defines deterministic standalone conversion of Godot 4 `uid://` identifiers into secure project-local resource paths without requiring Godot or external registries.

## ADDED Requirements

### Requirement: UID text has one canonical project identity
The system SHALL recognize canonical Godot 4 `uid://` text as a stable non-negative resource identifier and SHALL reject empty, malformed, non-canonical, or unsupported future UID syntax with a typed reason. Equivalent canonical UID text MUST compare by decoded identity rather than process locale, host path rules, or map order.

#### Scenario: Canonical Godot UID
- **WHEN** a project reference contains a canonical lowercase Godot 4 `uid://` value
- **THEN** repeated decoding produces the same stable resource identifier on every supported host

#### Scenario: Invalid UID text
- **WHEN** a UID is empty, contains an invalid character, does not round-trip to canonical Godot text, or uses unsupported future syntax
- **THEN** the value is rejected with typed malformed or unsupported evidence and no filesystem path is guessed

### Requirement: The project UID index uses supported local evidence
The project UID index SHALL inspect only the selected project and SHALL recognize supported Godot 4 ownership evidence from `.tscn` and `.tres` resource headers, version-applicable `.uid` sidecars, imported-resource `.import` remap metadata, and the configured project-data `uid_cache.bin` when present. It SHALL derive the project-data directory from `application/config/use_hidden_project_data_directory`, apply only metadata forms supported by the project's declared Godot feature version, skip unsupported future forms explicitly, and MUST NOT launch Godot, read a remote registry, execute scripts, or follow a directory symlink outside the project.

#### Scenario: Cold source checkout
- **WHEN** a Godot 4 project has no project-data cache but a text scene header declares the requested UID
- **THEN** the index maps that UID to the scene from source-controlled project evidence alone

#### Scenario: Version-applicable sidecar
- **WHEN** a project version that supports UID sidecars contains a well-formed adjacent `.uid` file for a regular resource
- **THEN** the sidecar contributes ownership evidence for that resource

#### Scenario: Configured project-data cache
- **WHEN** the project selects its hidden or non-hidden project-data directory and that directory contains a supported `uid_cache.bin`
- **THEN** only that configured cache path contributes cache evidence

### Requirement: Direct ownership evidence outranks generated cache evidence
Unique direct ownership evidence from a resource header, applicable `.uid` sidecar, or `.import` remap SHALL outrank `uid_cache.bin` evidence for the same UID. Conflicting direct claims for distinct canonical files MUST make the UID ambiguous; a cache-only mapping MAY resolve only when its target is a valid in-project regular file. A stale or conflicting cache entry MUST NOT override a unique direct claim, and duplicate or conflicting evidence SHALL be retained in deterministic canonical-path order.

#### Scenario: Stale cache and current resource header
- **WHEN** a cache maps a UID to an old path while one current resource header uniquely claims that UID at another valid path
- **THEN** the current direct ownership path wins and the stale cache cannot redirect analysis

#### Scenario: Duplicate direct UID
- **WHEN** two distinct project resources directly claim the same UID
- **THEN** lookup is ambiguous and the system does not select either resource by traversal or map order

#### Scenario: Cache-only valid mapping
- **WHEN** no direct source claims a UID and the supported cache uniquely maps it to an existing regular file inside the project
- **THEN** lookup returns that securely validated project resource

### Requirement: Every UID lookup has typed deterministic evidence
A lookup SHALL return either one resolved project path or typed missing, malformed, ambiguous, stale, unsupported-version, outside-project, or filesystem evidence. A successful path MUST preserve the original UID, normalized `res://` display path, canonical absolute I/O identity, and evidence source. Returned results and evidence collections SHALL be owned and deterministic, and no unresolved outcome may expose an outside-project path as usable.

#### Scenario: Unknown UID
- **WHEN** no supported project-local source contains the requested canonical UID
- **THEN** lookup returns typed missing evidence with the original UID and no candidate path

#### Scenario: Unsafe cached path
- **WHEN** a cache entry points through a symlink or lexical path outside the canonical project root
- **THEN** lookup returns outside-project evidence and never opens or returns the outside target as usable

#### Scenario: Corrupt relevant metadata
- **WHEN** the only evidence for a requested UID has a truncated length, invalid text, or otherwise unsupported representation
- **THEN** lookup returns deterministic malformed or unsupported evidence without panic or unbounded allocation

### Requirement: UID indexing is lazy and invocation scoped
Path-only analysis SHALL NOT scan UID metadata. The first UID lookup for one selected project SHALL construct at most one immutable index for that application invocation, and all later root, resource, graph, inheritance, contribution, and report operations in that invocation SHALL reuse it. Index state MUST NOT persist on disk or across invocations, and traversal and parsing effects SHALL be independently controllable by tests.

#### Scenario: Path-only analysis
- **WHEN** an analysis closure contains no `uid://` root or resource evidence
- **THEN** no UID directory traversal, cache read, sidecar read, or resource-header scan occurs

#### Scenario: Repeated UID references
- **WHEN** one invocation resolves the same or different UIDs across repeated and diamond-shaped scene occurrences
- **THEN** the project index is built once and each lookup returns owned deterministic evidence without rescanning the project
