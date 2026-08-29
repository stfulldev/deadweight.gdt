## Why

The MVP implementation and acceptance corpus are complete, but the public README still describes an early skeleton and the repository lacks release-grade changelog, benchmark, and tag evidence. GitHub issue [#23](https://github.com/stfulldev/deadweight.gdt/issues/23) and linked Draft PR [#46](https://github.com/stfulldev/deadweight.gdt/pull/46) close the remaining documentation and release-hardening gap for `v0.1.0`.

## What Changes

- Rewrite README around the shipped standalone Godot 4 TSCN analyzer, covering all twelve §33.5 sections with accurate commands, metrics, configuration, reliability, support boundaries, exit codes, CI use, roadmap, contribution, and license guidance.
- State the heuristic-preset and static-analysis limitations prominently, including the required Steam Deck/Valve and target-hardware profiling disclaimers.
- Promote the changelog from the early implementation slice to a dated `0.1.0` release record with frozen preset values, experimental status, and known MVP limitations.
- Add minimal parser and repeated-scene graph benchmarks that can be run without Godot.
- Add a release checklist linked to the 25-row acceptance matrix, local quality gates, cross-platform CI evidence, dependency #19, and exact post-merge tag procedure.
- Update criterion 25 in the acceptance matrix from tracked to verified once README evidence exists.
- Prepare and create the `v0.1.0` tag only from the verified merge commit after PR #46 is merged and all release criteria are satisfied.

Goals:

- Make the first public documentation accurate for the code that ships.
- Keep every performance and compatibility claim conservative and auditable.
- Produce repeatable release evidence before the tag is pushed.

Non-goals:

- Adding roadmap behavior, binary packaging, a GitHub Action product, generated release assets, or a hosted release page.
- Changing CLI commands, metrics, preset values, configuration semantics, or runtime code.
- Publishing benchmark targets as performance guarantees.

Compatibility impact: none. Documentation, benchmarks, and release metadata do not change the Go API, CLI contract, config schema, or supported input subset.

Affected MVP acceptance criteria: §30 items 23–25, with the complete 1–25 matrix re-verified before tagging.

## Capabilities

### New Capabilities

None. This is a documentation, benchmark, and release-process change.

### Modified Capabilities

None. No normative product requirement changes; `.openspec.yaml` therefore declares `skip_specs: true`.

## Impact

- `README.md`, `CHANGELOG.md`, `docs/MVP_0.1_ACCEPTANCE.md`, and a new release checklist.
- Test-only benchmark files in `internal/tscn` and `internal/analysis`.
- Git tag `v0.1.0` after the verified PR merge.
- No new dependency, shipped runtime code, Godot requirement, or network access.
