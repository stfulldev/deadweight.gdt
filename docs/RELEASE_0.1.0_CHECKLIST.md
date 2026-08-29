# v0.1.0 release checklist

This checklist is the release evidence and operator procedure for `deadweight.gdt` `v0.1.0`. The tag must identify the verified merge commit from [PR #46](https://github.com/stfulldev/deadweight.gdt/pull/46); it must never be created from the feature branch.

## Scope and dependencies

- [x] MVP behavior is frozen by [`MVP_0.1_SPEC.md` §36](MVP_0.1_SPEC.md#36-frozen-decisions-для-01).
- [x] License issue [#19](https://github.com/stfulldev/deadweight.gdt/issues/19) is complete and `LICENSE` contains `Copyright (c) 2026 stfulldev`.
- [x] Acceptance/fixture issue [#22](https://github.com/stfulldev/deadweight.gdt/issues/22) is complete.
- [x] [`MVP_0.1_ACCEPTANCE.md`](MVP_0.1_ACCEPTANCE.md) contains one evidence row for every §30 criterion.
- [x] Roadmap `0.2+` features remain documented as non-shipping candidates, not included behavior.

## Public documentation

- [x] README contains all twelve §33.5 sections.
- [x] README describes the shipped commands, flags, eight metrics, strict config, presets/profiles, reliability, supported inputs, exits, CI, roadmap, contribution workflow, and license.
- [x] README states that built-ins are heuristic/experimental and require target-hardware profiling.
- [x] README states that `steam-deck` is not an official Valve certification profile or endorsement.
- [x] README explains runtime-created nodes and imported/binary scene limitations.
- [x] CHANGELOG records the frozen preset values, experimental lifecycle, shipped MVP, and known limitations.
- [x] Go installation is documented; prebuilt release binaries are explicitly future distribution work.

## Repository-controlled verification

Run from a clean exact PR head:

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

Before archiving PR #46, record the exact head SHA and results in its conversation. A failure in any executed repository command blocks the tag.

## Hosted CI evidence

The workflow must define and attempt:

- Linux, macOS, and Windows: `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`.
- Linux lint: golangci-lint v2.12.0.
- No Godot installation or invocation.

A completed job with failing repository steps blocks the tag. A job that has no runner, logs, or steps is an external workflow-start failure, not a green check and not evidence that repository commands failed; preserve its run/job IDs separately instead of rewriting its status.

## Pre-tag ref and merge checks

Before creating the public ref:

```bash
git fetch origin main --tags
git status --short --branch
git tag --list v0.1.0
git ls-remote --exit-code --tags origin refs/tags/v0.1.0
```

Expected before release: clean `main` after a fast-forward pull and no local or remote `v0.1.0`. `git ls-remote --exit-code` must return status `2` when the remote tag is absent.

Verify the merge identity through the GitHub connector and locally:

```bash
git switch main
git pull --ff-only origin main
git rev-parse HEAD
```

`HEAD` must equal PR #46's recorded merge commit, and issue #23 must be complete.

## Create and verify the tag

Only after every preceding release gate is satisfied:

```bash
git tag -a v0.1.0 -m "deadweight.gdt v0.1.0" <verified-merge-sha>
git push origin refs/tags/v0.1.0
git rev-list -n 1 v0.1.0
git ls-remote --tags origin refs/tags/v0.1.0 refs/tags/v0.1.0^{}
```

The peeled annotated-tag target must equal `<verified-merge-sha>`. Never move or replace the public tag silently; corrections use a follow-up version unless the repository owner explicitly authorizes replacement.

## Final tracker check

After the remote tag is verified:

- confirm issue #23 is closed;
- confirm all child issues in [MVP tracker #24](https://github.com/stfulldev/deadweight.gdt/issues/24) are closed;
- add the tag SHA and release evidence summary to #24;
- close #24 as completed.
