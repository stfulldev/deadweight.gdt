# Changelog

All notable changes to this project are documented here.

## Unreleased

## [0.1.1] — 2026-08-30

Corrective release for real-world Godot format-3 compatibility and truthful tagged-install version provenance.

### Fixed

- Accept literal LF, CRLF, and CR line endings inside quoted values and quoted property names in supported Godot `format=3` scenes, preserving source-aware malformed-input diagnostics.
- Derive the displayed version from semantic Go module build metadata when no explicit linker version is supplied, while keeping linker injection authoritative and untagged local builds labeled `dev`.

### Boundaries

- Godot `format=4`, `uid://` root inputs, and project-wide scanning remain unsupported. The published `v0.1.0` tag is immutable and still contains its original version-reporting behavior.

### Validation

- Rechecked 139 declared main scenes from official `godotengine/godot-demo-projects` commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`: 103 complete, 16 expected partial, 9 unsupported format-4 roots, 11 unsupported UID roots, and 0 unexpected fatal results.

## [0.1.0] — 2026-08-29

Initial MVP release of the standalone Go CLI for static Godot 4 `format=3` TSCN scene-complexity analysis.

### Added

- `inspect`, `check`, `presets`, and `presets show` commands with deterministic text reports and documented exit codes `0`, `1`, `2`, and `3`.
- Nearest-project discovery, explicit `--project`, absolute/relative/`res://` scene inputs, canonical root containment, and symlink-escape protection.
- Streaming TSCN subset lexer/parser with source positions, balanced skipping of unknown Variant values, malformed-input diagnostics, and fuzz coverage.
- Recursive nested `.tscn` expansion, dependency graphs, exact cycle chains, repeated-scene parse/summary caches, checked arithmetic, and unique resource/dependency sets.
- Eight frozen metrics: `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, `shadow_lights`, `external_resources`, and `scene_dependencies`.
- Honest `COMPLETE exact`, `PARTIAL lower bound`, and `PARTIAL approximate` evidence with grouped unresolved/inherited diagnostics.
- Strict `.deadweight.gdt.json` schema version 1, custom profile inheritance, four-layer policy merge, final CLI budget overrides, and partial-policy controls.
- Seven committed Godot project fixture groups, production application tests, CLI golden snapshots, acceptance traceability, and Linux/macOS/Windows quality-gate commands without Godot.
- Minimal parser and repeated-scene graph benchmark baselines.

### Frozen experimental presets

All built-ins have status `heuristic` and stability `experimental`. They are starting guardrails, not measured performance guarantees or certification profiles.

| Metric | `mobile` | `steam-deck` | `desktop` |
|---|---:|---:|---:|
| `nodes` | 1,500 | 3,000 | 6,000 |
| `tree_depth` | 15 | 20 | 30 |
| `scene_instances` | 100 | 250 | 500 |
| `mesh_instances` | 500 | 1,000 | 2,500 |
| `lights` | 16 | 32 | 64 |
| `shadow_lights` | 4 | 8 | 16 |
| `external_resources` | 150 | 300 | 600 |
| `scene_dependencies` | 40 | 80 | 160 |

| ID | Renderer | Target metadata | Quality | Status | Stability |
|---|---|---:|---|---|---|
| `mobile` | `mobile` | 30 FPS | `low` | `heuristic` | `experimental` |
| `steam-deck` | `forward_plus` | 60 FPS | `balanced` | `heuristic` | `experimental` |
| `desktop` | `forward_plus` | 60 FPS | `high` | `heuristic` | `experimental` |

### Known limitations

- Static analysis cannot see nodes or resources created by scripts at runtime and does not model physics, shaders, materials, visibility/culling, renderer cost, memory, draw calls, triangles, or FPS.
- Imported or binary scene internals (`.glb`, `.gltf`, `.blend`, `.scn`) are unavailable without Godot's import pipeline and therefore make nested analysis partial.
- Inherited base scenes are expanded, but Godot-compatible property/child override merging is not implemented; affected results are `PARTIAL approximate`.
- `.tres` and ordinary resources contribute to the unique external-resource count but are not deeply parsed.
- The supported root is exactly one Godot 4 text `.tscn` with `format=3`; JSON/SARIF/JUnit/HTML output, project-wide scans, on-disk caches, a GitHub Action product, and editor integrations remain roadmap work.
- Prebuilt release archives and package-manager distributions are not included in this release; installation is supported through Go or a source build.

See [the MVP specification](docs/MVP_0.1_SPEC.md) and [release checklist](docs/RELEASE_0.1.0_CHECKLIST.md) for the frozen contract and verification evidence.

[0.1.1]: https://github.com/stfulldev/deadweight.gdt/releases/tag/v0.1.1
[0.1.0]: https://github.com/stfulldev/deadweight.gdt/releases/tag/v0.1.0
