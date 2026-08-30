## 1. UID Metadata and Index

- [ ] 1.1 Add focused test fixtures for canonical/non-canonical UID text, format-3/format-4 scene and text-resource headers, version-applicable `.uid` sidecars, `.import` remap metadata, and valid/truncated/oversized binary caches; verify the targeted UID metadata tests enumerate every supported source and corruption boundary without Godot.
- [ ] 1.2 Implement the pure streaming UID codec and bounded metadata readers without regex parsing or Variant materialization; verify canonical round trips, checked lengths, source positions, unsupported future forms, and deterministic owned results pass the targeted tests.
- [ ] 1.3 Add project-index tests for hidden/non-hidden project-data directories, cold checkout, direct-over-cache precedence, equivalent claims, duplicate direct claims, stale/cache-only mappings, symlink escapes, deterministic walk order, lazy construction, and repeated lookup reuse; verify every typed outcome is asserted without machine-specific files.
- [ ] 1.4 Implement the lazy invocation-scoped project UID index through injected walk/open/stat/canonicalization effects; verify path-only requests perform zero UID scans and add a representative scan benchmark before deciding whether any in-scope optimization is justified.

## 2. Secure Root and Resource Resolution

- [ ] 2.1 Add finder and root-resolution tests for UID input with explicit project, cwd discovery, missing project, unique scene mapping, wrong extension, malformed/missing/ambiguous UID, stale cache, and outside-project evidence; verify every success preserves original/canonical/display identities and every root failure remains typed and fatal.
- [ ] 2.2 Update project discovery and root resolution to query the lazy index only after project selection; verify all existing absolute, relative, and `res://` finder/resolver tests remain byte-for-byte compatible and UID input never triggers filesystem-scene validation before discovery.
- [ ] 2.3 Add combined UID/path resource-resolution tests covering UID precedence, matching aliases, unknown-UID path fallback, ambiguity/stale/unsafe fallback evidence, UID-only ordinary resources, relative bases, non-regular targets, and caller mutation; verify every mapped path passes existing lexical and symlink containment.
- [ ] 2.4 Implement the combined reference contract while preserving the path-only resolver wrapper; verify new stable UID reasons are valid, deterministic, and independently consumable without parsing messages or changing canonical graph identities.

## 3. Application and Recursive Analysis

- [ ] 3.1 Add application, analyzer, and CLI fixtures for a UID root, UID-only nested scene, UID/path aliases to one child, UID-resolved inherited base, ordinary UID resource, repeated/diamond reuse, unresolvable UID, duplicate UID, and path fallback; verify inspect, check, tree, text, and JSON cover exact, lower-bound, approximate, and fatal outcomes.
- [ ] 3.2 Route retained external-resource UID/path evidence through combined resolution in graph discovery, occurrence expansion, inheritance, closure identities, contributions, coverage, diagnostics, and per-metric confidence; verify one canonical child/cache identity is reused and only unresolved UID evidence produces `SB1006`.
- [ ] 3.3 Run frozen default-text acceptance, JSON schema-v1, configuration, preset, budget, contribution, and tree goldens; verify issue #68 changes no existing path-only bytes, metric values, schema shapes, selector semantics, or process exit meanings.

## 4. Official Corpus and Documentation

- [ ] 4.1 Remove `UNSUPPORTED_UID_ROOT` preclassification from the PowerShell corpus runner, run the standalone binary against pinned corpus commit `0db80ca5fd22b9a40e05b9bc1e00af867fb7c712`, and commit the measured expectation of 139 main scenes, 121 complete, 18 partial, and zero unexpected fatal outcomes; verify all eleven UID roots enter ordinary analysis and any category drift fails hosted `official-demo-e2e`.
- [ ] 4.2 Update README and CHANGELOG current-source compatibility and limitations for standalone UID evidence, fallback, ambiguity, project scans, and unchanged bridge/import/inheritance boundaries; verify historical MVP 0.1/0.2 specifications and release evidence remain unchanged.

## 5. Verification and Delivery

- [ ] 5.1 Run targeted metadata, index, resolver, finder, application, analyzer, CLI, corruption, containment, lazy-effect, corpus, frozen-golden, and JSON-schema tests plus `git diff --check`; verify each new scenario and unchanged compatibility boundary passes deterministically.
- [ ] 5.2 Run `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint run`, and `openspec validate --all --strict`; verify every repository-controlled gate passes without installing or invoking Godot.
- [ ] 5.3 Commit production feature changes separately from tests/fixtures, keep corpus/tooling and documentation evidence separately reviewable where practical, push every commit, and update Draft PR #75 with exact local and hosted gate/corpus evidence.
- [ ] 5.4 Sync and archive `issue-68-resolve-uid-resources-standalone`, commit the archive separately, verify strict OpenSpec status and local/remote head equality, and mark PR #75 ready only after every task and exact-head hosted check passes.
