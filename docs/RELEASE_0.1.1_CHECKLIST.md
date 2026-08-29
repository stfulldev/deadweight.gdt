# v0.1.1 release checklist

This checklist is the operator procedure and repository-controlled evidence plan for `deadweight.gdt` `v0.1.1`. The tag must identify the verified merge commit from [PR #59](https://github.com/stfulldev/deadweight.gdt/pull/59); it must never identify the release branch head.

## Scope and immutable boundary

- [x] [Issue #47](https://github.com/stfulldev/deadweight.gdt/issues/47) and [PR #48](https://github.com/stfulldev/deadweight.gdt/pull/48) contain the parser, version-provenance, focused-test, and official-demo E2E evidence.
- [x] The published `v0.1.0` annotated tag remains immutable and resolves to its original MVP merge commit.
- [x] The release contains no MVP 0.2 feature, metric/configuration/preset change, Godot dependency, or prebuilt binary promise.
- [x] Godot format 4, UID-only root resolution, imported-scene expansion, and project-wide scanning remain explicit non-goals.

## Public documentation

- [x] README identifies `v0.1.1` as current and uses it in tagged install/source-build examples.
- [x] CHANGELOG has a dated `0.1.1` entry for multiline strings, quoted property names, module-derived versions, E2E counts, and unchanged unsupported boundaries.
- [x] Frozen `0.1.0` preset values and analysis semantics remain labeled as such.
- [x] Installation remains Go/source based; the GitHub release has no promised binary archives.

## Repository-controlled verification

Run from the clean exact PR head:

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0 run
go test ./internal/tscn -run '^$' -bench '^BenchmarkParseRepresentativeScene$' -benchtime=100ms
go test ./internal/analysis -run '^$' -bench '^BenchmarkRecursiveRepeatedScene100$' -benchtime=100ms
openspec validate --all --strict
git diff --check
```

Every command must pass before PR #59 is marked ready. GitHub Actions must also complete lint and Linux/macOS/Windows build, test, race, and vet jobs successfully on the exact head.

The external official-demo E2E result inherited from PR #48 is pinned to corpus commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`:

```text
MAIN_SCENES 139
COMPLETE 103
PARTIAL 16
UNSUPPORTED_FORMAT_4 9
UNSUPPORTED_UID_ROOT 11
UNEXPECTED_FATAL 0
```

## Pre-tag ref and merge checks

Before creating the public ref:

```bash
git fetch origin main --tags
git status --short --branch
git tag --list v0.1.0 v0.1.1
git rev-list -n 1 v0.1.0
git ls-remote --exit-code --tags origin refs/tags/v0.1.1
```

The remote `v0.1.1` lookup must return status `2` before release. Verify PR #59 through the GitHub connector, mark it ready only after all checks pass, and squash-merge it. Then fast-forward local `main` and record the exact merge SHA; that merge SHA is the only valid tag target.

## Create and verify the tag

```bash
git tag -a v0.1.1 -m "deadweight.gdt v0.1.1" <verified-merge-sha>
git push origin refs/tags/v0.1.1
git rev-list -n 1 v0.1.1
git ls-remote --tags origin refs/tags/v0.1.1 refs/tags/v0.1.1^{}
```

The local and remote peeled annotated-tag targets must equal `<verified-merge-sha>`. Never move or replace `v0.1.1`; a correction requires a later version.

## Publish and install verification

Publish a non-prerelease GitHub Release for the existing `v0.1.1` tag with no binary assets. Its notes must summarize the two parser fixes, tagged-install version provenance, validation matrix, and unchanged unsupported boundaries.

Verify a clean tagged module installation outside the repository:

```bash
<temporary GOBIN> go install github.com/stfulldev/deadweight.gdt/cmd/deadweight.gdt@v0.1.1
<temporary GOBIN>/deadweight.gdt --version
```

Expected output:

```text
deadweight.gdt 0.1.1
```

## Final cleanup and evidence

- Confirm issue #49 closed automatically through PR #59.
- Record the merge SHA, tag object/peeled SHA, hosted CI run, release URL, and tagged-install output in PR #59.
- Delete merged remote branches `release/49-v0.1.1` and `fix/47-real-world-format3-compatibility` only after all evidence is preserved.
- Return the local repository to a clean `main` synchronized with `origin/main`.
