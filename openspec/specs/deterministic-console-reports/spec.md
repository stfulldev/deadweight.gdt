# deterministic-console-reports Specification

## Purpose

Defines byte-stable human-readable console reports that expose analysis completeness, budget verdicts, preset metadata, diagnostics, and process outcomes without hiding uncertainty or requiring color.

## Requirements

### Requirement: Inspect reports expose complete and partial analysis evidence
The `inspect` command SHALL render the binary version, preferred scene display path, project root, analysis status, accuracy, all eight metrics in frozen groups and order, and coverage counts. Complete exact results SHALL render unmarked integers. Lower-bound results SHALL use a consistent `+` marker and a prominent warning; inherited approximate results SHALL use a consistent `~` marker and a warning that overrides can change values in either direction. A successful partial inspect MUST retain exit code `0`.

#### Scenario: Complete inspect report
- **WHEN** inspect receives a complete exact application result
- **THEN** the report renders all metric and coverage values without reliability markers or a partial warning

#### Scenario: Lower-bound inspect report
- **WHEN** inspect receives a partial lower-bound result with unresolved scene instances
- **THEN** every metric affected by closure uncertainty uses `+`, unresolved evidence is shown, a lower-bound warning names the unresolved occurrence count, and the command exits `0`

#### Scenario: Inherited approximate inspect report
- **WHEN** inspect receives a partial approximate result with inherited-scene evidence
- **THEN** affected metrics use `~` and the warning states that inherited overrides can change values in either direction

### Requirement: Check reports expose policy, comparisons, and final verdict
The `check` command SHALL render scene and analysis evidence, effective preset/profile or override identity, available policy metadata, and every configured comparison in canonical metric order with stable actual, budget, textual result, and positive exceedance delta columns. The final verdict SHALL distinguish passed, failed, and incomplete outcomes. Built-in preset checks MUST include the heuristic performance disclaimer.

#### Scenario: Passing check
- **WHEN** every observed metric is within its inclusive budget and partial analysis is not rejected
- **THEN** each comparison renders `PASS`, the final verdict is `PASSED`, and the command exits `0`

#### Scenario: Failing check
- **WHEN** one or more observed metrics exceed their budgets and partial analysis is not rejected
- **THEN** exceeded comparisons render `FAIL` and a positive delta, the summary reports the exceeded count, and the command exits `1`

#### Scenario: Partial check rejected
- **WHEN** analysis is partial, effective `fail_on_partial=true`, and observed comparisons include any mix of pass and fail results
- **THEN** all comparisons remain visible, the final verdict explains the incomplete evidence and policy rejection, and the command exits `3` even when a budget is exceeded

#### Scenario: Custom profile metadata
- **WHEN** a check uses an effective custom profile
- **THEN** the report identifies it as a profile and renders all available effective renderer, target, quality, status, and stability metadata deterministically

### Requirement: Preset reports remain deterministic and explicit
The `presets` report SHALL retain built-in product order and stable lifecycle metadata. `presets show` SHALL render every budget in canonical metric order with comma-separated integers and the performance disclaimer. Neither report SHALL require project or configuration evidence.

#### Scenario: List built-in presets
- **WHEN** the built-in catalog is rendered
- **THEN** `mobile`, `steam-deck`, and `desktop` appear in product order with renderer, target FPS, quality, heuristic, and experimental context

#### Scenario: Show a built-in preset
- **WHEN** one preset is rendered
- **THEN** its metadata, all eight budgets, and the starting-guardrail disclaimer are present in stable text

### Requirement: Visible evidence is normalized without mutating results
All user-visible integers SHALL use comma thousands separators independent of locale. Metrics SHALL use the frozen metric order, diagnostics SHALL sort by severity, code, display path, line, and resource, and unresolved groups SHALL sort by display path and reason. Rendering MUST operate on owned copies or projections and MUST NOT reorder or mutate application results.

#### Scenario: Unordered diagnostics and unresolved evidence
- **WHEN** a report result contains diagnostics and unresolved items in arbitrary input order
- **THEN** repeated renders are byte-identical in the frozen output order and the original slices remain unchanged

#### Scenario: Large and signed presentation values
- **WHEN** report values or exceedance deltas contain thousands
- **THEN** decimal groups use commas and no OS locale affects the bytes

### Requirement: Color policy is contextual and never semantic
ANSI styling SHALL be enabled only for terminal output when neither `--no-color` nor the presence of `NO_COLOR` disables it. Non-terminal output, redirected output, golden output, and fatal stderr output MUST contain no ANSI. Color MUST NOT be the only signal: all states SHALL retain textual labels such as `PASS`, `FAIL`, `WARNING`, `FAILED`, and `INCOMPLETE`.

#### Scenario: Redirected output
- **WHEN** a report is written to a non-terminal stream
- **THEN** output contains no ANSI escape sequence and remains fully understandable

#### Scenario: NO_COLOR on a terminal
- **WHEN** output is a terminal and `NO_COLOR` is present in the environment
- **THEN** output contains no ANSI escape sequence

#### Scenario: Explicit no-color on a terminal
- **WHEN** output is a terminal and `--no-color` is set
- **THEN** output contains no ANSI escape sequence

#### Scenario: Color-enabled terminal
- **WHEN** output is a terminal, `NO_COLOR` is absent, and `--no-color` is not set
- **THEN** status emphasis may contain ANSI while every status remains textually explicit

### Requirement: Fatal errors and exit outcomes are deterministic
Fatal usage, project, configuration, parse, cycle, and internal errors SHALL render once on stderr with an `ERROR` prefix, a stable diagnostic code when one exists, actionable text, and no Go stack trace. Multiline coded evidence SHALL use deterministic indentation. Outcome priority MUST remain fatal `2`, rejected partial `3`, exceeded budget `1`, then success `0`, and reports for non-fatal codes `1` and `3` MUST be emitted before exit.

#### Scenario: Coded cycle error
- **WHEN** analysis returns a coded scene dependency cycle with a display-path chain
- **THEN** stderr renders one code, one heading, and the complete indented chain with exit code `2`

#### Scenario: Uncoded fatal error
- **WHEN** a fatal error has no diagnostic code
- **THEN** stderr renders one plain `ERROR` line without a stack trace and the command exits `2`

#### Scenario: Report-first non-fatal failure
- **WHEN** check produces a failed or incomplete result
- **THEN** stdout contains the complete report before the centralized runtime returns exit code `1` or `3`

### Requirement: Inspect can append an opt-in top-contributors section
When valid top-contributor selectors are present, the inspect text report SHALL append a section after the existing analysis evidence and warnings that names the selected metric and requested limit and renders the deterministically ranked rows with value, occurrence count, reliability, portable scene identity, and mount context. The section SHALL use the selected metric's reliability marker semantics, SHALL remain understandable without color, and MUST NOT alter any existing report bytes before the appended section. When selectors are absent, the complete default text report MUST remain byte-compatible.

#### Scenario: Default inspect remains unchanged
- **WHEN** inspect is rendered without top-contributor selectors
- **THEN** its text bytes match the established report for the same application result

#### Scenario: Partial top row
- **WHEN** a selected contribution is lower-bound or approximate
- **THEN** its value and textual reliability identify that uncertainty without relying on ANSI color

#### Scenario: Stable top ordering
- **WHEN** tied contribution rows are rendered repeatedly from equivalent checkouts
- **THEN** the appended section is byte-identical and uses only portable identities and deterministic context order

### Requirement: Tree text explains dependencies without unbounded expansion
The dependency-tree text report SHALL render the tool version, portable root scene, project, analysis status, and accuracy followed by a rooted tree whose deterministic connector indentation exposes every edge's kind, occurrence count, target identity, reliability, and back-reference or unresolved classification when applicable. Resolved children SHALL appear as branches, repeated compacted edges SHALL show multiplicity, and later diamond paths SHALL use an explicit back-reference marker instead of duplicating an expanded subtree. The report SHALL retain grouped diagnostics and the established partial or approximate warning semantics after the tree.

#### Scenario: Complete chain
- **WHEN** a complete graph contains root A, instance child B, and inherited base C
- **THEN** text renders A as the root followed by nested instance and inheritance branches with their multiplicities and reliability

#### Scenario: Diamond and repeated branches
- **WHEN** a graph contains a repeated edge and a diamond target
- **THEN** text renders the checked repeated count once and marks the later diamond target as a back-reference
- **AND** no reachable edge disappears or expands without bound

#### Scenario: Partial unresolved branch
- **WHEN** a tree contains an imported or unavailable target
- **THEN** text names its portable target, stable classification, and non-exact reliability without relying on ANSI color

### Requirement: Tree text is portable, deterministic, and non-mutating
Tree rendering SHALL sort from a caller-owned projection using portable identities and stable edge context, SHALL use forward-slash resource identities, and MUST NOT expose canonical absolute paths, OS-specific separators, map order, locale-formatted values, or ANSI-dependent meaning. Repeated renders MUST be byte-identical, use exactly one trailing LF, and MUST NOT mutate the recursive result, graph, diagnostics, or contribution evidence.

#### Scenario: Equivalent checkouts
- **WHEN** equivalent tree results use different canonical checkout prefixes and Windows-style internal paths
- **THEN** their text output is byte-identical after holding the tool version and portable project identity constant

#### Scenario: Caller-owned graph
- **WHEN** text tree presentation receives graph edges in caller-owned storage
- **THEN** rendering leaves the original node, edge, diagnostic, and contribution order and values unchanged

### Requirement: Console metrics use their own confidence
Inspect metrics, check comparison actuals, and selected contribution values SHALL use the reliability marker of the rendered metric rather than blindly applying the report-wide or row-wide summary. When one or more metric classifications materially differ from the report summary, text output SHALL add one concise deterministic qualification section listing only those differing metrics, their confidence, and machine-readable reasons in canonical order. When every metric matches the summary, no redundant per-metric section SHALL be added.

#### Scenario: One uncertain resource metric
- **WHEN** only `external_resources` is lower-bound and the report summary is lower-bound
- **THEN** the seven exact metric values are unmarked, the resource value uses `+`, and the qualification section identifies the exact metrics that differ from the summary

#### Scenario: Uniform unresolved closure
- **WHEN** all eight metrics are lower-bound for the same unresolved scene closure
- **THEN** all values use `+` and no redundant mixed-confidence section is emitted

#### Scenario: Contribution metric differs from row summary
- **WHEN** a selected contribution has lower-bound row reliability but its selected additive metric is exact
- **THEN** the selected value is unmarked while the row's conservative reliability remains visible
