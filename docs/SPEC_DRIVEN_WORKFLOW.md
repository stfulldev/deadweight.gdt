# Spec-driven development workflow

deadweight.gdt uses [OpenSpec](https://openspec.dev/) as a lightweight planning
layer for changes implemented with Codex. GitHub Issues and pull requests remain
the work-tracking and review system.

## Decision

OpenSpec was selected over [GitHub Spec Kit](https://github.com/github/spec-kit)
for this repository.

| Criterion | OpenSpec 1.11.0 | Spec Kit 1.0.1 |
|---|---|---|
| Codex integration | Native repository skills | Native repository skills |
| Default lifecycle | `propose → apply → archive` | `constitution → specify → plan → tasks → implement → converge` |
| Empty Codex scaffold measured during evaluation | 8 files, 6 skills, about 1.2k lines | About 30 files, 10 skills, about 5k lines |
| Existing-project model | Incremental change deltas, brownfield-first | Full structured feature pipeline |
| Fit for deadweight.gdt | Small addition to an existing detailed specification | Duplicates more of the existing specification and GitHub roadmap |

Both projects are capable. OpenSpec is the better fit because deadweight.gdt
already has a detailed MVP contract and issue roadmap. The repository needs a
durable change envelope for Codex, not a second end-to-end planning hierarchy.

## Sources of truth

- [`docs/MVP_0.1_SPEC.md`](MVP_0.1_SPEC.md) is the frozen product and technical
  contract for MVP 0.1.
- [GitHub Issue #24](https://github.com/stfulldev/deadweight.gdt/issues/24) tracks
  delivery of that MVP.
- `openspec/specs/` will contain capability specifications accumulated from
  completed OpenSpec changes.
- `openspec/changes/` contains active proposals, designs, spec deltas, and task
  lists.
- GitHub Issues describe units of work; linked pull requests contain their code
  and OpenSpec artifacts.

Do not migrate the whole MVP specification into OpenSpec. Adopt capability specs
incrementally when a real change needs them. If an OpenSpec change conflicts with
the frozen MVP contract, stop and resolve the product decision explicitly.

## Prerequisites

OpenSpec is a contributor tool only. It is not a dependency of the Go module or
the distributed `deadweight.gdt` binary.

OpenSpec 1.11.0 requires Node.js 20.19 or newer:

```bash
npm install -g @fission-ai/openspec@1.11.0
openspec --version
```

The repository-local Codex skills were generated with OpenSpec 1.11.0. Upgrade
them deliberately in a reviewed pull request; do not silently regenerate them
with an arbitrary newer version.

OpenSpec collects anonymous command-name telemetry by default. To opt out:

```bash
openspec config set telemetry.enabled false
```

Alternatively, set `OPENSPEC_TELEMETRY=0` or `DO_NOT_TRACK=1`.

## Repository integration

The integration consists of:

- `.agents/skills/openspec-*`: generated Codex skills;
- `.agents/skills/.openspec-target`: the selected Codex integration marker;
- `openspec/config.yaml`: project context and artifact/operation rules;
- this workflow document.

Generated skills should be updated with `openspec update`, not edited manually.
Project-specific policy belongs in `openspec/config.yaml` or this document.

## Change workflow

Use one OpenSpec change for one GitHub Issue.

1. Create a GitHub Issue with scope and acceptance criteria.
2. Create a focused branch and linked Draft PR. Include `Closes #<issue>` in the
   PR description.
3. Optionally explore an uncertain idea without creating artifacts:

   ```text
   $openspec-explore
   ```

4. Create the complete planning change. Prefer a name such as
   `issue-42-add-json-report`:

   ```text
   $openspec-propose issue-42-add-json-report
   ```

   This step creates the proposal, spec delta, design, and tasks. It does not
   authorize implementation. Review the artifacts before continuing.

5. In a new Codex request, apply the reviewed change:

   ```text
   $openspec-apply-change issue-42-add-json-report
   ```

   Keep task checkboxes current and update the proposal instead of silently
   changing scope.

6. When requirements or design change during implementation, reconcile the
   artifacts:

   ```text
   $openspec-update-change issue-42-add-json-report
   ```

7. Run targeted tests and the repository quality gates:

   ```bash
   go test ./...
   go test -race ./...
   go vet ./...
   golangci-lint run
   ```

8. Verify every Issue acceptance criterion, then archive the completed change:

   ```text
   $openspec-archive-change issue-42-add-json-report
   ```

   Archiving promotes the spec delta into the capability specs and preserves the
   completed change history. Commit the archive result in the same PR.

9. Mark the PR ready, review it, merge it, and let its closing keyword close the
   Issue.

## Boundaries

- OpenSpec does not replace GitHub Issues, Draft PRs, or code review.
- An OpenSpec task list is implementation detail, not a second public backlog.
- Proposal generation and implementation happen in separate Codex requests so
  there is a real review boundary.
- Do not create OpenSpec artifacts for trivial typo-only changes unless the
  change affects requirements or the user explicitly asks for them.
- Do not backfill every existing package or MVP Issue. Add capability specs when
  changes establish or modify observable behavior.
- OpenSpec and Node.js must never become runtime requirements for the Go CLI.

## Maintenance

Check the integration after installing the pinned CLI:

```bash
openspec doctor
openspec list
```

To upgrade:

1. create a maintenance Issue and Draft PR;
2. install the intended OpenSpec version;
3. run `openspec update`;
4. inspect every generated diff and update the version in this document;
5. run `openspec doctor` and the Go quality gates;
6. merge only after the generated workflow remains compatible with Codex and
   this repository's Issue/PR process.
