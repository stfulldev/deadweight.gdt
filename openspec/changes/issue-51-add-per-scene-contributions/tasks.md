## 1. Analysis Contribution Model

- [x] 1.1 Add stable contribution kinds, context, direct additive values, optional depth candidate, row reliability, and unique-evidence/referrer models; verify model validation rejects invalid identities, kinds, reliability, negative values, and absent required context.
- [x] 1.2 Add deep clone and deterministic compaction/sort helpers for contributions and unique referrers; verify caller mutation cannot alter cached or separately returned results.
- [x] 1.3 Derive each parsed scene's one-occurrence self contribution from its local literal nodes and known depths without copying nested aggregate totals; verify focused root-only and local mixed-node tests reconcile direct values.

## 2. Recursive Attribution and Unique Evidence

- [x] 2.1 Apply resolved child rows at each immediate mount context with checked occurrence multiplication, incoming scene-instance attribution, and root-relative depth composition; verify chain, repeated-instance, and diamond tests reuse cached work and reconcile every additive root metric.
- [x] 2.2 Create unresolved contribution rows with one known node and scene instance per occurrence and apply inherited base/derived rows without a false scene-instance edge; verify imported, unavailable, inherited, override, and unknown-parent cases retain lower-bound or approximate reliability.
- [x] 2.3 Capture external-resource declaration referrers and graph dependency referrers, compact each unique identity once, and keep shared targets ownerless; verify shared-resource and diamond-dependency evidence matches authoritative union cardinalities.
- [x] 2.4 Validate additive sums, complete-depth maxima, unique cardinalities, checked overflow behavior, and owned result publication before returning `RecursiveResult`; verify invariant failures and overflow return fatal `SB2004` or deterministic internal errors with no usable partial result.

## 3. Inspect Selection and Text Presentation

- [x] 3.1 Add paired inspect-only `--metric` and signed-int64 `--top` parsing at the CLI boundary; verify missing pairs, non-positive/overflowing limits, unknown metrics, and unique-union metrics fail before the injected application is called.
- [x] 3.2 Add a non-mutating portable top-contributor projection for additive metrics and known depth candidates with the specified total tie-break order; verify tied rows, absent depth candidates, and limits larger or smaller than the eligible collection are deterministic.
- [x] 3.3 Append the opt-in text section with value markers, occurrences, reliability, portable scene identity, and mount context; verify existing inspect goldens remain byte-identical without selectors and new exact/partial/approximate top goldens contain no ANSI dependency.

## 4. Version-One JSON Contract

- [x] 4.1 Extend inspect and check JSON wire models with all eight ordered contribution metric entries, portable contexts, row reliability, unique evidence, and optional inspect top selection; verify rendering is deterministic, clone-safe, and excludes canonical paths and OS separators.
- [x] 4.2 Extend `schema/deadweight.gdt.report-v1.schema.json` compatibly with contribution, aggregation, unique-referrer, and top-selection definitions; verify every new inspect/check golden validates and existing version-one fields and kinds remain unchanged.
- [x] 4.3 Add complete, lower-bound, approximate, shared-resource, failed-check, and selected-top JSON goldens plus two-checkout and Windows-style portability coverage; verify repeated renders are byte-identical and preserve exactly one trailing LF.

## 5. Documentation and Delivery

- [x] 5.1 Document contribution semantics, supported top metrics, unique-union limitations, examples, and the row-level reliability boundary in README/CHANGELOG/roadmap material; verify examples match CLI help and preserve the static-analysis disclaimer.
- [x] 5.2 Run `gofmt`, focused package tests, `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and golangci-lint; verify every command succeeds without a shipped runtime dependency change.
- [x] 5.3 Run `openspec validate --all --strict` and `git diff --check`; verify all repository specs and the issue #51 change are valid and the worktree contains no formatting defects.
- [ ] 5.4 Commit implementation and documentation separately from test fixtures/goldens as distinct feature and test commits, push the Draft PR, and verify the remote head contains both auditable commits before OpenSpec archive and hosted CI.
