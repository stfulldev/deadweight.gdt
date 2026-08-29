## Why

The published `v0.1.0` command installs and its core analysis works, but an end-to-end sweep of the current official Godot demo projects exposed fatal parser failures on valid format-3 multiline strings and quoted property names. The same tagged Go install reports `dev`, so users cannot reliably identify the release they installed; GitHub issue [#47](https://github.com/stfulldev/deadweight.gdt/issues/47) and Draft PR [#48](https://github.com/stfulldev/deadweight.gdt/pull/48) track the compatibility hardening.

## What Changes

- Extend the supported format-3 lexer contract to preserve literal LF and CRLF content inside quoted strings while keeping physical source positions correct.
- Extend section-body parsing so identifier and quoted-string property names are both accepted, including unknown quoted properties that the streaming parser skips.
- Preserve typed `SB2001` failures for malformed strings and preserve all existing recognized-property semantics.
- Resolve the displayed version from semantic Go module build metadata when no explicit release-time linker override is supplied; keep explicit `-ldflags` injection authoritative and local untagged builds visibly developmental.
- Add focused parser and version-provenance regression tests plus a repeatable official demo-project E2E verification record.

Goals:

- Make supported Godot format-3 scenes from real projects analyzable without requiring Godot or preprocessing.
- Make installed release identity truthful across the documented tagged `go install` path and the source-build linker path.
- Preserve the streaming parser architecture, deterministic output, and existing unsupported-input boundaries.

Non-goals:

- Supporting Godot `format=4`, resolving `uid://` root inputs, or changing the frozen format-3 product boundary.
- Adding project-wide scanning, main-scene discovery, runtime-created scene discovery, or Godot integration.
- Expanding imported/binary scene semantics or changing COMPLETE/PARTIAL policy.
- Moving the immutable published `v0.1.0` tag or publishing a corrective release tag inside this code PR; the fixed provenance applies to fixed-source builds and future releases.
- Adding Node.js, OpenSpec, network access, or another shipped runtime dependency.

Compatibility impact: additive for valid format-3 inputs and corrected for tagged version output. Existing valid identifier properties, explicit linker version injection, local development labeling, metrics, diagnostics, configuration, and exit codes remain compatible.

Affected MVP acceptance criteria: §30 parser/diagnostic coverage, CLI version availability, cross-platform quality gates, and release installation evidence; the change clarifies §12.4's rule that a newline terminates a property value only outside a string.

## Capabilities

### New Capabilities

- `format3-scene-parsing`: Defines lexical and section-body support for real-world Godot format-3 multiline strings and quoted property names, including deterministic positions and typed malformed-input failures.

### Modified Capabilities

- `repository-foundation`: Requires truthful version provenance for explicit linker builds, tagged Go module installs, and untagged development builds without adding runtime services.

## Impact

- `internal/tscn` lexer, parser, and focused regression fixtures/tests.
- `cmd/deadweight.gdt` build-version resolution and isolated tests.
- OpenSpec capability contracts and E2E validation evidence in PR #48.
- No new Go module dependency, CLI flag, config field, metric, diagnostic code, or runtime service.
