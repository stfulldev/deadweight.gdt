# MVP 0.1 acceptance evidence

This matrix traces every criterion in [`MVP_0.1_SPEC.md` §30](MVP_0.1_SPEC.md#30-acceptance-criteria-mvp-01) to executable or documented evidence. Test names are stable Go symbols and can be run with `go test ./... -run '<symbol>'` when a focused check is useful.

Status meanings:

- **Verified** — committed automated evidence passes in the repository quality gate.
- **Tracked** — the criterion has explicit delivery evidence in the named follow-up issue and is not claimed complete here.

| # | Acceptance contract | Evidence | Status |
|---:|---|---|---|
| 1 | `inspect` analyzes the simple fixture | `TestDefaultApplicationFixtureMatrix` (absolute simple case), `TestAcceptanceGoldens/inspect_complete`, [`testdata/projects/complete`](../testdata/projects/complete/README.md) | Verified |
| 2 | Absolute, relative, and `res://` root inputs | `TestDefaultApplicationFixtureMatrix`, `TestResolveSceneInputSuccess` | Verified |
| 3 | Nearest `project.godot` wins | `TestFindFilesystemSceneUsesNearestProject`, `TestDefaultApplicationFixtureMatrix` (nearest discovery case) | Verified |
| 4 | Relative external resources resolve from the declaring scene | `TestDefaultApplicationFixtureMatrix` (relative-paths case), `TestRecursiveAnalyzerUsesRealProjectResolverAndParsedScenes`, [`testdata/projects/relative-paths`](../testdata/projects/relative-paths/README.md) | Verified |
| 5 | Repeated scenes parse once and aggregate by occurrence | `TestDefaultApplicationFixtureMatrix` (repeated case), `TestRecursiveAnalyzerAppliesRepeatedSummaryOneHundredTimesAndResetsInvocation`, `TestRecursiveAnalyzerReusesDiamondDescendantAndUnionsEvidence` | Verified |
| 6 | Resolved instance roots are not double-counted | `TestRecursiveAnalyzerExpandsChainFromEachDeclaringScene`, `TestRecursiveAnalyzerAppliesRepeatedSummaryOneHundredTimesAndResetsInvocation` | Verified |
| 7 | Tree depth follows root-depth and mount formulas | `TestBuildLocalSummaryComputesOrdinaryDepthsIndependentOfOrder`, `TestRecursiveAnalyzerExpandsChainFromEachDeclaringScene`, `TestDefaultApplicationFixtureMatrix` | Verified |
| 8 | All eight metrics have exact tests | `TestRecursiveAnalyzerFinalizesSupportedLiteralMetrics`, `TestFinalizeMetricValues`, `TestEvaluateAllMetricsInCanonicalOrder`, `TestDefaultApplicationFixtureMatrix` | Verified |
| 9 | External resources and scene dependencies are unique sets | `TestRecursiveAnalyzerFinalizesCityBuildingLampExample`, `TestFinalizeMetricsUsesRetainedRootEvidenceInCanonicalOrder`, complete nested fixture assertion | Verified |
| 10 | A three-scene cycle exits `2` with its complete chain | `TestDefaultApplicationFixtureFailuresAreCoded`, `TestAcceptanceGoldens/cycle_error`, [`testdata/projects/cyclic`](../testdata/projects/cyclic/README.md) | Verified |
| 11 | Missing/imported nested scenes never claim `COMPLETE` | `TestDefaultApplicationFixtureMatrix` (missing and imported cases), `TestAcceptanceGoldens/inspect_partial`, [`testdata/projects/unresolved`](../testdata/projects/unresolved/README.md) | Verified |
| 12 | Inherited scenes are `PARTIAL approximate` | `TestDefaultApplicationFixtureMatrix` (inherited case), `TestAcceptanceGoldens/inspect_approximate`, [`testdata/projects/inherited`](../testdata/projects/inherited/README.md) | Verified |
| 13 | Complex unknown parser values are skipped safely | `TestParseExtractsSupportedSceneSubset`, `TestParseSkipsLargePackedArrayWithoutBuildingVariantAST`, `FuzzParse` | Verified |
| 14 | Built-in presets support list and show | `TestPresetCommandsUseInjectedApplicationOnly`, `TestPresetsListUsesProductOrderAndLifecycleLabels`, `TestPresetsShowSteamDeck` | Verified |
| 15 | Built-in values match the frozen table | `TestBuiltinsAreFrozenAndOrdered`, `TestBuiltinsLoadFromEmbeddedData` | Verified |
| 16 | Steam Deck checks apply all eight limits | `TestBuiltinsAreFrozenAndOrdered`, `TestEvaluateAllMetricsInCanonicalOrder`, `TestCheckForwardsFlagsAndMapsNonFatalOutcomes` | Verified |
| 17 | Project, profile, and CLI policy merge order is preserved | `TestResolveFourLayerBudgetsAndOwnership`, `TestResolveSelectorPrecedenceAndDomains` | Verified |
| 18 | Configuration/profile errors are actionable exit `2` | `TestResolveValidatesCompleteProfileGraphDeterministically`, `TestDecodeRejectsInvalidJSONAndStaticDeclarations`, `TestAcceptanceGoldens/config_error` | Verified |
| 19 | Budget exceed exits `1` | `TestAcceptanceGoldens/check_fail`, `TestCheckForwardsFlagsAndMapsNonFatalOutcomes` | Verified |
| 20 | Rejected partial analysis exits `3` | `TestAcceptanceGoldens/check_partial_rejected`, `TestEvaluateVerdictAndReliabilityMatrix` | Verified |
| 21 | Partial inspect warns and exits `0` | `TestAcceptanceGoldens/inspect_partial`, `TestDefaultApplicationFixtureMatrix` | Verified |
| 22 | Console output is deterministic and ANSI-free in snapshots | `TestAcceptanceGoldens`, `TestReportGoldens`, `TestColorPolicyUsesTerminalAndBothSuppressionInputs` | Verified |
| 23 | Build, tests, race tests, and vet pass | [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) runs every command on Linux, macOS, and Windows; local commands are listed below | Verified locally; CI runner execution required |
| 24 | CLI neither finds nor requires a Godot executable | `TestDefaultApplicationInspectsAndChecksTextSceneWithoutGodot`, fixture suite and CI contain no Godot setup | Verified |
| 25 | README contains required positioning and disclaimers | [`README.md`](../README.md) sections “What and why”, “Complete vs partial”, “Supported and unsupported Godot inputs”, and “Presets and custom profiles”; GitHub issue [#23](https://github.com/stfulldev/deadweight.gdt/issues/23) | Verified |

## Verification commands

```bash
go build ./...
go test ./...
go test -race ./...
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.0 run
go test ./internal/tscn -run '^$' -fuzz '^FuzzParse$' -fuzztime=5s
```

The fixture corpus is hand-authored text and inert placeholder assets. These commands do not install, locate, launch, or communicate with Godot.
