## 1. Cross-feature release corpus

- [x] 1.1 Add the committed `testdata/projects/mvp-0.2` baseline/candidate project, nested-scene, resource, and custom-profile fixture, and verify every fixture scene parses through the real CLI.
- [x] 1.2 Add a real-CLI integration test and complete schema-v1 JSON goldens for contributors, tree, profile checks and inspection, confidence, and diff, and verify a second checkout prefix produces byte-identical output.
- [x] 1.3 Prove the frozen MVP 0.1 text contract remains unchanged by verifying the acceptance golden directory has no diff from tag `v0.1.1` and its existing tests still pass.

## 2. Cross-platform and external validation

- [x] 2.1 Extend the hosted workflow with a pinned official-demo E2E job that installs no Godot runtime and verifies 139 main scenes with zero unexpected fatal results.
- [ ] 2.2 Run the integrated corpus in the existing Linux, macOS, and Windows matrix and the pinned demo sweep on hosted Ubuntu, and record successful job evidence in the release artifacts.

## 3. Acceptance and release documentation

- [x] 3.1 Add `docs/MVP_0.2_ACCEPTANCE.md` mapping every tracker #57 release criterion to stable tests, artifacts, child issues and PRs, archived OpenSpec changes, hosted jobs, or post-merge release evidence.
- [x] 3.2 Update README installation, current-version, compatibility, roadmap, and checklist guidance for `v0.2.0`, and verify shipped and deferred boundaries remain explicit.
- [x] 3.3 Promote the changelog's Unreleased entries into a complete dated `0.2.0` section covering issues #50–#55, compatibility, validation, and deferred work.
- [x] 3.4 Add `docs/RELEASE_0.2.0_CHECKLIST.md` with immutable dependency links, exact local and hosted gates, expected benchmark/demo evidence, and guarded merge/tag/Release/tagged-install procedures.

## 4. Verification and delivery

- [ ] 4.1 Run build, unit/integration tests, race detection, vet, exact CI-version lint, strict OpenSpec validation, benchmarks, and compatibility-diff checks successfully.
- [ ] 4.2 Archive the completed OpenSpec change, verify every hosted PR job, merge PR #65, publish annotated tag and GitHub Release `v0.2.0` from the exact merge SHA, verify tagged installation, then close tracker #57 and remove the merged branch.
