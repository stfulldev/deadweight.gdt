## Context

See `proposal.md` for motivation. Issues #50–#55 have landed six independently tested capabilities and archived their OpenSpec changes. The repository already has feature-specific goldens, a Linux/macOS/Windows quality matrix, two benchmark baselines, and a PowerShell official-demo sweep pinned to corpus commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`. What is missing is one integrated fixture/evidence path and release metadata that treats those slices as a single versioned product.

This change intentionally has `skip_specs: true`: it validates already specified behavior and changes no normative product contract or production Go code. Release publication is externally visible and difficult to reverse, so exact commit identity and pre-existing tag checks remain controlling gates.

## Goals / Non-Goals

**Goals:**

- Exercise every 0.2 surface through the real command composition and committed filesystem evidence.
- Make portable output comparison reproducible across checkout prefixes and supported operating systems.
- Make tracker, acceptance, release checklist, CI, tag, GitHub Release, and tagged-install evidence agree on one verified commit.
- Keep all external-corpus and release mutations after repository-controlled validation.

**Non-Goals:**

- Adding a generic snapshot framework, test-only production hooks, or a new runtime package.
- Treating pinned demo-project results as a promise of full Godot compatibility.
- Adding performance thresholds or publishing benchmark values as guarantees.
- Moving or replacing any existing public tag.

## Decisions

### 1. Add one committed integration project with before/candidate scene inputs

Create `testdata/projects/mvp-0.2` with strict config, a custom profile inheriting a built-in, a root scene, nested repeated scene evidence, resources, and a candidate form of the same root. Tests copy the fixture to a temporary checkout and replace `root.tscn` with the candidate contents only after capturing the baseline. This lets `diff` compare two documents with the same portable root identity while keeping both source states reviewable.

Alternative considered: build scenes entirely from test strings. Rejected because that would not verify real project/config discovery or leave a contributor-readable fixture.

### 2. Drive the shipped CLI and compare complete JSON goldens

A single integration test will invoke the normal CLI boundary for `inspect` with top contributors, `tree`, `check --profile`, `profiles`, `profiles show`, and `diff`. It will assert exit outcomes and compare complete schema-version-one documents against committed LF-normalized goldens. A second checkout-prefix run will compare bytes before golden matching, proving portable identity rather than hiding canonical paths through normalization.

Alternative considered: assert selected JSON fields only. Rejected because envelope/payload exclusion, ordering, confidence, contributions, chains, and provenance can drift independently.

### 3. Extend existing CI with a pinned official-demo E2E job

Keep the supported-OS matrix unchanged so every committed integration golden runs on Linux, macOS, and Windows. Add one Ubuntu job that builds the binary, checks out `godotengine/godot-demo-projects` at the recorded full commit, and executes the existing PowerShell sweep with `-ExpectedCommit`. The job must report 139 main scenes and zero unexpected fatal results; unsupported format-4 and UID roots remain separately counted. The job installs no Godot executable and therefore verifies the standalone boundary.

Alternative considered: use an unpinned upstream default branch. Rejected because upstream scene additions would make release evidence non-reproducible and unrelated to code changes.

### 4. Keep acceptance evidence human-readable and test-addressable

`docs/MVP_0.2_ACCEPTANCE.md` will map each tracker #57 release-gate row to stable test names, committed artifacts, child issue/PR/archive evidence, hosted jobs, or post-merge release evidence. It will distinguish pre-merge planned entries from verified entries and only mark release/tag rows complete after connector verification.

Alternative considered: infer acceptance from issue checkboxes alone. Rejected because issue state does not record exact commands, fixtures, schemas, or commit identity.

### 5. Prepare release documents on the PR and publish only its merge commit

README and CHANGELOG will describe 0.2.0 as the release being prepared and contain final tagged installation examples. `docs/RELEASE_0.2.0_CHECKLIST.md` records all child issue/PR/archive links, immutable 0.1 tags, local and hosted commands, expected demo counts, absence checks for `v0.2.0`, and the post-merge procedure. After PR #65 is archived, green, ready, and squash-merged, fast-forward `main`; create an annotated tag on that exact merge SHA; push it; create a non-prerelease GitHub Release without binary assets; verify peeled tag identity and a temporary tagged `go install` reports `0.2.0`.

Alternative considered: tag the release branch head before review. Rejected because it would bypass protected main history and make the public tag omit the reviewed squash identity.

## Risks / Trade-offs

- [Full JSON goldens are intentionally broad and can be noisy] → Keep one compact fixture, review every update, and preserve feature-specific unit tests for precise failures.
- [Windows path or newline behavior can change bytes] → Use temporary native paths but require `res://` JSON identities and LF output; run the same test in the existing OS matrix.
- [External demo checkout or runner infrastructure can fail before analysis] → Pin the full corpus SHA and distinguish setup/runner failure from an unexpected analyzer fatal result; never label an unexecuted job green.
- [Release docs claim a version before the tag exists] → Merge, tag, Release, and install verification are one continuous checklist operation; do not leave PR #65 merged without completing publication unless an external outage blocks mutation.
- [Public tag/release correction is destructive] → Verify local and remote absence, exact merge SHA, clean `main`, and all evidence first; never force-update a published tag.

## Migration Plan

1. Land fixture, cross-feature tests, CI job, acceptance matrix, README, CHANGELOG, and release checklist in Draft PR #65 with production/docs and tests committed separately.
2. Run local build/test/race/vet/lint, strict OpenSpec, benchmark, compatibility, and pinned official-demo gates; archive the change and wait for hosted supported-OS plus E2E success.
3. Mark PR #65 ready, squash-merge it, fast-forward clean local `main`, and verify issue #56 closure.
4. Create and push annotated `v0.2.0` at the exact merge SHA, publish the GitHub Release, verify peeled refs and tagged installation, then complete tracker #57 evidence and close it.
5. Delete the merged release branch only after all public release evidence is preserved.

Before public tag creation, rollback is a normal PR update or revert. After publication, corrections require a later semantic version; existing public tags must remain immutable.
