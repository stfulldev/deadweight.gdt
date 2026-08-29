## Context

See `proposal.md` for motivation. The CLI now implements every frozen MVP flow and issue #22 added reusable fixtures, deterministic goldens, cross-platform workflow commands, and a 25-row acceptance matrix. README and CHANGELOG still describe the repository before those vertical slices landed. No benchmark currently establishes a repeatable parser or repeated-scene baseline.

The release process must distinguish repository gate results from GitHub-hosted runner startup failures. Recent workflow runs failed with no runner, logs, or steps, while the same build/test/race/vet/lint commands passed locally. Release evidence must record that distinction without relabeling an unexecuted job as a code failure or silently claiming it ran.

## Goals / Non-Goals

**Goals:**

- Make README the accurate public entry point for the shipped `0.1.0` contract.
- Keep every example aligned with the real Cobra flags, strict JSON v1 schema, metric order, preset metadata, and exit behavior.
- Provide small deterministic benchmarks that measure parser work and repeated recursive analysis without filesystem or Godot noise.
- Make the release decision auditable from committed acceptance and checklist documents.
- Ensure the release tag identifies exactly the verified merge commit.

**Non-Goals:**

- Creating binary archives, checksums, package-manager definitions, or release automation.
- Adding benchmark thresholds to CI or treating benchmark numbers as promises.
- Changing default version injection, runtime code, or any frozen MVP semantics.
- Hiding or reclassifying external CI infrastructure failures.

## Decisions

### 1. Organize README as the twelve required public sections

Use explicit headings for what/why, terminal example, install, quick start, metrics, presets/profiles, config, complete/partial, supported inputs, exit codes/CI, roadmap, and contributing/license. Keep installation honest: `go install ...@v0.1.0` is the available release path, while downloadable release binaries are described as future distribution rather than already published assets.

Alternatives considered:

- Keeping a short README and linking almost everything to the technical spec would satisfy maintainers but not first-time users or §33.5.
- Documenting unreleased binary URLs would create a broken public promise.

### 2. Use real CLI language and frozen data, not aspirational examples

Examples use `inspect`, `check`, `presets`, `presets show`, `--project`, `--config`, `--budget`, and partial-policy flags exactly as implemented. The eight metrics and all three built-in preset tables are copied from embedded JSON and labeled `heuristic`/`experimental`. Configuration examples validate against `schema/deadweight.gdt.schema.json` and demonstrate one inheritance/override path without adding undocumented keys.

Alternatives considered:

- A simplified pseudo-config would be easier to read but could teach invalid strict JSON.
- Omitting frozen preset values would leave the release changelog and README unable to substantiate the public baseline.

### 3. Keep disclaimers adjacent to the value proposition and policy docs

The required heuristic warning appears near the opening and again where preset calibration is explained. The static-analysis limitation appears in the reliability section. Steam Deck is explicitly not an official Valve certification profile or endorsement, and the document names runtime-created nodes, scripts, physics, shaders, materials, visibility/culling, imported/binary scenes, and target hardware profiling.

### 4. Add self-contained Go benchmarks with semantic assertions outside timed loops

`BenchmarkParseRepresentativeScene` parses a fixed representative in-memory TSCN and reports allocations. `BenchmarkRecursiveRepeatedScene100` analyzes one in-memory root with one child mounted 100 times; setup happens before timing, every invocation creates its normal in-memory cache, and the benchmark validates result metrics before measuring.

Alternatives considered:

- Benchmarking repository fixtures through `os.Open` would mix filesystem cache behavior into parser/graph baselines.
- Reusing one invocation cache across iterations would measure an impossible CLI lifetime and overstate cache benefit.

### 5. Commit release evidence before performing the tag operation

`docs/RELEASE_0.1.0_CHECKLIST.md` records dependency #19, the full acceptance matrix, documentation review, benchmark smoke execution, quality commands, CI configuration/run evidence, PR merge, and tag verification steps. The OpenSpec change is archived and PR #46 merged first. Then update local `main`, verify its exact commit against the merged PR, create an annotated `v0.1.0` tag, push it, and verify the remote ref.

A final GitHub workflow run with executed repository steps must pass. If GitHub again produces only zero-step jobs with no logs, record that external infrastructure fact separately together with exact-head local gates; never describe the zero-step run as green or as a repository command failure.

Alternatives considered:

- Tagging the feature branch would make the release omit the PR merge identity and violate the normal protected-history shape.
- Creating the tag before checklist/gates complete would make rollback public and error-prone.

## Risks / Trade-offs

- [README examples drift from the CLI] → Use only implemented flags and verify examples against help/schema/source during review.
- [Frozen preset tables are duplicated] → Treat embedded JSON as authoritative and compare every changelog/README value during release review.
- [Benchmarks are optimized away or include setup] → Consume validated results, call `ReportAllocs`, and reset the timer after setup.
- [A public tag is hard to correct] → Verify no existing local/remote `v0.1.0`, exact PR merge SHA, clean tree, and all gates before creating it.
- [Hosted runners fail before execution] → Preserve raw run/job evidence and separate it from local repository gate results.

## Migration Plan

1. Land documentation and benchmark changes in Draft PR #46.
2. Run benchmark smoke tests, all repository quality gates, and strict OpenSpec validation.
3. Archive the change and merge the exact verified PR head.
4. Fast-forward local `main`, verify the merge SHA, create and push annotated `v0.1.0`, then verify the remote tag target.
5. Close the MVP tracker only after issue #23 and the tag are both verified.

Rollback before tagging is a normal PR revert. After publishing the tag, do not silently move it; correct release metadata with a follow-up version unless the repository owner explicitly authorizes tag replacement.
