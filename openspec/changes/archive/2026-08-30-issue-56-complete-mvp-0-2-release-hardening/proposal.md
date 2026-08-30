## Why

All focused MVP 0.2 capabilities are implemented, but the release still lacks one cross-feature acceptance corpus, supported-OS and official-demo evidence, complete public documentation, and an auditable `v0.2.0` publication procedure. GitHub issue #56 and linked Draft PR #65 close that integration gap only after dependencies #50–#55 have been reviewed, archived, and merged.

## What Changes

- Add a committed MVP 0.2 acceptance matrix mapping every tracker #57 release criterion to automated, hosted, external-corpus, documentation, and release evidence.
- Add one cross-feature Godot project fixture and deterministic CLI goldens spanning schema-v1 JSON, contributions, dependency trees, metric confidence, report diffs, and custom-profile inspection.
- Run that cross-feature corpus on Linux, macOS, and Windows through the existing CI matrix, and add a pinned official `godotengine/godot-demo-projects` E2E job that installs no Godot runtime.
- Update README installation, compatibility, roadmap, and release-evidence guidance from the 0.1.1 baseline to shipped 0.2.0 behavior while keeping the heuristic-preset and honest-static-analysis disclaimers.
- Promote the accumulated Unreleased changelog into a complete dated 0.2.0 record that distinguishes shipped behavior from deferred 0.2.1, 0.3, and ecosystem work.
- Add and execute a dedicated `v0.2.0` release checklist covering local gates, benchmarks, strict OpenSpec validation, hosted CI, pinned official-demo E2E, exact merge identity, annotated tag publication, GitHub Release publication, and tagged `go install` verification.
- Preserve the standalone Go binary, the version-one config/report schemas, the frozen eight metrics and preset values, default text bytes, and exit-code meanings.

### Goals

- Make every MVP 0.2 tracker criterion traceable to reproducible evidence.
- Prove the integrated feature set is deterministic, portable, and compatible on every supported OS.
- Publish `v0.2.0` only from the verified squash-merge commit with immutable tag and release evidence.

### Non-goals

- Adding new analysis, policy, report, or CLI behavior.
- Deep `.tres` parsing, UID resolution, full inherited-scene merging, Godot import bridging, or imported-scene expansion.
- Publishing binary archives, a GitHub Action product, package-manager definitions, SARIF/JUnit output, or editor integration.

### Compatibility and acceptance impact

This release-hardening change is behavior-neutral for the shipped CLI. It proves that the existing 0.1 acceptance goldens remain byte-identical, schema version one remains compatible, and every new 0.2 flow preserves portable deterministic output and established exit semantics. It satisfies issue #56 and tracker #57 only when all child PRs and archived OpenSpec changes, repository gates, supported-OS CI, pinned official-demo E2E, documentation review, exact tag, GitHub Release, and tagged installation are verified.

## Capabilities

### New Capabilities

None. This is an acceptance, test-fixture, CI, documentation, and release-process change.

### Modified Capabilities

None. No normative product requirement changes; `.openspec.yaml` therefore declares `skip_specs: true`.

## Impact

- Documentation: `README.md`, `CHANGELOG.md`, a new MVP 0.2 acceptance matrix, and a dedicated 0.2.0 release checklist.
- Test evidence: one committed cross-feature fixture, deterministic CLI goldens, and cross-checkout/OS assertions.
- Automation: `.github/workflows/ci.yml` gains a pinned official-demo E2E job while retaining the supported-OS matrix.
- Release metadata: annotated tag and non-prerelease GitHub Release `v0.2.0` created only after PR #65 merges to verified `main`.
- Runtime: no production Go-code change, new dependency, Godot requirement, network access, Node.js, or OpenSpec dependency in the shipped binary.
