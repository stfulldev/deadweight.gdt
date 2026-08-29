## Context

See `proposal.md` for motivation. The domain packages already expose focused APIs for project discovery and resolution, strict optional configuration loading, recursive analysis, effective policy resolution, partial-policy resolution, budget evaluation, and built-in presets. The remaining CLI commands are placeholders, while preset handlers reach directly into the preset package. The frozen implementation order requires Cobra to orchestrate an application service and display domain results, with centralized exit codes and no Godot dependency.

## Goals / Non-Goals

**Goals:**

- Establish one application boundary that composes existing domain services and returns owned, report-ready inspect/check/preset models.
- Keep Cobra handlers limited to syntax, request translation, presentation dispatch, and non-fatal outcome selection.
- Make both application effects and CLI application calls injectable for deterministic tests.
- Preserve exact fatal versus non-fatal outcome priority while always presenting available non-fatal results.

**Non-Goals:**

- Define the final console layout, ANSI/TTY detection, diagnostic ordering, or golden report fixtures owned by the follow-up report slice.
- Change parser, analysis, configuration, policy, budget, or preset semantics.
- Add concurrency, persistent caches, alternative report formats, or any Godot integration.

## Decisions

### Add an `internal/app` orchestration package

The application package will define requests and report-ready results for inspect, check, preset list, and preset show. A concrete `Application` will sequence the existing domain packages and own no parsing, graph, policy-merge, or comparison logic itself. This keeps `internal/cli` independent of filesystem composition details and gives the next reporting slice stable inputs.

The inspect result will retain the discovered project root, resolved root scene, optional configuration source/presence, and full recursive analysis result. The check result will embed the inspect result and add the effective policy plus budget evaluation. Preset calls will return owned catalog/item values.

Alternative considered: perform orchestration directly in Cobra handlers. Rejected because it couples syntax to filesystem/domain ordering, makes failures harder to test, and contradicts the frozen Step 8 boundary.

### Inject narrow effect functions into the concrete application

`Application` will accept a dependency set for working-directory lookup, project discovery, resolver construction, configuration loading, recursive analysis, policy resolution, partial-policy resolution, budget evaluation, and preset loading. Nil entries will receive production defaults. A small resolver interface will combine root-scene resolution with the analyzer's resource-resolution boundary.

This function-based seam avoids creating service interfaces for pure domain functions and lets application tests record sequence, arguments, short-circuiting, and ownership without the host filesystem. Production defaults will construct the existing secure project resolver and recursive analyzer using `os.Open` and the streaming TSCN parser.

Alternative considered: inject only one high-level analyzer facade. Rejected because it would either hide application sequencing from tests or duplicate policy/config behavior in test fakes. Alternative considered: expose every existing concrete type through global variables. Rejected because mutable globals are unsafe and leak test state.

### Inject one application interface into the command tree

`internal/cli` will define the minimal four-method interface consumed by commands and provide constructors/execution helpers that accept it. The normal `Execute` path will build the production application; tests can supply a fake with in-memory streams. `presets` will use this boundary too, proving it performs no project work.

Global flag state will be allocated per fresh root command. Scene handlers will translate flags into typed application requests. Cobra exact-argument validators and mutually-exclusive flag annotations will reject invalid syntax before the application is called. A string-array flag will preserve repeated budget order.

Alternative considered: keep preset commands directly coupled to `preset.Builtins`. Rejected because it leaves multiple composition paths and weakens the promised command test seam.

### Represent non-fatal command verdicts as an internal exit signal

Handlers will render a successful report first, then return a private typed signal only when a non-zero non-fatal exit is required. The central executor will detect that signal and return its code without printing it as an error. All other errors will continue through the deterministic diagnostic renderer and return `2`. Passed checks and inspect/preset flows return normally with `0`.

The check mapping is `INCOMPLETE → 3`, `FAILED → 1`, and `PASSED → 0`; the budget evaluator already applies partial-before-exceeded priority. Inspect remains `0` for non-fatal partial analysis. This preserves report availability for codes 1 and 3 while preventing domain handlers and `main.go` from duplicating policy.

Alternative considered: add mutable exit state to the root command. Rejected because state can become stale across execution and is harder to compose. Alternative considered: treat codes 1 and 3 as normal errors. Rejected because the executor would incorrectly print a fatal error line after a valid report.

### Use a minimal deterministic presenter until the final report slice

This change will replace placeholder errors with a compact deterministic projection of the report models and keep existing preset rendering. The presenter will accept a color-disabled option but emit no ANSI. The next issue can replace these projections with the complete frozen layout without changing application orchestration or exit policy.

Alternative considered: implement the complete report specification here. Rejected because deterministic formatting, color/TTY policy, sorting, and golden tests are a distinct reviewable issue and would make this orchestration slice too broad.

## Risks / Trade-offs

- [The dependency set has several function fields] → Keep it internal, default every field in one constructor, and test each orchestration edge rather than exposing it as public API.
- [A provisional compact presenter could be mistaken for the final report contract] → Document it as a presentation boundary, avoid ANSI and map iteration, and leave exact text requirements to the next OpenSpec change.
- [Cobra mutual-exclusion errors can depend on framework wording] → Assert actionable flag names and behavior rather than copying a whole upstream sentence.
- [A report result could alias mutable injected data] → Existing domain constructors return owned values; application methods will clone configuration/policy/limits where needed and tests will verify request/result isolation.
- [Explicit config paths are resolved relative to the process working directory by the existing loader] → Preserve this frozen existing behavior and inject working-directory discovery for deterministic application tests.

## Migration Plan

1. Add the application package and focused orchestration tests while retaining existing domain APIs.
2. Replace placeholder commands and route preset handlers through the injected application interface.
3. Update command tests for syntax, request forwarding, report-first outcomes, and no-call usage failures.
4. Run targeted and repository-wide gates; rollback is a branch revert because no persisted data or configuration migration is introduced.
