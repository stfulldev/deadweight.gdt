# deadweight.gdt

## What and why

`deadweight.gdt` is a standalone Go CLI that statically analyzes Godot 4 `.tscn` scenes, expands nested text scenes, and checks scene-complexity metrics against project budgets.

It does not require Godot to be installed. The tool is intended for fast local feedback and reproducible CI guardrails: it tells you when a serialized scene crosses limits chosen for your project, and it makes unavailable static data visible instead of silently guessing.

> [!IMPORTANT]
> **Built-in presets are heuristic guardrails, not performance guarantees. Always profile your game on target hardware.**
>
> The `steam-deck` preset is not an official Valve certification profile or endorsement. Equal node counts can have very different runtime costs; renderer settings, scripts, physics, shaders, materials, visibility/culling, and other runtime behavior are outside these presets.

`deadweight.gdt` is not an FPS predictor, compatibility test, certification tool, or performance profiler. Presets are experimental starting points that should be calibrated against measurements from your own project and target hardware.

The current release is `v0.1.1`. The complete MVP implementation contract is in [the MVP specification](docs/MVP_0.1_SPEC.md), and its evidence is mapped in [the acceptance matrix](docs/MVP_0.1_ACCEPTANCE.md).

## Terminal example

```text
$ deadweight.gdt check res://levels/city.tscn --preset steam-deck

deadweight.gdt 0.1.1

Scene:       res://levels/city.tscn
Analysis:    COMPLETE
Preset:      steam-deck
Status:      heuristic (experimental)
Renderer:    Forward+
Target FPS:  60
Quality:     Balanced

Metric                        Actual     Budget   Result
--------------------------------------------------------
Nodes                            2,841      3,000   PASS
Tree depth                          17         20   PASS
Scene instances                    184        250   PASS
Mesh instances                 1,024      1,000   FAIL  +24
Lights                            43         32   FAIL  +11
Shadow lights                       9          8   FAIL  +1
External resources                218        300   PASS
Scene dependencies                 12         80   PASS

FAILED — 3 budgets exceeded
```

The real report includes every configured metric plus preset metadata, coverage, diagnostics, and reliability. Output order is deterministic; redirected output and `NO_COLOR` output contain no ANSI escapes.

## Install

Requirements: Go 1.24 or newer. Godot is not a build-time or runtime dependency.

Install the tagged Go command:

```bash
go install github.com/stfulldev/deadweight.gdt/cmd/deadweight.gdt@v0.1.1
```

Or build from a source checkout and inject the release version:

```bash
git clone https://github.com/stfulldev/deadweight.gdt.git
cd deadweight.gdt
git checkout v0.1.1
go build -ldflags "-X main.version=0.1.1" -o ./bin/deadweight.gdt ./cmd/deadweight.gdt
./bin/deadweight.gdt --version
```

Prebuilt release binaries are planned for a future distribution update; `v0.1.1` does not promise downloadable binary archives or package-manager formulas.

## Quick start

Run from anywhere inside a Godot project or pass its directory/`project.godot` explicitly:

```bash
# Show effective metrics without enforcing a budget.
deadweight.gdt inspect res://levels/city.tscn

# Check all limits in an experimental built-in preset.
deadweight.gdt check res://levels/city.tscn --preset steam-deck

# Use a custom profile from .deadweight.gdt.json.
deadweight.gdt check res://levels/city.tscn --profile shipping

# Apply final one-off limits. --budget is repeatable; the last duplicate wins.
deadweight.gdt check scenes/city.tscn \
  --budget mesh_instances=1600 \
  --budget shadow_lights=6

# Make incomplete static evidence a distinct CI failure.
deadweight.gdt check res://levels/city.tscn \
  --profile shipping \
  --fail-on-partial
```

Useful global flags:

```text
--project PATH   explicit project directory or project.godot
--config PATH    explicit configuration file
--no-color       disable ANSI color
--version        print the binary version
```

Filesystem paths may be absolute or relative. For `res://` input without `--project`, project discovery starts at the current directory and selects the nearest ancestor containing a regular `project.godot`.

## Versioned JSON reports

The development branch for MVP 0.2 adds a machine-readable format to the two scene commands. Text remains the default and retains its existing bytes and color behavior:

```bash
deadweight.gdt inspect res://levels/city.tscn --format json
deadweight.gdt check res://levels/city.tscn --preset steam-deck --format json
```

Every document has `schema_version: 1`, a kind discriminator (`inspect`, `check`, or `error`), and tool name/version metadata. Inspect and check documents contain all eight ordered metrics, checked coverage, grouped diagnostics, portable scene/configuration identity, direct per-scene contributions, and shared unique-union evidence. Check documents additionally contain effective policy metadata, `fail_on_partial`, every configured comparison in canonical metric order, the exceeded count, and the final verdict.

JSON framing follows the command outcome:

- successful inspect and every report-producing check write exactly one JSON document to stdout;
- a fatal error after JSON selection writes exactly one `error` document to stderr and leaves stdout empty;
- exit codes remain `0`, `1`, `2`, and `3` with the meanings below;
- JSON is UTF-8, two-space indented, LF-terminated, deterministic, and independent of terminal color settings.

Successful documents use normalized `res://` paths and do not expose canonical checkout directories. Configuration provenance distinguishes `absent`, `implicit`, and `explicit` selection; an external explicit configuration remains identified as explicit without publishing its machine-specific absolute path. Metric, coverage, comparison, and signed delta values use the signed 64-bit integer domain. Consumers whose default number type cannot represent every `int64` exactly should select an integer-preserving decoder.

The authoritative Draft 2020-12 contract is [`schema/deadweight.gdt.report-v1.schema.json`](schema/deadweight.gdt.report-v1.schema.json). Existing version-one fields and meanings are stable; compatible consumers should ignore unknown optional fields and explicitly reject unsupported `schema_version` values. The committed [JSON golden fixtures](internal/report/testdata/golden/json) are generated by the current encoders, including complete/lower-bound/approximate inspect, passed/failed/incomplete check, and coded/uncoded fatal examples.

## Per-scene contributions

The MVP 0.2 development branch can append a ranked contribution view to `inspect`. Both selectors are explicit and presentation-only:

```bash
deadweight.gdt inspect res://levels/city.tscn --metric nodes --top 10
deadweight.gdt inspect res://levels/city.tscn --metric mesh_instances --top 5 --format json
```

Supported selectors are `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, and `shadow_lights`. The five occurrence metrics are direct additive contributions: literal nodes belong to the scene that declares them, and each non-inheritance mount belongs to its mounted-target row. Their rows reconcile exactly to the root totals without also copying nested values into a parent.

`tree_depth` is a maximum candidate, not a sum. `external_resources` and `scene_dependencies` are shared unique unions, so the CLI deliberately rejects them as top-owner selectors. JSON instead lists every unique target once with its deterministic referrers; a shared texture or diamond dependency is never assigned to an arbitrary owner.

Each row retains its portable scene identity, immediate declaring scene and mount context, occurrence multiplicity, and row-level `exact`, `lower_bound`, or `approximate` reliability. Imported, unavailable, and otherwise unresolved targets cannot appear exact; inherited/override evidence is approximate. This is conservative row-level qualification, not the per-metric confidence model planned separately for MVP 0.2.

## Metrics

All metrics are non-negative integers and always appear in this order:

| ID | Meaning | Aggregation |
|---|---|---|
| `nodes` | Effective node count after nested `.tscn` expansion; a resolved instance root is not counted twice | Per occurrence |
| `tree_depth` | Maximum effective tree depth, with the root at depth `1` | Maximum |
| `scene_instances` | Nested scene instance occurrences, including instances inside every expanded copy; an inherited root is excluded | Per occurrence |
| `mesh_instances` | Nodes whose literal type is `MeshInstance3D` | Per occurrence |
| `lights` | Literal `DirectionalLight3D`, `OmniLight3D`, and `SpotLight3D` nodes | Per occurrence |
| `shadow_lights` | Supported 3D lights with explicit `shadow_enabled = true`; an absent property uses Godot's `false` default | Per occurrence |
| `external_resources` | Unique external resource targets across successfully parsed scenes | Unique union |
| `scene_dependencies` | Unique resolved dependent `.tscn` files, excluding the root scene | Unique union |

These are intentionally narrow static definitions. For example, `mesh_instances` is not a draw-call or triangle estimate, and custom script subclasses are not inferred.

## Presets and custom profiles

List or inspect the immutable built-ins:

```bash
deadweight.gdt presets
deadweight.gdt presets show steam-deck
```

Every built-in is marked `heuristic` and `experimental`. The frozen `0.1.0` limits are:

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

| ID | Renderer | Target metadata | Quality | Status |
|---|---|---:|---|---|
| `mobile` | `mobile` | 30 FPS | `low` | `heuristic` / `experimental` |
| `steam-deck` | `forward_plus` | 60 FPS | `balanced` | `heuristic` / `experimental` |
| `desktop` | `forward_plus` | 60 FPS | `high` | `heuristic` / `experimental` |

Target FPS is descriptive metadata, not a prediction or guarantee. In particular, the Steam Deck row is not official Valve guidance or proof that a project is Steam Deck compatible.

Custom profiles live in project configuration and may extend a built-in or another custom profile. Effective limits merge from lowest to highest priority:

1. built-in preset or ancestor profile;
2. descendant custom profile;
3. top-level project `budgets`;
4. repeated CLI `--budget metric=limit` overrides.

`--preset` selects only built-ins; `--profile` selects only custom profiles. The flags are mutually exclusive.

## Configuration

The default file is `${project_root}/.deadweight.gdt.json`. `--config PATH` selects another file. Configuration is strict JSON schema version 1: unknown fields, unknown metrics, negative/float/string/null limits, profile cycles, missing parents, and built-in ID collisions are errors.

```json
{
  "version": 1,
  "profile": "shipping",
  "fail_on_partial": true,
  "budgets": {
    "shadow_lights": 6
  },
  "profiles": {
    "shipping": {
      "name": "Shipping / Steam Deck",
      "description": "Project-calibrated shipping guardrail",
      "extends": "steam-deck",
      "renderer": "forward_plus",
      "target_fps": 60,
      "quality": "balanced",
      "budgets": {
        "nodes": 5000,
        "mesh_instances": 1600
      }
    }
  }
}
```

`fail_on_partial` affects `check`; it never turns a successful partial `inspect` into a fatal error. `--fail-on-partial` and `--allow-partial` override the configured value for one check and are mutually exclusive.

The canonical Draft 2020-12 schema is [`schema/deadweight.gdt.schema.json`](schema/deadweight.gdt.schema.json).

## Complete vs partial

`COMPLETE exact` means all reachable supported text scenes and their declared resources were available, with no cycle or unsupported static semantics. It is exact only within the deliberately narrow `0.1.0` model.

`PARTIAL lower bound` means missing, imported, binary, placeholder, unsupported, or out-of-project data may contain additional nodes or resources. Known occurrence/closure values are marked with `+`.

`PARTIAL approximate` is used for inherited scenes because `0.1.0` detects and expands a base scene but does not implement full Godot-compatible override merging. Values are marked with `~` because an override can change them in either direction.

> `deadweight.gdt` reads serialized text scenes. It cannot see nodes created by scripts at runtime and cannot expand imported or binary scenes without Godot's import pipeline. When this affects the result, the report is marked `PARTIAL`.

Partial analysis remains visible in coverage, grouped unresolved evidence, and diagnostics. Use `--fail-on-partial` when CI must distinguish “within budget” from “not enough static evidence.”

## Supported and unsupported Godot inputs

Supported root input:

- one existing Godot 4 text scene with the case-sensitive `.tscn` extension and `[gd_scene ... format=3]`;
- string or numeric external-resource IDs;
- absolute, relative, and `res://` paths contained by the canonical project root;
- nested resolved `.tscn` scenes, repeated instances, and safe relative resource paths;
- unknown sections/properties and balanced Variant values, which are skipped by the streaming parser.

Unsupported root input is fatal: Godot 3 `format=2`, `.scn`, `.escn`, `.tres`, `.glb`, `.gltf`, `.blend`, `uid://`-only paths, `user://`, or malformed supported text scenes.

Unsupported nested PackedScene targets are reported as `PARTIAL`, not ignored. This includes imported/binary scenes, missing scenes, `SubResource`, placeholders, and UID-only/user paths. Existing non-scene resources such as scripts, textures, materials, audio, and `.tres` contribute to `external_resources`, but their contents are not deeply parsed.

The CLI never searches for, launches, or communicates with a Godot executable.

## Exit codes and CI

| Code | Meaning |
|---:|---|
| `0` | Command succeeded; `inspect` may be partial, and `check` passed or partial was allowed |
| `1` | One or more configured budgets were exceeded |
| `2` | CLI usage, config, project/root, parser, cycle, or other fatal error |
| `3` | Analysis was partial and the effective policy requires complete evidence |

Priority is fatal error, partial rejected, budget exceeded, success. A partial-rejected check still prints observed comparisons so CI can distinguish insufficient evidence from a known budget failure.

Example GitHub Actions step:

```yaml
- name: Check Godot scene complexity
  run: >-
    deadweight.gdt check res://levels/city.tscn
    --project .
    --profile shipping
    --fail-on-partial
```

The repository quality gate runs build, tests, race tests, and vet on Linux, macOS, and Windows, plus golangci-lint on Linux. No job installs Godot. See [`.github/workflows/ci.yml`](.github/workflows/ci.yml) and the [current release checklist](docs/RELEASE_0.1.1_CHECKLIST.md).

## Roadmap

MVP 0.2 is tracked in [GitHub issue #57](https://github.com/stfulldev/deadweight.gdt/issues/57). Versioned JSON reports and per-scene contribution evidence are its first automation/explainability slices; they do not add report reading, baselines, diffs, dependency-tree rendering, SARIF, project-wide scans, or runtime profiling.

The following remain possible `0.2+` work, not features shipped or promised by `0.1.1`:

- deeper `.tres` dependency analysis;
- dependency-tree rendering, baselines, and diffs;
- full inherited-scene/override semantics and optional Godot import bridging;
- imported scene expansion, UID resolution, and custom class hierarchy support;
- official CI integrations, annotations, release binaries, and package-manager distribution;
- project-wide scans, incremental caches, editor/UI integrations, and additional precisely defined metrics.

Standalone static mode remains the primary product boundary. Runtime, import-pipeline, or broader automation features require separate specifications.

## Contributing and license

Changes are developed through a focused GitHub Issue, branch, linked Draft PR, and one OpenSpec change. Read [the contributor workflow](docs/SPEC_DRIVEN_WORKFLOW.md) and keep roadmap behavior out of unrelated fixes.

Before opening or merging a PR, run:

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0 run
```

Optional benchmark smoke checks:

```bash
go test ./internal/tscn -run '^$' -bench '^BenchmarkParseRepresentativeScene$' -benchtime=100ms
go test ./internal/analysis -run '^$' -bench '^BenchmarkRecursiveRepeatedScene100$' -benchtime=100ms
```

`deadweight.gdt` is distributed under the [MIT License](LICENSE). Copyright © 2026 stfulldev.
