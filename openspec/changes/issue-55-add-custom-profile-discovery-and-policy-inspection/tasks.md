## 1. Project and configuration evidence

- [ ] 1.1 Add scene-free project-context discovery that reuses explicit-project validation and nearest-marker traversal, and verify focused finder tests cover explicit, nested, missing, and invalid working-directory cases
- [ ] 1.2 Preserve explicit `fail_on_partial` presence through strict version-one decoding and cloning, and verify absent, true, and explicit-false decoder tests

## 2. Shared policy explanation

- [ ] 2.1 Add typed policy layers, metadata/budget provenance, inheritance chains, summaries, and owned clone behavior, and verify focused model tests cover every field
- [ ] 2.2 Instrument the existing graph and project-budget merges to produce list and explain results without a duplicate resolver, and verify focused tests cover built-in/custom parents, defaults, overrides, unknown profiles, invalid unselected profiles, cycles, and ordinary-resolution parity

## 3. Application and CLI flows

- [ ] 3.1 Add project/config-only application requests and list/show results with injected dependencies, and verify application tests prove missing-context behavior and no scene resolution, analysis, or budget evaluation
- [ ] 3.2 Add `profiles` and `profiles show <id>` commands with validated `--format text|json`, and verify command tests cover request translation, argument/format failures before application work, streams, and exit codes

## 4. Deterministic presentation

- [ ] 4.1 Add deterministic profile list/show text renderers with canonical fields, metrics, sources, chains, empty states, and trailing-LF behavior, and verify focused golden tests and repeated-render ownership
- [ ] 4.2 Add schema-version-one `profiles` and `profile` JSON payloads plus portable configuration context, and verify JSON goldens, schema validation, checkout independence, kind exclusion, and compatibility of earlier version-one documents

## 5. Integration and delivery

- [ ] 5.1 Add end-to-end CLI fixtures for absent/invalid configuration, cycles, custom parents, built-in parents, top-level overrides, and text/JSON output, and verify effective values match an equivalent `check --profile` invocation
- [ ] 5.2 Update user-facing CLI documentation for the additive commands and verify examples and compatibility statements match the implemented syntax
- [ ] 5.3 Run `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint run`, and strict OpenSpec validation, fixing every failure before delivery
- [ ] 5.4 Archive the completed OpenSpec change, verify main specs and archive validation, update linked Draft PR #64, and merge it only after hosted CI passes
