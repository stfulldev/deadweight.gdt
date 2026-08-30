## Context

See `proposal.md` for motivation. The parser currently accepts only `format=3` in `applySceneHeader`, while all later parsing and analysis layers operate on a version-neutral minimal `Document`. Godot's format-4 text resource revision adds base64-string `PackedByteArray` values and `PackedVector4Array`; the parser already tokenizes strings and skips unknown balanced constructors without constructing a Variant AST. The pinned official-demo gate nevertheless short-circuits nine format-4 roots into a documented unsupported category before their actual scene contents are analyzed.

The change crosses the parser version gate, recursive/inherited integration evidence, external-corpus tooling, CI expectations, and compatibility documentation. It must not broaden into UID resolution or change the frozen eight metrics.

## Goals / Non-Goals

**Goals:**

- Make supported format versions an explicit small set rather than an upper-bound comparison.
- Reuse the one streaming parser and one analysis path for formats 3 and 4.
- Exercise both format-4 value additions with minimal committed fixtures and the pinned real-world corpus.
- Prove format-3 behavior remains compatible at parser, analyzer, CLI, and golden boundaries.
- Keep the MVP 0.1 specification as a frozen historical 0.1 contract while updating current-release documentation and OpenSpec capabilities.

**Non-Goals:**

- Building separate AST types, parser modes, or aggregation algorithms by text-scene version.
- Decoding base64 packed arrays, validating their engine-level payload semantics, or retaining node `unique_id` in the MVP document.
- Optimizing arbitrary large-string token allocation unless measurement on the pinned corpus exposes a concrete regression.
- Changing resolution, confidence, metrics, reports, configuration, or shipped dependencies.

## Decisions

### 1. Admit only the explicit set `{3, 4}` at the scene header

The parser will centralize supported text-scene versions and accept exactly 3 and 4. The parsed header retains the declared integer, while missing, non-integer, format 2, and format 5-or-later values remain positioned `SB2001` failures. Error text and focused negative tests will name both supported versions.

Alternatives considered:

- Accepting every version greater than or equal to 3 would silently reinterpret future syntax and undermine the tool's completeness claims.
- Rewriting format 4 to format 3 before parsing would erase provenance and add a preprocessing representation with different source positions.
- A second parser entry point would duplicate an otherwise identical grammar and invite compatibility drift.

### 2. Treat the format-4 packed additions as opaque property values

The existing lexer and balanced-value skipper will consume `PackedByteArray("base64...")` and `PackedVector4Array(...)` wherever the minimal analyzer does not recognize the property. It will enforce ordinary TSCN string, escape, newline, and delimiter structure but will not decode bytes or vectors. The values never enter `Document`, local summaries, or metric evidence.

Minimal fixtures will include a base64 string, a nested packed-vector value, content resembling section delimiters inside the base64 string, a large opaque payload, and malformed string/delimiter cases. This proves that accepting the header is backed by the actual format-4 syntax relevant to the upstream format revision.

Alternatives considered:

- Decoding packed values would add memory and validation work with no consumer in the frozen metrics.
- Adding constructor-specific parser branches would duplicate the generic balanced grammar and imply semantic support the analyzer does not provide.
- Importing full upstream demo scenes into testdata would duplicate large third-party assets; the dedicated pinned-corpus gate is the appropriate real-world evidence.

### 3. Keep node unique IDs outside current analysis identity

The generic section-header parser already accepts scalar attributes it does not project into the minimal AST. Format-4 `unique_id` node attributes will remain in that category. Canonical scene paths, serialized parent/name paths, and existing resource IDs continue to drive cache, graph, contribution, and metric identities.

Alternatives considered:

- Persisting the value now would create an unused public field and prejudge UID or inherited-override designs owned by issues #68 and #69.
- Replacing node paths with unique IDs would change aggregation and report identity far beyond a format-compatibility slice.

### 4. Reuse recursive and inherited behavior without a format branch

Once parsing succeeds, local summaries, dependency graph construction, recursive cache reuse, checked aggregation, completeness, contributions, and reports will receive the same minimal document shape. Integration fixtures will cover a format-4 root, a mixed format-3/format-4 nested chain, a format-4 inherited base using the existing approximate contract, and an unknown future nested format that remains fatal.

Alternatives considered:

- Marking every format-4 result partial solely because of its header would discard exact evidence even when every metric-relevant field is supported.
- Downgrading an unknown future nested format to an unresolved partial target would contradict the existing fatal boundary for resolved malformed or unsupported `.tscn` files.

### 5. Convert the corpus gate from exception accounting to compatibility evidence

The PowerShell runner will stop preclassifying or accepting `UNSUPPORTED_FORMAT_4`. It will run the binary normally, retain the independent UID-root classification, and treat any format-4 rejection as unexpected fatal. After implementation, the exact `COMPLETE`, `PARTIAL`, UID-root, and unexpected-fatal counts from corpus commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712` will replace the 0.2 expectations in CI and issue/PR evidence.

The gate continues to install no Godot executable and performs no network access itself; CI supplies the pinned checkout exactly as before.

Alternatives considered:

- Preserving a zero-valued format-4 category would keep obsolete product terminology in the acceptance contract.
- Checking only the nine known files would miss format-4 dependencies reached by other main scenes and weaken whole-corpus drift detection.

### 6. Preserve release contracts and historical documentation boundaries

Default text bytes for existing format-3 fixtures, JSON schema v1, exit-code meanings, metric IDs, preset values, and configuration remain unchanged. README and changelog will describe current format-3/format-4 support; `docs/MVP_0.1_SPEC.md` remains unchanged because it intentionally freezes what 0.1 promised. New OpenSpec capability deltas become the current normative contract after archive.

Alternatives considered:

- Editing the frozen 0.1 specification would make historical acceptance claims inaccurate.
- Introducing report schema v2 merely to expose the already retained numeric scene format has no user requirement in this slice.

## Risks / Trade-offs

- [A future format-4 construct reaches metric-relevant syntax that the minimal fixtures miss] → Run the entire pinned official corpus and keep any unexpected parse failure fatal to the hosted gate.
- [Large base64 strings increase transient allocations] → Retain token-at-a-time parsing, add a representative large-payload regression/benchmark, and introduce a discard-string optimization only if measured evidence warrants it.
- [Relaxing the version check accidentally accepts future formats] → Use an explicit membership test and negative coverage for both format 2 and format 5.
- [Corpus counts change for reasons unrelated to the nine roots] → Pin the corpus commit, print every category deterministically, and review the measured delta before changing CI expectations.
- [Documentation implies UID support because format-4 files contain UIDs] → State that path-backed resource UIDs remain metadata and UID-only resolution is issue #68.
- [Format-3 output drifts while integration tests focus on format 4] → Run the frozen acceptance goldens and add paired format-3/format-4 equivalence fixtures.

## Migration Plan

1. Add parser fixtures and tests for the explicit version set and format-4 packed-value forms, then update the minimal header gate.
2. Add analyzer/CLI integration coverage for root, nested, inherited-base, equivalence, and future-format failure behavior.
3. Update the external-corpus runner, execute it against the pinned commit, and commit the measured CI expectations with zero unexpected fatal outcomes.
4. Update current compatibility documentation and changelog without modifying the frozen MVP 0.1 specification.
5. Run focused tests, full build/test/race/vet/lint gates, strict OpenSpec validation, and the hosted supported-OS/corpus checks before archive and merge.

No data or configuration migration is required. Before release, rollback is a normal PR revert; after a release tag, any correction requires a later immutable semantic version rather than moving the tag.
