## Purpose

Defines deterministic inclusive budget comparisons and reliability-aware partial-analysis policy as a pure domain result for later application and reporting layers.

## Requirements

### Requirement: Only configured limits are compared
Each of the eight budget limits SHALL remain optional. Evaluation MUST skip every absent limit and MUST compare every present limit, including zero, as an inclusive upper bound: the comparison passes when `actual <= limit`, fails when `actual > limit`, and records `delta = actual - limit`. Results SHALL contain metric, actual, limit, delta, and pass/fail fields in the frozen order `nodes`, `tree_depth`, `scene_instances`, `mesh_instances`, `lights`, `shadow_lights`, `external_resources`, `scene_dependencies`, independent of input construction order.

#### Scenario: Absent and zero limits
- **WHEN** only `shadow_lights=0` is configured and the observed value is zero
- **THEN** evaluation returns exactly one passing shadow-lights result with delta zero
- **AND** every absent metric is skipped

#### Scenario: Inclusive boundary and first exceedance
- **WHEN** one configured metric equals its limit and another equals its limit plus one
- **THEN** the boundary comparison passes with delta zero
- **AND** the exceeded comparison fails with delta one

#### Scenario: All metrics use canonical order
- **WHEN** all eight limits are configured in any construction order
- **THEN** evaluation returns eight results in the frozen metric order

### Requirement: Evaluation retains analysis reliability
An evaluation SHALL retain exactly one validated reliability value: `exact`, `lower_bound`, or `approximate`. Non-exact evaluation results MUST remain numerically identical to exact comparisons so known lower-bound exceedances still fail and approximate exceedances remain visible. The retained reliability SHALL allow reporting to label non-exact values as observed and distinguish lower-bound `+` from approximate `~`/`FAIL*` semantics without the checker producing console text.

#### Scenario: Lower-bound exceedance remains failed
- **WHEN** lower-bound analysis observes a configured metric above its limit
- **THEN** that comparison remains failed with its positive delta
- **AND** evaluation retains `lower_bound` reliability

#### Scenario: Approximate result remains qualified
- **WHEN** approximate inherited-scene analysis evaluates configured budgets
- **THEN** evaluation retains every comparison and `approximate` reliability for later report qualification

### Requirement: Verdicts summarize budget and partial outcomes
Evaluation SHALL publish exactly one verdict: `PASSED` when no configured budget is exceeded and partial analysis is allowed, `FAILED` when one or more budgets are exceeded and partial analysis is allowed or analysis is exact, and `INCOMPLETE` when analysis is non-exact and effective `fail_on_partial` is true. It SHALL also retain the exact exceeded-count and every comparison. Partial rejection MUST take priority over budget failure, but MUST NOT suppress known failures.

#### Scenario: Complete passing analysis
- **WHEN** exact analysis meets every configured limit
- **THEN** the verdict is `PASSED` and exceeded count is zero

#### Scenario: Complete failed analysis
- **WHEN** exact analysis exceeds three configured limits
- **THEN** the verdict is `FAILED` and exceeded count is three

#### Scenario: Partial allowed with exceedance
- **WHEN** non-exact analysis exceeds a limit and effective `fail_on_partial` is false
- **THEN** the verdict is `FAILED` and the failed comparison remains present

#### Scenario: Partial rejected with exceedance
- **WHEN** non-exact analysis exceeds a limit and effective `fail_on_partial` is true
- **THEN** the verdict is `INCOMPLETE`
- **AND** the failed comparison and exceeded count remain present for observed reporting

### Requirement: Partial override precedence is explicit
The effective partial policy SHALL be resolved from config `fail_on_partial` plus one domain override: `inherit` retains the config value, `fail` forces true, and `allow` forces false. The zero-value override SHALL mean `inherit`, preserving the version-one config default false. An unknown override MUST fail instead of silently inheriting.

#### Scenario: Default remains allowing
- **WHEN** config uses its default false value and no override is supplied
- **THEN** effective `fail_on_partial` is false

#### Scenario: CLI fail overrides config false
- **WHEN** config is false and the domain override is `fail`
- **THEN** effective `fail_on_partial` is true

#### Scenario: CLI allow overrides config true
- **WHEN** config is true and the domain override is `allow`
- **THEN** effective `fail_on_partial` is false

### Requirement: Invalid domain input publishes no evaluation
Evaluation MUST reject negative metric values, negative configured limits, unknown reliability values, and unknown partial overrides. Any such failure SHALL return a zero evaluation rather than ordered results, a verdict, or a partially applied policy. Successful results and slices MUST be owned by the caller and equivalent inputs SHALL produce identical outputs.

#### Scenario: Invalid metric or limit
- **WHEN** any actual metric or configured limit is negative
- **THEN** evaluation fails and returns no usable result

#### Scenario: Invalid reliability or override
- **WHEN** reliability or the partial override is outside its frozen catalog
- **THEN** policy resolution or evaluation fails explicitly

#### Scenario: Caller mutation is isolated
- **WHEN** a caller mutates a returned result slice and evaluates the same inputs again
- **THEN** the later result is unchanged

### Requirement: Budget evaluation is a pure domain operation
The budget evaluator SHALL accept domain metric values, optional limits, analysis reliability, and effective partial policy directly. It MUST NOT discover or read scenes/configuration, access the filesystem or network, invoke Godot, parse CLI flags, choose process exit codes, or render console output.

#### Scenario: Evaluate without runtime services
- **WHEN** a caller supplies in-memory validated domain inputs
- **THEN** evaluation completes without project paths, files, Godot, Cobra, report streams, OpenSpec, or network access
