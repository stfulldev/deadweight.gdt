## 1. Application and command flow

- [x] 1.1 Factor the shared single-scene analysis work behind `Inspect` and add explicit `TreeRequest`, `TreeResult`, and `Application.Tree` boundaries; verify equivalent discovery, configuration, resolution, recursive evidence, and the absence of policy or budget evaluation with application tests.
- [x] 1.2 Add the injected `tree <scene>` command with text default and `--format text|json`; verify absolute, working-directory-relative, and `res://` request translation, plus argument and format rejection before the application is invoked.
- [x] 1.3 Preserve command exit and stream contracts; verify successful partial trees exit `0`, fatal/usage failures exit `2`, successful output uses stdout, and JSON fatal errors use stderr without a successful tree document.

## 2. Authoritative tree projection

- [x] 2.1 Add owned report-domain tree types and validate root, source, resolved-target, portable-identity, occurrence, and reachability invariants against `analysis.DependencyGraph`; verify root-only and inconsistent injected graphs without filesystem access.
- [x] 2.2 Implement portable sibling ordering and deterministic depth-first preorder with explicit depth and sibling-last metadata; verify equivalent checkout prefixes, shuffled input order, Windows-style internal paths, and repeated rendering.
- [x] 2.3 Emit every compacted graph edge once, expand each resolved target once, stop later paths with `back_reference=true`, and keep unresolved edges as leaves; verify repeated edges, diamonds, and the nodes-plus-edges bound.
- [x] 2.4 Derive exact resolved-instance, lower-bound unresolved-instance, and approximate inheritance row reliability while retaining safe unresolved context; verify imported, missing, placeholder, sub-resource, unavailable, unsupported, overflow, and fatal-cycle boundaries.

## 3. Deterministic text report

- [x] 3.1 Add a tree text renderer over the shared projection with the standard version/root/project/status/accuracy header, UTF-8 connectors, edge kind, checked multiplicity, reliability, target, back-reference, and unresolved classification; verify complete-chain and root-only goldens.
- [x] 3.2 Retain grouped diagnostics and established partial/approximate warnings without semantic ANSI dependence; verify partial, inherited, repeated, and diamond goldens with exactly one trailing LF.
- [x] 3.3 Prove text rendering is byte-stable, portable, and non-mutating; verify canonical checkout paths, backslashes, locale/map ordering, and caller-owned graph, diagnostic, and contribution order never leak or change.

## 4. Version-one JSON and schema

- [x] 4.1 Add a compatible `tree` document kind that reuses the complete portable analysis payload and attaches required `dependency_tree.root` and flat preorder entries; verify the text and JSON renderers consume equivalent projection semantics.
- [x] 4.2 Extend `schema/deadweight.gdt.report-v1.schema.json` with tree definitions and required entry invariants without changing existing inspect, check, or error meanings; verify all established and new fixtures against Draft 2020-12 validation.
- [x] 4.3 Add JSON goldens for complete, repeated, diamond, partial, inherited, root-only, and fatal-cycle cases; verify positive signed-64-bit depth/occurrences, deterministic ordering, ANSI-free output, exactly one trailing LF, and non-mutation.
- [x] 4.4 Prove clone and OS portability with equivalent Unix/Windows canonical inputs in two checkout prefixes; verify byte-identical JSON containing only forward-slash portable identities and no canonical checkout names.

## 5. Documentation and delivery

- [x] 5.1 Document `tree`, its local declaring-edge multiplicity, back-reference, partial-evidence, and JSON contracts in command help, README, and CHANGELOG without implying transitive totals, full inheritance merging, UID/import expansion, or project-wide visualization.
- [x] 5.2 Run formatting, build, all tests, race tests, vet, and lint; record zero failures and preserve the repository's offline/static-analysis runtime boundary.
- [x] 5.3 Run `openspec validate --all --strict`, `openspec status`, and `git diff --check`; reconcile the implemented behavior with every requirement and scenario before archiving the change.
- [ ] 5.4 Keep production/documentation work and test/fixture work in separate commits, push each reviewed stage to linked Draft PR #61, complete the PR checklist, then archive the OpenSpec change only after all local and hosted gates pass.
