## Why

The MVP behavior is implemented, but its end-to-end contract is not yet demonstrated by reusable Godot projects, production-path CLI snapshots, and one traceable acceptance matrix. GitHub issue [#22](https://github.com/stfulldev/deadweight.gdt/issues/22) and linked Draft PR [#45](https://github.com/stfulldev/deadweight.gdt/pull/45) close that evidence gap before release hardening.

## What Changes

- Add seven minimal `testdata/projects` groups covering complete, repeated, relative-path, unresolved, inherited, cyclic, and malformed projects, with adjacent expected-behavior notes.
- Add production-path integration tests that exercise the real application and CLI against those fixtures, including stable `<PROJECT>` golden normalization and documented exit codes.
- Strengthen parser fuzz seeds and resource limits so arbitrary bytes cannot panic, hang, or trigger unbounded test allocations.
- Add an acceptance matrix mapping all 25 MVP criteria in `docs/MVP_0.1_SPEC.md` §30 to automated tests or documented verification.
- Repair and enforce the Linux/macOS/Windows GitHub Actions quality gate without installing or invoking Godot.
- Preserve the frozen CLI, parser, metric, policy, and report behavior; this change adds verification and tooling only.

Goals:

- Make every issue #22 acceptance criterion reproducible from committed fixtures and tests.
- Keep snapshots deterministic across machines and operating systems.
- Make the cross-platform repository gate execute the same supported checks developers run locally.

Non-goals:

- Changing user-visible commands, reports, exit-code semantics, parser scope, metric definitions, or budget behavior.
- Adding Godot, a GitHub Action product integration, network access, or runtime dependencies to the shipped binary.
- Duplicating every existing unit case in integration tests when a stable unit test already provides direct evidence.

Compatibility impact: none. The production Go interfaces and CLI contract remain unchanged; only test, fixture, documentation, and CI assets are added or corrected.

Affected MVP acceptance criteria: all §30 criteria are traced, with new end-to-end evidence focused on 1–13 and 16–24. Criterion 25 remains release-documentation work for issue #23 and is explicitly identified as documented pending evidence until that issue lands.

## Capabilities

### New Capabilities

None. This is a pure verification, fixture, documentation, and CI change.

### Modified Capabilities

None. No normative product requirement changes; `.openspec.yaml` therefore declares `skip_specs: true`.

## Impact

- New repository fixtures under `testdata/projects` and golden snapshots under the relevant Go test package.
- New or expanded integration, parser fuzz, and acceptance tests in existing internal packages.
- A traceability document for `docs/MVP_0.1_SPEC.md` §30.
- `.github/workflows/ci.yml` action pins and matrix commands.
- Test-only filesystem I/O; no runtime dependency or shipped binary change.
