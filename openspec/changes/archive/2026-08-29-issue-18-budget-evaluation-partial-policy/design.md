## Context

See `proposal.md` for motivation and `specs/budget-evaluation/spec.md` for behavior. `internal/budget` already owns optional `Limits` and a canonical-order `Check` helper, `internal/metrics` owns validated non-negative metric values, `internal/analysis` owns the exact/lower-bound/approximate reliability taxonomy, and `internal/policy` guarantees a non-empty effective budget before the check flow.

The missing contract is a side-effect-free summary around those comparisons plus a small tri-state override for config `fail_on_partial`. The application and report layers do not exist yet, so this issue must expose sufficient evidence without selecting exits or formatting strings.

## Goals / Non-Goals

**Goals:**

- Extend `internal/budget` additively while preserving existing `Check` callers and result fields.
- Make verdict priority, reliability evidence, and exceeded counts explicit domain data.
- Validate domain inputs at the evaluation boundary and return zero output on failure.
- Represent CLI intent without importing Cobra or flag-specific types.
- Keep output slices owned and deterministic.

**Non-Goals:**

- Re-resolving selectors/budgets, validating config JSON, mapping verdicts to numeric process exits, or producing `Actual`/`Observed` report labels.
- Deciding Cobra mutual-exclusion errors for `--fail-on-partial` plus `--allow-partial`; issue #20 maps its already-validated flags to one domain override.
- Adding warning thresholds, scores, percentages, or per-metric reliability guesses.

## Decisions

### Preserve `Check` and add a validated `Evaluate` boundary

`Check(metrics.Values, Limits) []Result` remains the small comparison primitive and compatibility API. A new `Evaluate(values, limits, reliability, failOnPartial) (Evaluation, error)` validates metrics, limits, and reliability, obtains canonical comparisons from `Check`, counts failures, applies verdict precedence, and returns an owned aggregate.

`Evaluation` records `Status`, `analysis.Reliability`, `FailOnPartial`, `Exceeded`, and `[]Result`. Status values are the product-facing frozen tokens `PASSED`, `FAILED`, and `INCOMPLETE`; their validator accepts only those values. The result slice is always freshly allocated by `Check`, including for an incomplete outcome.

Alternatives considered:

- Change `Check` to return the aggregate and an error: rejected because it would break the focused primitive and existing callers/tests for no gain.
- Put the aggregate in `internal/analysis`: rejected because verdicts depend on policy limits and belong to the budget/check layer, not scene analysis.
- Store only a boolean failed/incomplete pair: rejected because it permits invalid combinations and makes final-priority logic repeat downstream.

### Treat reliability as evaluation-wide evidence

The evaluator imports the existing `analysis.Reliability` type and retains it once on the aggregate. All metric comparisons use the supplied observed numeric values without changing pass/fail for lower-bound or approximate analysis. Reports can therefore choose `Actual` for exact, `Observed` plus `+` for lower-bound, and `Observed` plus `~`/notes for approximate without the budget package knowing console syntax.

Alternatives considered:

- Add reliability to every `Result`: rejected as redundant because one analysis result supplies one reliability class for the whole metric set in MVP 0.1.
- Suppress failures for approximate data: rejected because the frozen contract requires known/observed exceedances to remain visible.

### Resolve partial policy with a tri-state domain override

`PartialOverride` uses a zero value for inherit plus explicit `fail` and `allow` values. `ResolveFailOnPartial(configured, override)` returns the config value for inherit, true for fail, false for allow, and an error for any unknown value. The application layer will convert mutually exclusive flags into this single value before evaluation.

Alternatives considered:

- Pass two CLI booleans into the budget package: rejected because flag parsing and mutual exclusion are command concerns.
- Use `*bool`: rejected because a named enum produces clearer diagnostics and makes the three supported intents discoverable.

### Validate actuals and limits before producing comparisons

`Evaluate` calls `metrics.Values.Validate`, a new `Limits.Validate`, and `Reliability.Valid` before invoking `Check`. `Limits.Validate` visits the frozen metric order and returns a typed error naming the first negative configured limit. Invalid input returns the zero `Evaluation`; it never publishes partial comparisons or a verdict. Empty limits remain a valid vacuous `PASSED` evaluation because the effective-policy resolver already enforces non-empty check policies, while the primitive stays independently usable.

Alternatives considered:

- Trust all callers: rejected because negative constructed values can bypass config/policy decoding in tests or future integrations.
- Reject empty limits again: rejected as duplicate policy-layer enforcement and inconsistent with optional-limit primitives.

### Compute verdict in one explicit priority block

Evaluation first counts failed comparisons. It then selects `INCOMPLETE` when reliability is non-exact and fail-on-partial is true; otherwise it selects `FAILED` when exceeded is positive; otherwise `PASSED`. Comparisons and the exceeded count are retained before status selection, so the incomplete result still carries observed failures.

Alternatives considered:

- Return early on partial rejection: rejected because it would discard the table evidence required by the frozen CLI contract.
- Let the report layer choose priority: rejected because every consumer must agree on the same domain outcome and eventual exit reason.

## Risks / Trade-offs

- [Risk] `internal/budget` now depends on the analysis reliability type. → Mitigation: the dependency is one-way (`analysis` does not import `budget`) and avoids duplicating a frozen domain taxonomy.
- [Risk] Keeping the legacy unvalidated `Check` primitive allows direct callers to compare invalid constructed inputs. → Mitigation: document it as the comparison primitive; application flows use `Evaluate`, whose validation is exhaustive.
- [Risk] Empty limits yield `PASSED` if `Evaluate` is called directly. → Mitigation: #17 already rejects empty effective check policies; retaining vacuous primitive semantics keeps responsibility in one layer.
- [Risk] Uppercase status values mix display-oriented tokens into domain data. → Mitigation: the tokens are frozen product states, and retaining them exactly prevents downstream mapping drift.

## Migration Plan

1. Add validation, status/override models, and evaluation code without changing current call sites.
2. Add exhaustive focused tests while retaining the existing checker tests unchanged.
3. Run full Go, race, vet, lint, and OpenSpec gates, then archive the new capability.
4. Issue #20 adopts the new API when wiring `check`; rollback before that point is a focused revert of additive budget files and fields.
