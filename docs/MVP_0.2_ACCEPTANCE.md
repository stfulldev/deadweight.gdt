# MVP 0.2 acceptance evidence

This matrix traces every release-gate criterion in [tracker #57](https://github.com/stfulldev/deadweight.gdt/issues/57) to executable, committed, hosted, or public-release evidence. Test names are stable Go symbols and can be selected with `go test ./... -run '<symbol>'`.

Status meanings:

- **Verified** — repository-controlled evidence passes before merge.
- **Hosted gate** — the named GitHub Actions job must pass on the exact PR head and merge commit.
- **Post-merge gate** — the operation can only be verified after the reviewed PR is merged; its immutable evidence is recorded in tracker #57 and PR #65.

## Release-gate matrix

| # | Tracker #57 criterion | Evidence | Status |
|---:|---|---|---|
| 1 | Default text and 0.1 exit semantics remain compatible | `TestAcceptanceGoldens`; `TestCheckForwardsFlagsAndMapsNonFatalOutcomes`; `TestDiffForwardsNormalizedPolicyAndMapsOutcomes`; `git diff --exit-code v0.1.1 -- internal/cli/testdata/golden/acceptance` | Verified |
| 2 | JSON schema v1 is committed, documented, validated, deterministic, and portable | [`schema/deadweight.gdt.report-v1.schema.json`](../schema/deadweight.gdt.report-v1.schema.json); `TestJSONReportGoldensValidateAgainstCommittedSchema`; `TestJSONSchemaRejectsInvalidKindsEnumsAndVersions`; `TestMVP02CrossFeatureAcceptance`; Linux/macOS/Windows `test` jobs | Verified + hosted gate |
| 3 | Per-scene contributions and dependency tree explain root metrics without false additive attribution | `TestContributionCompactionIsCheckedConservativeAndOwned`; `TestDependencyTreeProjectionIsBoundedAndUsesBackReferences`; `TestMVP02CrossFeatureAcceptance`; [`testdata/projects/mvp-0.2`](../testdata/projects/mvp-0.2/README.md) | Verified |
| 4 | Diffs qualify incomplete evidence and never claim a missing-data improvement | `TestCompareConfidenceAssessmentAndEnforcement`; `TestDiffQualifiedDecreaseUsesUncertainLanguage`; `TestDiffIntegrationOptInOutcomesRenderBeforeExit`; complete [`diff.golden`](../internal/cli/testdata/golden/mvp_0_2/diff.golden) | Verified |
| 5 | Every metric exposes deterministic evidence-grounded confidence | `TestConfidenceReasonMappingCoversStaticEvidenceClasses`; `TestJSONMetricConfidenceIsCompleteAndSchemaRemainsV1Compatible`; complete MVP 0.2 inspect/check/tree goldens | Verified |
| 6 | Custom profile inspection matches `check` policy | `TestProfilesEndToEndAndCheckPolicyParity`; `TestListAndExplainProfilesReuseEffectiveResolutionWithProvenance`; MVP 0.2 `check_profile`, `profiles`, and `profile_shipping` goldens | Verified |
| 7 | Every child issue is closed with implementation and archived OpenSpec evidence | Child-delivery table below; issues #50–#55, PRs #58/#60–#64, and six dated archive directories | Verified |
| 8 | Build, test, race, vet, lint, strict OpenSpec, supported-OS CI, and official-demo E2E pass | [`RELEASE_0.2.0_CHECKLIST.md`](RELEASE_0.2.0_CHECKLIST.md); [`.github/workflows/ci.yml`](../.github/workflows/ci.yml); `lint`, three `test` jobs, and `official-demo-e2e` | Hosted gate |
| 9 | README and CHANGELOG preserve heuristic and static-analysis disclaimers | [`README.md`](../README.md) “What and why”, “Compatibility”, “Complete vs partial”, and “Supported and unsupported Godot inputs”; [`CHANGELOG.md`](../CHANGELOG.md) compatibility/deferred sections | Verified |
| 10 | The verified release commit is tagged and published as `v0.2.0` | Exact-SHA tag, GitHub Release, and tagged `go install` procedure in [`RELEASE_0.2.0_CHECKLIST.md`](RELEASE_0.2.0_CHECKLIST.md); final connector evidence in [PR #65](https://github.com/stfulldev/deadweight.gdt/pull/65) and [tracker #57](https://github.com/stfulldev/deadweight.gdt/issues/57) | Post-merge gate |

## Child-delivery evidence

| Capability | Issue | Merged PR | Main commit | Archived OpenSpec change |
|---|---:|---:|---|---|
| Versioned JSON reports | [#50](https://github.com/stfulldev/deadweight.gdt/issues/50) | [#58](https://github.com/stfulldev/deadweight.gdt/pull/58) | `7fcb9fe` | [`issue-50-add-versioned-json-reports`](../openspec/changes/archive/2026-08-30-issue-50-add-versioned-json-reports) |
| Per-scene contributions | [#51](https://github.com/stfulldev/deadweight.gdt/issues/51) | [#60](https://github.com/stfulldev/deadweight.gdt/pull/60) | `154efdb` | [`issue-51-add-per-scene-contributions`](../openspec/changes/archive/2026-08-30-issue-51-add-per-scene-contributions) |
| Dependency tree | [#52](https://github.com/stfulldev/deadweight.gdt/issues/52) | [#61](https://github.com/stfulldev/deadweight.gdt/pull/61) | `b0222ed` | [`issue-52-add-explainable-scene-dependency-tree`](../openspec/changes/archive/2026-08-30-issue-52-add-explainable-scene-dependency-tree) |
| Portable report diffs | [#53](https://github.com/stfulldev/deadweight.gdt/issues/53) | [#63](https://github.com/stfulldev/deadweight.gdt/pull/63) | `10feb53` | [`issue-53-add-report-baselines-and-deterministic-diffs`](../openspec/changes/archive/2026-08-30-issue-53-add-report-baselines-and-deterministic-diffs) |
| Per-metric confidence | [#54](https://github.com/stfulldev/deadweight.gdt/issues/54) | [#62](https://github.com/stfulldev/deadweight.gdt/pull/62) | `4debda1` | [`issue-54-add-per-metric-confidence`](../openspec/changes/archive/2026-08-30-issue-54-add-per-metric-confidence) |
| Custom profile inspection | [#55](https://github.com/stfulldev/deadweight.gdt/issues/55) | [#64](https://github.com/stfulldev/deadweight.gdt/pull/64) | `6c8ef01` | [`issue-55-add-custom-profile-discovery-and-policy-inspection`](../openspec/changes/archive/2026-08-30-issue-55-add-custom-profile-discovery-and-policy-inspection) |

## Integrated corpus

`TestMVP02CrossFeatureAcceptance` copies the committed project to two different temporary checkout prefixes and drives the production CLI composition for:

- `inspect --metric nodes --top 3 --format json`;
- `tree --format json`;
- `check --profile shipping --format json`;
- `profiles --format json` and `profiles show shipping --format json`;
- `diff --format json` between baseline and candidate reports with the same `res://root.tscn` identity.

It compares each complete output byte-for-byte across prefixes before matching the committed LF-terminated goldens. The existing CI matrix runs this test on Linux, macOS, and Windows.

The separate hosted E2E job checks official `godotengine/godot-demo-projects` commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712` without installing Godot and requires this exact summary:

```text
MAIN_SCENES 139
COMPLETE 103
PARTIAL 16
UNSUPPORTED_FORMAT_4 9
UNSUPPORTED_UID_ROOT 11
UNEXPECTED_FATAL 0
```

The unsupported categories are honest product boundaries, not unexpected failures. Exact run, merge, tag, Release, and tagged-install identities are external events that happen after the release commit is immutable; they are therefore recorded on PR #65 and tracker #57 rather than retroactively changing the tagged document.
