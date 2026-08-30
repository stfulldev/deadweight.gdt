## Why

Issue [#68](https://github.com/stfulldev/deadweight.gdt/issues/68), tracked by MVP 0.3 issue [#66](https://github.com/stfulldev/deadweight.gdt/issues/66) and Draft PR [#75](https://github.com/stfulldev/deadweight.gdt/pull/75), is the next compatibility boundary after format-4 support. The pinned official corpus now has eleven otherwise analyzable projects whose main scenes are hidden behind `uid://` values even though each UID has unique project-local scene evidence, so rejecting every UID reference leaves deterministic standalone information unused.

## What Changes

- Build one lazy, invocation-scoped UID index from supported project-local Godot 4 evidence: text resource headers, version-applicable `.uid` sidecars, imported-resource `.import` metadata, and the configured project-data `uid_cache.bin` when present.
- Define deterministic source precedence, Godot-compatible UID text parsing, secure path validation, duplicate handling, stale-cache fallback, and typed missing, malformed, ambiguous, or unsafe UID outcomes without launching Godot.
- Accept a `uid://` root when the project is explicit or discoverable from the working directory, and resolve UID-only scene/resource references through the same canonical path, graph, cache, contribution, and reporting pipeline as path-backed references.
- Preserve Godot's path fallback for a declaration that contains both UID and usable path evidence; unresolved UID-only references remain fatal for a root and honest partial evidence for nested or ordinary resources.
- Move the eleven pinned-corpus UID main scenes into ordinary `COMPLETE` or `PARTIAL` analysis, record the measured deterministic baseline, and retain zero unexpected fatal outcomes.
- Preserve all existing `res://`, relative, absolute, metric, preset, budget, exit-code, text-report, and JSON schema-v1 behavior. This change adds no Godot, network, registry, script-execution, or runtime-service dependency.

## Capabilities

### New Capabilities

- `project-uid-resolution`: Construct and query a deterministic, secure, version-aware project-local UID index with explicit evidence precedence and failure classifications.

### Modified Capabilities

- `project-root-discovery`: Treat `uid://` scene inputs as project-context inputs whose automatic discovery starts at the working directory while preserving explicit-project precedence.
- `secure-path-resolution`: Resolve UID roots and resources through the project UID index, apply canonical containment to mapped paths, and preserve path fallback plus typed unresolved evidence.
- `application-command-flows`: Accept supported `uid://` scene inputs in inspect, check, and tree flows without changing their downstream orchestration or exit semantics.
- `recursive-scene-expansion`: Expand UID-resolved text-scene mounts through the existing one-occurrence cache and graph contracts instead of classifying every UID target as unresolved.
- `inherited-scene-analysis`: Apply a securely UID-resolved supported base once while keeping unresolvable UID bases approximate with explicit evidence.
- `analysis-completeness`: Treat only unresolved UID evidence as `SB1006` partial uncertainty; successful UID resolution must not reduce confidence.
- `repository-foundation`: Replace the pinned corpus's separate unsupported-UID-root bucket with measured ordinary complete/partial outcomes and deterministic drift checks.

## Impact

- Expected implementation areas: `internal/project` discovery and secure resolution, a focused UID index/metadata reader boundary, application composition, recursive and inherited analysis, completeness/confidence mapping, CLI integration fixtures, and the official-demo PowerShell runner.
- The index may inspect project-local metadata and bounded resource headers only when a UID is encountered; normal path-only analysis remains independent of UID scanning.
- The binary format reader must use checked lengths and bounded allocation, directory traversal must not follow escaping symlinks, and every returned identity remains an existing canonical regular file inside the selected project.
- The current pinned corpus proves a minimum movement of ten UID roots to `COMPLETE` and one to `PARTIAL`; the final committed counts will be measured against all 139 main scenes after nested UID resolution is enabled.
- Non-goals remain Godot bridge execution, inherited override merging, imported/binary scene expansion, remote UID registries, arbitrary script execution, deep resource parsing, and new metrics or report schema versions.
