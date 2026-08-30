# report-baselines-and-diffs Specification

## Purpose

Defines portable, deterministic semantic comparison of versioned reports and confidence-aware opt-in regression enforcement for local and CI baseline workflows.

## Requirements

### Requirement: Diff accepts only compatible portable reports
The diff flow SHALL read exactly two complete UTF-8 JSON documents with supported schema version `1`, tool identity `deadweight.gdt`, and kind `inspect`, `tree`, or `check`. Both reports MUST have the same kind and identical portable root scene identity. Every required analysis field, frozen metric, coverage value, diagnostic, dependency identity, and kind-specific evaluation SHALL satisfy its domain and signed-64-bit constraints before comparison. Malformed input, unsupported versions or kinds, duplicate or missing metrics, incompatible kinds or scenes, and invalid semantic evidence MUST fail with exit `2` and no partial diff output. Compatible unknown version-one fields SHALL be ignored.

#### Scenario: Compatible inspect baselines
- **WHEN** two schema-version-one inspect reports describe the same `res://` scene and satisfy the report contract
- **THEN** diff compares them without project discovery, scene parsing, Godot, or network access

#### Scenario: Different root scenes
- **WHEN** the baseline and candidate identify different portable root scenes
- **THEN** diff fails with actionable exit `2` output before producing a semantic comparison

#### Scenario: Unsupported or malformed input
- **WHEN** either file is malformed, has trailing JSON content, uses an unsupported schema version or kind, or omits required semantic evidence
- **THEN** diff fails without panic or partial stdout

### Requirement: Semantic diff covers stable report evidence
Diff SHALL compare all frozen metric values and effective confidence, analysis reliability, coverage fields, grouped diagnostics, unique `scene_dependencies` identities, and budget evaluation when both compatible reports are checks. It SHALL publish only semantic changes: metric and coverage changes use signed `after - before` absolute deltas; diagnostics distinguish added, removed, and occurrence-changed groups; dependencies distinguish added and removed portable identities; and evaluation changes retain before/after verdict, exceeded count, limits, and comparison states in canonical metric order. Tool version, JSON whitespace, object-field order, and other non-semantic envelope metadata MUST NOT create changes.

#### Scenario: Identical semantic reports
- **WHEN** two reports differ only in whitespace, object-field order, or tool version
- **THEN** the result contains no semantic changes and succeeds

#### Scenario: Metrics and evidence change
- **WHEN** metric values, coverage, diagnostic occurrences, dependency identities, and check evaluation differ
- **THEN** every change appears once in deterministic canonical order with exact before, after, and signed delta data where applicable

#### Scenario: One hundred to ninety
- **WHEN** a metric changes from `100` to `90`
- **THEN** its absolute delta is `-10` without a percentage or locale-dependent representation

### Requirement: Metric changes are assessed from confidence evidence
Every changed metric SHALL be assessed as `regression`, `improvement`, or `uncertain` from its before/after confidence. An increase is proven regression only when before is exact and after is exact or lower-bound. A decrease is proven improvement only when after is exact and before is exact or lower-bound. Any changed metric involving approximate evidence, an increase from a non-exact baseline, or a decrease to a non-exact candidate SHALL be uncertain. Older valid version-one metrics without a confidence object SHALL conservatively inherit report-wide reliability and identify `report_summary` as the confidence source.

#### Scenario: Smaller partial candidate
- **WHEN** a candidate lower-bound value is numerically smaller than an exact baseline
- **THEN** the decrease is `uncertain` and is not presented as a proven improvement

#### Scenario: Larger lower-bound candidate
- **WHEN** an exact baseline is smaller than a lower-bound candidate
- **THEN** the increase is a proven regression because the candidate's actual value cannot be smaller than its reported bound

#### Scenario: Legacy partial baseline
- **WHEN** an older schema-version-one metric lacks confidence and its report reliability is lower-bound
- **THEN** comparison uses lower-bound confidence from `report_summary` rather than assuming exact evidence

### Requirement: Regression enforcement is explicit and preserves exits
Regression enforcement SHALL be disabled unless at least one repeatable selected metric is supplied by `--fail-on-increase METRIC` or reliability degradation is selected by `--fail-on-reliability`. A proven increase of a selected metric SHALL produce failed status and exit `1`. An observed selected increase that is confidence-uncertain SHALL produce incomplete status and exit `3`. When reliability enforcement is selected, degradation from exact to lower-bound or approximate, or from lower-bound to approximate, SHALL produce incomplete status and exit `3`. Incomplete SHALL take priority over failed; fatal input remains exit `2`; and every non-fatal outcome MUST render its complete diff before returning its exit code.

#### Scenario: Default comparison reports change only
- **WHEN** metrics increase but no enforcement flags are supplied
- **THEN** diff renders the changes and exits `0`

#### Scenario: Selected exact increase
- **WHEN** an exactly comparable selected metric increases
- **THEN** enforcement reports failed and exits `1`

#### Scenario: Selected uncertain increase
- **WHEN** a selected metric increases but confidence cannot prove regression
- **THEN** enforcement reports incomplete and exits `3`

#### Scenario: Reliability degradation has priority
- **WHEN** reliability enforcement detects degradation and a selected exact metric also regresses
- **THEN** the full diff retains both triggers and exits `3`

### Requirement: Diff results are deterministic, portable, and owned
All collections SHALL be deeply owned and sorted independently of map order. Portable report identities SHALL remain checkout-independent forward-slash resource paths; canonical paths, baseline filesystem paths, locale formatting, ANSI-dependent meaning, and input ordering MUST NOT leak into semantic output. Repeated comparison and rendering SHALL be byte-identical and MUST NOT mutate decoded inputs.

#### Scenario: Equivalent checkout baselines
- **WHEN** semantically equivalent reports were captured from different supported operating systems and checkout locations
- **THEN** diff produces the same semantic result and output bytes

#### Scenario: Caller mutates returned result
- **WHEN** a caller mutates one returned diagnostic or dependency collection
- **THEN** decoded inputs and a repeated comparison retain their original values
