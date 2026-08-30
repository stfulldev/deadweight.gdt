## Why

The pinned official Godot demo corpus contains nine main scenes that the `v0.2.0` standalone analyzer rejects only because their text-scene header declares `format=4`, even though the parser already understands the structural data needed by the frozen metrics. GitHub issue [#67](https://github.com/stfulldev/deadweight.gdt/issues/67) and Draft PR [#73](https://github.com/stfulldev/deadweight.gdt/pull/73) begin MVP 0.3 by making that current Godot text-scene boundary explicit, tested, and compatible with format 3.

## What Changes

- Accept Godot text-scene headers with `format=3` or `format=4` for root, nested, and inherited `.tscn` inputs; continue rejecting older and unknown future formats as typed fatal parse failures.
- Recognize format 4's base64-encoded `PackedByteArray` and `PackedVector4Array` value forms through the existing balanced streaming subset without decoding them or adding them to the minimal document AST.
- Preserve format-3 parsing, metric definitions, completeness rules, diagnostics, default text output, JSON schema v1, exit codes, and standalone operation without Godot.
- Move the pinned official-corpus format-4 roots from the unsupported-format bucket into their honest `COMPLETE` or `PARTIAL` analysis outcomes and freeze the new summary in CI evidence.
- Update user-facing compatibility documentation and tests so support is described as Godot 4 format-3 and format-4 text scenes rather than format 3 alone.

Goals:

- Analyze the nine currently blocked format-4 main scenes and any supported format-4 `.tscn` dependencies through the normal recursive pipeline.
- Keep the implementation bounded to the minimal scene data required by the existing eight metrics.
- Preserve deterministic, source-positioned `SB2001` failures for malformed supported files and unsupported format versions.

Non-goals:

- Resolving UID-only paths, merging inherited overrides, invoking Godot, expanding imported or binary scenes, or inferring custom class hierarchies.
- Decoding or semantically inspecting packed array payloads for new metrics.
- Supporting Godot 3 `format=2`, unknown future text formats, `.tres` roots, or project-wide scans.
- Changing presets, budgets, report schema, report diff semantics, or runtime dependencies.

Compatibility impact: additive for valid format-4 text scenes. Existing valid format-3 analysis and presentation contracts remain compatible; diagnostics for unsupported versions will identify the supported set as formats 3 and 4.

Affected MVP acceptance criteria: supported-root and recursive-scene compatibility, inherited-base loading, fatal-versus-partial boundaries, parser fixture coverage, supported-OS quality gates, and the pinned official-demo corpus gate under parent tracker [#66](https://github.com/stfulldev/deadweight.gdt/issues/66).

## Capabilities

### New Capabilities

- `format4-scene-parsing`: Defines the accepted format-4 header and opaque packed-value forms required for deterministic standalone analysis.

### Modified Capabilities

- `analysis-completeness`: Treats successfully resolved supported format-3 and format-4 text-scene closures consistently while preserving fatal unsupported-format boundaries.
- `recursive-scene-expansion`: Expands supported format-4 `.tscn` dependencies through the existing canonical graph and cache rules.
- `inherited-scene-analysis`: Allows a supported format-4 base to participate in the existing deliberately approximate inherited-scene behavior.
- `repository-foundation`: Updates the pinned external-corpus compatibility gate so format-4 roots must be analyzed rather than classified as documented unsupported inputs.

## Impact

- `internal/tscn` format validation and focused format-4 fixtures, parser tests, and benchmarks.
- `internal/analysis` and CLI integration tests proving root, nested, and inherited format-4 behavior.
- `scripts/e2e-godot-demo-projects.ps1`, `.github/workflows/ci.yml`, release acceptance evidence, README, and changelog compatibility wording.
- OpenSpec deltas for the new parser capability and four existing analysis/repository capabilities.
- No new Go module dependency, CLI flag, config field, metric, diagnostic code, schema version, runtime service, or Godot requirement.
