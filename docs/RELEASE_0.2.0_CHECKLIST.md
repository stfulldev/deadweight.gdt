# v0.2.0 release checklist

This is the operator procedure and repository-controlled evidence plan for `deadweight.gdt` `v0.2.0`. The public tag must identify the verified squash-merge commit from [PR #65](https://github.com/stfulldev/deadweight.gdt/pull/65), never the release branch head.

## Scope and immutable dependencies

- [x] [Issue #50](https://github.com/stfulldev/deadweight.gdt/issues/50) / [PR #58](https://github.com/stfulldev/deadweight.gdt/pull/58): versioned JSON reports.
- [x] [Issue #51](https://github.com/stfulldev/deadweight.gdt/issues/51) / [PR #60](https://github.com/stfulldev/deadweight.gdt/pull/60): per-scene contributions.
- [x] [Issue #52](https://github.com/stfulldev/deadweight.gdt/issues/52) / [PR #61](https://github.com/stfulldev/deadweight.gdt/pull/61): explainable dependency tree.
- [x] [Issue #53](https://github.com/stfulldev/deadweight.gdt/issues/53) / [PR #63](https://github.com/stfulldev/deadweight.gdt/pull/63): portable baselines and deterministic diffs.
- [x] [Issue #54](https://github.com/stfulldev/deadweight.gdt/issues/54) / [PR #62](https://github.com/stfulldev/deadweight.gdt/pull/62): per-metric confidence.
- [x] [Issue #55](https://github.com/stfulldev/deadweight.gdt/issues/55) / [PR #64](https://github.com/stfulldev/deadweight.gdt/pull/64): custom-profile discovery and effective-policy inspection.
- [x] Every dependency issue is closed and its dated OpenSpec change is archived; [`MVP_0.2_ACCEPTANCE.md`](MVP_0.2_ACCEPTANCE.md) records the exact mapping.
- [x] Annotated tags `v0.1.0` and `v0.1.1` remain immutable at peeled commits `4596b4fbc195b59dc1883b648b6c35dd913ce1a0` and `d1b6a4df06976e67951cec7129e31aee9a57b125`.

The release adds no production runtime dependency, new metric, changed preset value, new config/report schema version, deep `.tres` parsing, UID resolution, full inherited-scene merging, Godot bridge, official GitHub Action product, prebuilt archive, or package-manager promise.

## Public documentation and compatibility

- [x] README identifies `v0.2.0` as current and uses the exact tag in Go-install and source-build examples.
- [x] README separately describes shipped MVP 0.2 behavior and deferred 0.2.1, 0.3, and ecosystem work.
- [x] README and CHANGELOG retain the heuristic-preset, Steam Deck non-endorsement, static-analysis, partial-evidence, and no-Godot-runtime boundaries.
- [x] CHANGELOG has a dated `0.2.0` entry for every capability delivered by issues #50–#55, compatibility, validation, and deferred work.
- [x] The frozen MVP 0.1 text goldens have no byte diff from `v0.1.1`; exits, metrics, presets, config schema v1, and report schema v1 remain compatible.
- [x] Installation remains Go/source based and the GitHub Release contains no promised binary assets.

## Repository-controlled verification

Run from the clean exact PR head:

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0 run
openspec validate --all --strict
openspec status --change issue-56-complete-mvp-0-2-release-hardening
go test ./internal/tscn -run '^$' -bench '^BenchmarkParseRepresentativeScene$' -benchtime=100ms
go test ./internal/analysis -run '^$' -bench '^BenchmarkRecursiveRepeatedScene100$' -benchtime=100ms
git diff --exit-code v0.1.1 -- internal/cli/testdata/golden/acceptance
git diff --check
git status --short --branch
```

Every command must pass. Archive `issue-56-complete-mvp-0-2-release-hardening` only after its tasks are complete, validate the archive strictly, and confirm all main OpenSpec capabilities remain valid.

## Hosted and external-corpus verification

GitHub Actions must complete these jobs successfully on the exact final PR head:

- `lint` on Ubuntu with golangci-lint `v2.12.0`;
- `test (ubuntu-latest)`, `test (macos-latest)`, and `test (windows-latest)`, each running build, tests, race tests, and vet;
- `official-demo-e2e` on Ubuntu.

The E2E job checks out official `godotengine/godot-demo-projects` commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`, builds `deadweight.gdt 0.2.0`, installs no Godot executable, and requires exactly:

```text
CORPUS_COMMIT 0db80ca5fd22b9a40e05b9bc1e00af867fb7c712
MAIN_SCENES 139
COMPLETE 103
PARTIAL 16
UNSUPPORTED_FORMAT_4 9
UNSUPPORTED_UID_ROOT 11
UNEXPECTED_FATAL 0
```

An unexecuted, cancelled, stale-head, or infrastructure-failed job is not successful release evidence.

## Pre-tag ref and merge checks

Before creating the public ref:

```bash
git fetch origin main --tags
git status --short --branch
git rev-parse v0.1.0^{}
git rev-parse v0.1.1^{}
git tag --list v0.2.0
git ls-remote --exit-code --tags origin refs/tags/v0.2.0 refs/tags/v0.2.0^{}
```

The local tag listing must be empty. The remote lookup must return status `2`; any existing `v0.2.0` ref is a hard stop. Verify PR #65 through the GitHub connector, mark it ready only after the archived change and every hosted job pass, and squash-merge it. Then fast-forward local `main`, require a clean worktree, and record the exact connector-reported merge SHA. That merge SHA is the only valid tag target.

## Create and verify the tag

```bash
git tag -a v0.2.0 -m "deadweight.gdt v0.2.0" <verified-merge-sha>
git push origin refs/tags/v0.2.0
git rev-parse v0.2.0^{}
git ls-remote --tags origin refs/tags/v0.2.0 refs/tags/v0.2.0^{}
```

The local and remote peeled annotated-tag targets must equal `<verified-merge-sha>`. Never move, delete, or replace `v0.2.0`; a correction requires a later semantic version.

## Publish and install verification

Create a non-draft, non-prerelease GitHub Release through the GitHub connector for the existing `v0.2.0` tag, with no binary assets. Its notes summarize the six explainability/automation slices, 0.1 compatibility, all quality gates, exact official-demo counts, standalone operation, and deferred boundaries.

Verify a clean tagged module installation outside the repository:

```bash
release_verify_gobin=$(mktemp -d)
GOBIN="$release_verify_gobin" go install github.com/stfulldev/deadweight.gdt/cmd/deadweight.gdt@v0.2.0
"$release_verify_gobin/deadweight.gdt" --version
```

Expected output:

```text
deadweight.gdt 0.2.0
```

## Final cleanup and evidence

- Confirm issue #56 closed automatically through PR #65.
- Record final hosted run/job links, merge SHA, tag object/peeled SHA, GitHub Release URL, and tagged-install output in PR #65.
- Update every remaining tracker #57 checkbox, including phase 3 and the exact-release row, then close tracker #57 as completed.
- Delete merged remote branch `release/56-v0.2.0` only after all public release evidence is preserved.
- Return the local repository to a clean `main` synchronized with `origin/main`.
