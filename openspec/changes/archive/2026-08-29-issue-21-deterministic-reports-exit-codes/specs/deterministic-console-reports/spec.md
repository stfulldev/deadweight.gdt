## Purpose

Defines byte-stable human-readable console reports that expose analysis completeness, budget verdicts, preset metadata, diagnostics, and process outcomes without hiding uncertainty or requiring color.

## ADDED Requirements

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
