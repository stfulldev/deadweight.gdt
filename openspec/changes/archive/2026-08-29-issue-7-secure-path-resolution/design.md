## Context

See `proposal.md` for motivation and `specs/secure-path-resolution/spec.md` for the behavioral contract. `internal/project` currently discovers an absolute cleaned root with a narrow `StatFunc` seam, but deliberately does not canonicalize symlinks or resolve scene/resource paths. Later graph and completeness work needs a single path contract before it can safely open or classify referenced targets.

The implementation remains inside the standalone Go binary. It cannot depend on Godot's UID database, import pipeline, network services, OpenSpec, Node.js, parser state, Cobra handlers, or process streams. Go's native `filepath` semantics and the Linux/macOS/Windows CI matrix define host-path behavior.

## Goals / Non-Goals

**Goals:**

- Extend the existing `internal/project` boundary rather than introduce a second filesystem-policy package.
- Separate fatal root-scene failures from expected nonfatal unresolved resource evidence while giving both stable machine-readable reasons and wrapped causes.
- Produce one canonical path pipeline for lexical containment, existing-target symlink evaluation, missing-target ancestor evaluation, regular-file validation, and display conversion.
- Keep filesystem metadata and symlink evaluation injectable without abstracting the entire filesystem.

**Non-Goals:**

- Make later file opens race-free against a malicious concurrent filesystem mutation; the resolver validates a snapshot and does not own downstream file descriptors.
- Emulate foreign host path syntax on the current OS or treat backslashes as separators on Unix.
- Assign diagnostic codes or completeness/reliability verdicts; later application and graph layers map stable resolution reasons into those policies.
- Combine root discovery and path resolution into one operation or silently canonicalize the existing `Root` model.

## Decisions

### Keep resolver models beside discovery in `internal/project`

Additive types will live in focused resolver files within the current package:

```go
type ResolvedPath struct {
    Canonical string
    Display   string
    Original  string
}

type ResolutionReason string

type Resolution struct {
    Path      ResolvedPath
    Reason    ResolutionReason
    Candidate string
    Err       error
}

type ResolveError struct {
    Reason    ResolutionReason
    Original  string
    Candidate string
    Err       error
}
```

`Resolution` represents resource success or expected unresolved evidence and exposes `Resolved()` plus `Unwrap()` behavior. `ResolveError` is returned by resolver construction and root-scene resolution, where failure is fatal input rather than analysis partiality. Both share the same stable reason vocabulary so later callers do not translate equivalent path failures twice.

Reasons include resolved, invalid project root, invalid cwd, invalid scene input, empty, UID-only, user-data, unsupported target, missing, outside project, filesystem, and invalid declaring scene. Human-readable text remains actionable, but branching uses the reason value and `errors.Is`/`errors.As`.

Alternative: extend the existing discovery `ErrorReason` and `Error`. Rejected because discovery failures and nonfatal resource resolutions have different control flow, and growing one error type would blur the issue #6 contract.

Alternative: create a general filesystem package. Rejected because project-relative `res://` semantics, raw diagnostic identity, and unresolved classifications are domain policy rather than generic filesystem utilities.

### Construct a resolver from an absolute project directory and canonicalize once

`NewResolver(projectRoot string)` validates an absolute existing directory, evaluates its symlinks once, cleans the result, and stores that canonical root. An injected constructor accepts only the existing `StatFunc` plus a narrow `EvalSymlinksFunc`; production uses `os.Stat` and `filepath.EvalSymlinks`. It does not re-run project marker discovery or inspect `project.godot` contents.

Canonicalizing once makes all cache identities and containment comparisons share one root. It also allows an explicitly selected project directory that itself passes through a symlink while ensuring returned target paths use the real root.

Alternative: accept `Root` and assume its directory is canonical. Rejected because issue #6 guarantees only an absolute cleaned path and explicitly deferred symlink semantics.

Alternative: inject `fs.FS`. Rejected because `fs.FS` uses slash-separated relative names and does not model host absolute paths, volumes, or `EvalSymlinks` directly.

### Resolve candidates in two phases: lexical placement, then canonical placement

Each input is classified and converted into an absolute cleaned candidate:

- `res://` trims only the exact lower-case scheme and joins from the canonical root;
- root-scene relative input joins from the supplied absolute cwd;
- resource relative input joins from the validated canonical declaring-scene directory;
- host-absolute input remains absolute.

The candidate first passes `filepath.Rel` containment against the canonical root. A result equal to `..`, beginning with `..` plus the native separator, or absolute is outside. This is segment-aware and rejects sibling-prefix collisions and lexical `..` escapes.

For an existing candidate, the resolver calls `EvalSymlinks`, cleans the canonical result, repeats the same containment check, and then inspects regular-file status. For a missing candidate, it walks upward with `filepath.Dir` until metadata identifies the nearest existing ancestor, evaluates that ancestor's symlinks, appends the still-missing relative suffix, and checks canonical containment before returning the missing reason. This distinguishes an ordinary in-project miss from a miss below an escaping symlink.

No content is opened in this pipeline. Non-not-found metadata or evaluation failures retain their causes as filesystem reasons. A broken symlink or a race that yields not-found remains a missing result after all resolvable existing ancestors have been checked.

Alternative: use `strings.HasPrefix`. Rejected because `/game` does not contain `/game-old`, and case/volume rules are host-specific.

Alternative: call `EvalSymlinks` only on the final target. Rejected because a missing final target below an existing escaping parent symlink cannot then be distinguished from a safe in-project miss.

Alternative: use `filepath.Abs` as canonicalization. Rejected because it cleans lexical segments but does not resolve symlinks.

### Treat root scenes and declared resources differently after the shared pipeline

`ResolveSceneInput(input, cwd)` validates a nonempty absolute cwd for relative inputs, accepts exact `res://` or host paths, uses the shared containment pipeline, and requires an existing regular file with the exact `.tscn` extension. Any failure becomes `*ResolveError` because there is no analysis without a valid root scene.

`ResolveResource(fromScene, raw)` classifies empty, exact lower-case `uid://`, and exact lower-case `user://` before filesystem calls. Other strings containing `://` are unsupported. Relative values require `fromScene` to be absolute, cleaned, already canonical, regular, and inside the project. A successful target may use any extension because later layers distinguish nested `.tscn`, imported/binary scenes, and ordinary resources. Missing, directory/non-regular, outside, and inaccessible candidates return `Resolution` values rather than Go errors.

Alternative: reject non-`.tscn` resource targets in the resolver. Rejected because ordinary textures, materials, scripts, and audio are valid external resources, while imported-scene partiality belongs to the later target classifier.

Alternative: return `error` for every unresolved resource. Rejected because expected unsupported and missing references contribute typed partial-analysis evidence and should not be confused with fatal root input.

### Derive display paths only from validated canonical in-root paths

Display conversion uses `filepath.Rel(canonicalRoot, canonicalTarget)`, `filepath.ToSlash`, and the exact `res://` prefix. The root maps to `res://`; descendants map to `res://path/to/file`. `DisplayPath(abs string) string` returns an empty string for non-absolute, non-clean, or outside-root input, while successful resolution constructs `ResolvedPath` only after containment has passed.

This keeps reports deterministic and avoids producing a plausible project path for an unsafe target. It does not scan directories or repair case; native metadata lookup determines whether case is valid.

Alternative: store only a `res://` value and reconstruct host paths later. Rejected because canonical absolute identity is required for I/O, graph keys, and caches.

Alternative: preserve native separators in display strings. Rejected because output and golden files must be stable across hosts.

### Test security behavior with real temporary trees plus narrow injected failures

Table-driven tests use temporary directories for normal relative, parent-relative, absolute, `res://`, nested, missing, and real symlink cases. Injected `StatFunc` and `EvalSymlinksFunc` record paths and force permission or race-like errors that are unreliable to reproduce through host permissions. Pure reason/display contract tests cover stable values; native Windows CI covers Windows `filepath` volume and separator behavior.

The package remains free of parser, diagnostic, CLI, report, Godot, and network imports. A focused CLI boundary test is unnecessary here because issue #7 does not alter orchestration; issue #6 already proves typed project failures retain exit code 2.

Alternative: mock every filesystem entry. Rejected because real symlink and path behavior is the security boundary, while fully synthetic metadata can hide platform mistakes.

## Risks / Trade-offs

- [Filesystem mutation after resolution can create a time-of-check/time-of-use race] → Resolver output is a validated snapshot; later I/O code must not bypass resolver policy and can re-resolve immediately before opening. Descriptor-relative secure opens are outside the MVP's read-only local-project threat model.
- [Symlink creation may be unavailable or require privileges on some Windows environments] → Keep real symlink tests conditional only when the OS reports an unsupported/privilege error, and retain injected canonicalization tests so the behavior remains covered; CI still exercises native separators and volumes.
- [Case behavior differs across filesystems rather than only operating systems] → Assert no recovery logic everywhere and run the wrong-case existence test only when the fixture proves the host filesystem is case-sensitive.
- [The shared reason vocabulary is larger than the immediate graph consumer needs] → Keep reasons finite, validated by table tests, and avoid prematurely assigning diagnostic codes or analysis verdicts.
- [Canonical project paths can differ from user-entered paths] → Preserve `Original` on resolved values and errors while using only canonical values for containment, I/O, and identity.
- [Validating the nearest existing ancestor adds metadata calls for missing paths] → Paths are short local references and correctness dominates; stop at the canonical root/filesystem root and do not scan directory contents.

## Migration Plan

1. Add resolver/result/error contracts and their table-driven tests without changing existing finder APIs.
2. Add canonical root construction, containment/display helpers, and root-scene resolution tests and implementation.
3. Add resource classification, declaring-scene bases, missing-ancestor canonicalization, and real/injected symlink tests.
4. Run all repository quality gates and cross-platform CI while PR #32 remains draft.
5. Archive and sync the OpenSpec change before marking the PR ready and merging it.

Rollback is additive: revert the resolver production and test commits plus the synced/archive artifacts. Existing discovery and CLI behavior require no data or configuration migration.
