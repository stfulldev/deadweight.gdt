## Context

See `proposal.md` for motivation. The current streaming lexer rejects every physical line ending while reading a quoted token, even though the scanner already tracks LF, CRLF, and CR positions correctly. The section-body parser accepts only identifier tokens as property names. Separately, the composition root always passes the linker variable's default `dev` value and never consults the Go build metadata embedded by versioned module builds.

The change crosses `internal/tscn`, the `cmd/deadweight.gdt` composition root, and contributor-only E2E validation. It must preserve the frozen format-3 boundary, stable `SB2001` failures, and a shipped binary with only Go runtime dependencies.

## Goals / Non-Goals

**Goals:**

- Extend the existing lexer/parser state machines rather than add preprocessing or a second parser path.
- Keep literal string content and post-string source positions correct for all three physical line-ending forms.
- Isolate version precedence in deterministic, directly testable logic at the executable boundary.
- Make the official demo-project sweep repeatable against a recorded corpus commit and fail on any unexpected fatal result.

**Non-Goals:**

- Changing AST fields, analysis aggregation, diagnostics taxonomy, project resolution, or COMPLETE/PARTIAL semantics.
- Treating format 4, UID root resolution, dynamic project flows, or imported scenes as parser fixes.
- Rewriting or moving `v0.1.0`; only a future release built from the fixed source can change installed tagged behavior.
- Running the external demo corpus in normal unit tests or adding its contents to this repository.

## Decisions

### 1. Consume physical line endings inside `readString` as string content

`readString` will write literal `\n` and `\r` runes into the current token instead of returning the existing unescaped-newline error. The shared rune consumer remains responsible for line and column accounting; its existing `previousCR` state already counts CRLF as one physical line while retaining both runes. The lexer will not emit `tokenNewline` for line endings consumed inside a string, so section and property state machines continue to see a single scalar token.

Alternatives considered:

- Pre-normalizing files or multiline strings would lose exact content and make source positions depend on a second representation.
- Emitting newline tokens inside strings would leak lexical context into every parser consumer and violate the existing token contract.
- Limiting support to LF would keep Windows-authored CRLF and legacy CR scenes inconsistent despite the scanner already supporting them.

Malformed EOF and escape paths remain unchanged and continue to construct typed, positioned `SB2001` errors.

### 2. Widen only section-body property-name acceptance

`parseSectionBody` will accept `tokenIdentifier` and `tokenString` before `=` and pass either token through the existing recognized-property or balanced-skip flow. Header section names and header attribute names remain identifier-only. This matches the observed Godot serialization without broadly allowing strings in structural positions.

Alternatives considered:

- Converting quoted names to identifiers in the lexer would erase the grammar distinction globally and could make malformed headers appear valid.
- Special-casing names containing `nodes/` would encode one resource type's serialization instead of the general property grammar.

### 3. Resolve build version at the composition root from embedded inputs

The executable keeps the link-injectable `version` variable. A small resolver chooses:

1. an explicit non-`dev` linker value exactly as supplied;
2. otherwise `runtime/debug.ReadBuildInfo().Main.Version` when it is non-empty and not `(devel)`, removing one leading `v`;
3. otherwise `dev`.

The resolver accepts plain values in tests, while only the thin composition root reads Go build metadata. No report or CLI package learns about module metadata, and no file or network lookup occurs at runtime.

Alternatives considered:

- Generating a version source file would complicate `go install module@version` and require release tooling not present in the repository.
- Moving version discovery into the CLI/report layer would mix executable provenance with command orchestration.
- Retagging `v0.1.0` is rejected because published release refs are immutable; a future corrective tag is a separate post-merge operation.

### 4. Keep the official demo corpus external and add an opt-in E2E runner

A contributor PowerShell runner will accept explicit paths to a built `deadweight.gdt` binary and a `godot-demo-projects` checkout. It will enumerate projects with `run/main_scene`, run no-color `inspect`, classify declared UID roots and format-4 scenes as documented unsupported boundaries, summarize COMPLETE/PARTIAL/fatal outcomes, and exit nonzero for any other fatal result. Validation for this change records the corpus commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712` used to reproduce issue #47.

The runner is contributor tooling, is not embedded in the Go binary, does not clone or mutate the corpus, and is not part of the default Go quality gates.

Alternatives considered:

- Copying official demo scenes into repository fixtures would duplicate third-party content and quickly drift.
- Making network checkout part of `go test ./...` would make the core gate slow and non-reproducible.
- Keeping only an ad-hoc shell transcript would make the acceptance matrix difficult to repeat after parser changes.

## Risks / Trade-offs

- [Allowing physical newlines hides a genuinely unclosed string until EOF] → Preserve the opening-position unterminated-string error and add multiline EOF regression coverage.
- [CRLF position accounting regresses while content is retained] → Assert both exact token content and the next token's line/column for LF, CRLF, and CR cases.
- [Quoted names accidentally become valid in headers] → Change only the section-body gate and add negative structural tests where useful.
- [Build metadata contains development or unexpected values] → Accept only non-empty values other than Go's `(devel)` sentinel; keep `dev` as the safe fallback and linker injection as the explicit override.
- [The external demo repository changes after validation] → Record the exact corpus commit in PR evidence and require callers to choose the checkout explicitly.
- [The first parser fixes reveal another valid format-3 construct] → Let the E2E runner fail as an unexpected fatal; reconcile issue/spec/design before expanding implementation scope.

## Migration Plan

1. Land parser, version, tests, and the opt-in E2E runner in Draft PR #48.
2. Run targeted tests, full repository quality gates, strict OpenSpec validation, and the E2E runner against the recorded demo commit.
3. Archive and sync the completed OpenSpec change, then mark the PR ready for review.
4. After merge, a separate verified release operation may publish a corrective tag built from the fixed source; do not move `v0.1.0`.

No data or configuration migration is required. Before a future release, rollback is a normal PR revert. After a future tag is published, correct any provenance defect with another release rather than moving the tag.
