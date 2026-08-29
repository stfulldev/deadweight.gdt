## Context

See `proposal.md` for motivation and the two delta specs for observable behavior. The current application boundary already returns owned, report-ready `InspectResult` and `CheckResult` values. `internal/report` validates those values and renders text directly, while `internal/cli` owns Cobra parsing, output streams, and centralized exit codes. Analysis results also retain canonical filesystem identities internally, so encoding domain structs directly would leak implementation fields and checkout-specific paths into a contract that later baseline/diff work must treat as portable.

The change crosses CLI parsing/error routing, report projection/encoding, committed schemas, and golden/integration tests. It must preserve the existing text output, non-fatal report-first behavior, diagnostic taxonomy, and a standalone Go binary with no runtime services.

## Goals / Non-Goals

**Goals:**

- Create an explicit wire model whose compatibility can evolve independently from internal Go domain structs.
- Reuse one semantic projection for inspect evidence inside both inspect and check documents.
- Make stream framing, path portability, ordering, number representation, and fatal behavior directly testable.
- Keep application/domain packages unaware of JSON, Cobra, terminal color, and process streams.
- Reserve additive version-one evolution for later MVP 0.2 document kinds and optional evidence without weakening established fields.

**Non-Goals:**

- Generalizing all report rendering behind a plugin framework or exporting a public Go API.
- Reading JSON reports, comparing baselines, rendering dependency trees, or computing per-scene contributions.
- Refactoring text renderers beyond extracting shared validation/projection helpers needed to prevent semantic drift.
- Adding code generation, a runtime schema validator, or a non-standard-library JSON dependency to the shipped binary.

## Decisions

### 1. Project application results into dedicated version-one wire structs

`internal/report` will own private schema-versioned wire structs and constructors for inspect, check, and fatal documents. Constructors validate the existing application/domain result, clone and deterministically order retained slices, normalize portable display identities, and then encode only explicitly selected fields. The application service interfaces remain unchanged.

Alternatives considered:

- Encoding `application.InspectResult` and `analysis.RecursiveResult` directly would expose canonical absolute paths, internal graph/cache fields, Go field names, and future refactors as accidental API commitments.
- Generic `map[string]any` construction would make required fields, integer invariants, ordering, and schema drift harder to review and test.
- Moving JSON tags into analysis, budget, policy, or diagnostic models would couple pure domain packages to one presentation contract.

### 2. Use one discriminated envelope with kind-specific payloads

Every document begins with numeric `schema_version: 1`, a `kind` discriminator, and `{name, version}` tool metadata. Kind-specific fields follow:

- `inspect`: portable scene/configuration context plus one analysis payload;
- `check`: the same analysis payload plus effective policy and evaluation;
- `error`: one fatal diagnostic payload with optional stable code and structured context.

The committed Draft 2020-12 schema uses conditional validation keyed by `kind`. Required established fields are strict, while object readers are documented to ignore unknown optional fields so later MVP 0.2 slices can add compatible evidence or kinds without changing existing meanings. Incompatible changes require a new integer schema version and schema filename.

Alternatives considered:

- Separate unrelated root schemas would duplicate the shared analysis contract and make discriminator-based readers harder to implement.
- A string tool-version-derived schema version would conflate release cadence with wire compatibility.
- Serializing text output inside JSON would preserve prose parsing rather than create a machine contract.

### 3. Preserve current stream and exit ownership in the CLI layer

The scene commands validate `--format` before application work and pass a small presentation enum only to the presentation boundary. Successful, failed-budget, and rejected-partial reports are fully encoded to stdout before the existing centralized exit signal returns `0`, `1`, or `3`. If JSON mode was successfully selected and a later operation fails, the centralized fatal path encodes one error document to stderr and returns `2`; malformed invocation state that prevents reliable format selection retains the existing text usage error.

The CLI will determine the selected format from parsed command state rather than making application requests format-aware. JSON rendering ignores terminal/color runtime state entirely.

Alternatives considered:

- Writing fatal documents to stdout would mix report data and errors in pipelines and differ from the established CLI stream contract.
- Returning JSON bytes from the application layer would couple domain orchestration to process presentation.
- Changing non-fatal outcomes to exit zero for easier JSON consumption would break documented CI behavior.

### 4. Normalize portable identities at the wire boundary

Successful documents use the existing preferred in-project scene display identity and normalize separators to forward slashes. Canonical absolute project/scene paths and temporary-directory values are not serialized as semantic fields. Configuration records preserve presence and selection provenance (`implicit`, `explicit`, or absent) without embedding a checkout-specific absolute filename. Diagnostic locations use their existing stable display/source identity after the same portable normalization.

The projection fails rather than silently emitting an absolute canonical path where a required portable scene identity cannot be derived. Fatal documents may echo actionable user-supplied input or source context because they are not baseline identities, but they do not expose internal stack traces.

Alternatives considered:

- Emitting both canonical and display paths would invite baseline consumers to depend on non-portable values.
- Replacing paths during golden tests only would make production documents differ across machines.
- Hashing absolute paths would remain checkout-dependent while being less actionable.

### 5. Encode byte-stable JSON with the Go standard library

Encoders use explicit structs and ordered slices, `encoding/json` with HTML escaping disabled, two-space indentation, LF separators, and one trailing LF. Metrics and comparisons follow the canonical metric catalog; diagnostics reuse the frozen sorting rules on owned copies. Domain `int64` values remain JSON integers, including signed comparison deltas. Rendering validates first and buffers the complete document before writing, so an encoding/validation failure cannot leave truncated stdout.

Alternatives considered:

- Streaming fields directly to the destination risks partial documents on late validation or encoding failure.
- Compact JSON is smaller but less reviewable in fixtures and does not materially benefit scene reports at this scale.
- A third-party canonical JSON package would add a shipped dependency without solving a requirement that explicit structs and ordered slices already satisfy.

### 6. Validate producers through one committed schema without runtime schema loading

The repository adds `schema/deadweight.gdt.report-v1.schema.json`. Golden fixtures for every outcome are validated against it in tests, and focused unit tests assert compatibility rules, portable paths, stream framing, and immutability. Production encoding does not read the schema file or require a schema library; typed projection validation remains the runtime guard.

Alternatives considered:

- Runtime schema validation would increase binary size and duplicate typed validation on every invocation.
- Generating Go structs from JSON Schema would introduce generation tooling and make small compatibility reviews harder.
- Testing only Go decoding would not prove the committed public schema matches emitted documents.

## Risks / Trade-offs

- [A field accidentally exposes a canonical absolute path] → use dedicated wire structs, portable projection helpers, cross-checkout golden tests, and a test that rejects known temp-root fragments.
- [JSON and text semantics drift] → construct both from the same authoritative application results and add paired outcome tests for metrics, coverage, diagnostics, policy, comparisons, and exit codes.
- [A fatal error occurs before format selection is reliable] → retain the established text usage path for pre-selection parse failures and document the boundary; once JSON is selected, every later fatal path uses the error envelope.
- [Later MVP 0.2 fields make version one ambiguous] → freeze existing required meanings, permit only additive optional fields/new kinds, and require a new integer version for incompatible changes.
- [Buffering increases peak memory] → reports contain summaries rather than source blobs; keep the wire model bounded by existing analysis evidence and benchmark only if fixtures reveal material growth.
- [JSON numeric consumers lose precision outside Go] → keep the existing signed `int64` contract and document it in the schema; consumers that cannot represent 64-bit integers exactly must use an appropriate integer decoder.

## Migration Plan

1. Land the validated planning artifacts in Draft PR #58 without changing runtime code.
2. Add wire projections, schema, and focused encoder validation while keeping text as the default.
3. Add CLI selection and fatal routing with paired text/JSON exit tests.
4. Add cross-platform golden/schema evidence and documentation, then archive and sync the completed OpenSpec change.
5. Merge only after the full repository gates pass and the PR is ready; no data/config migration is required.

Rollback before release is a normal PR revert. After a release publishes schema version one, correct compatible producer defects without reinterpreting existing fields; incompatible corrections require schema version two.
