## Context

See [proposal.md](proposal.md#why) for motivation and [project-root-discovery/spec.md](specs/project-root-discovery/spec.md) for the behavior contract.

The repository has no `internal/project` package today. CLI orchestration already centralizes fatal errors at exit code `2`, while TSCN parsing accepts in-memory readers and has no filesystem dependency. This change must establish only the root-discovery half of MVP section 13: issue #7 will add canonical scene/resource resolution, containment, display paths, and symlink policy, and issue #20 will wire final Cobra flags and application flows.

Root discovery needs filesystem metadata but never file contents. The relevant environmental values are the raw scene input, current working directory, optional explicit project argument, and `stat` results. Keeping all four explicit makes tests deterministic and lets later application composition own `os.Getwd` and process concerns.

## Goals / Non-Goals

**Goals:**

- Establish a small project package with a request/result contract reusable by the resolver and application layer.
- Make every branch observable through typed errors and controlled filesystem metadata.
- Preserve host-platform path semantics by using the Go filepath model rather than manual separator or root logic.

**Non-Goals:**

- Canonicalize through symlinks or prove that a scene lies inside the returned root.
- Open the scene or `project.godot`; only file metadata is inspected.
- Add a new diagnostic `SB` code: the frozen MVP catalog does not assign one to project-discovery failures.
- Introduce a virtual filesystem abstraction broad enough for later scene reading.

## Decisions

### 1. Model one discovery request and one root result

The project package will accept a request containing `SceneInput`, an explicit absolute `WorkingDirectory`, and optional `ExplicitProject`. It will return a small root value containing the absolute cleaned project directory and its `project.godot` marker path. A `res://` prefix is classified before host-path operations; every other scene input is treated as a host filesystem path.

The application layer will obtain the actual cwd later and pass it into the request. Relative scene and explicit-project values are joined to that cwd; absolute values are cleaned without rebasing. The finder does not call `os.Getwd`, which avoids hidden process state in tests.

Alternative considered: separate public methods for filesystem, resource, and explicit discovery. Rejected because precedence and shared validation could drift across methods, while callers naturally possess one combined request.

### 2. Inject only the metadata operation discovery uses

`Finder` will contain a stat-function seam with the same input/output shape as `os.Stat`. The production constructor supplies `os.Stat`; tests can wrap or replace it to record traversal and inject inaccessible-path errors. Most success cases will still use isolated temporary directories because they exercise real platform parent/root behavior without depending on developer files.

Alternative considered: define a full filesystem interface or adopt `io/fs.FS`. Rejected because discovery needs host absolute paths and exactly one operation; `io/fs.FS` paths are slash-separated relative names and would distort the behavior issue #6 is meant to verify.

Alternative considered: use temporary directories only. Rejected because permission tests are unreliable across operating systems and privileged environments, and an injected stat seam cleanly proves short-circuit and error behavior.

### 3. Validate explicit project first, then filesystem scene preconditions

If `ExplicitProject` is non-empty, its shape and marker are validated first so an invalid explicit value never falls back to auto-discovery. A valid explicit root is retained while filesystem scene input is independently validated as an existing regular `.tscn`; `res://` input needs no scene stat. The retained explicit root is then returned without an ancestor search.

This order preserves both contracts: explicit project controls root selection, while a missing or non-scene filesystem root input is still fatal. Containment between a filesystem scene and explicit project is deliberately deferred to issue #7.

Alternative considered: return immediately after validating the explicit project. Rejected because it would allow an invalid filesystem root scene to cross the discovery boundary, contrary to section 13.1.

### 4. Traverse parents using filepath identity

Automatic discovery starts at the validated filesystem scene directory or at the supplied cwd for `res://`. At each level it stats `Join(current, "project.godot")`; a regular file succeeds, absence or a non-regular entry continues, and any other stat error becomes a typed filesystem failure. Traversal stops when `Dir(current) == current`, which handles Unix roots and Windows volume/UNC roots without string-prefix logic.

The root is lexical absolute/clean at this stage. `EvalSymlinks`, segment-aware containment, case behavior of target paths, and canonical cache keys remain one coherent resolver responsibility in issue #7.

Alternative considered: search for a filename using directory listings. Rejected because one exact stat per ancestor is simpler, deterministic, and does not require reading unrelated directory entries.

### 5. Use project-specific typed error reasons, not diagnostic codes

The package will define a typed error with a stable reason enum, relevant path, optional detail, and wrapped cause. Reasons cover invalid working directory/request state, invalid filesystem scene, invalid explicit project, filesystem inspection failure, and project not found. Callers can use `errors.As`, and filesystem causes remain available through `errors.Is`/`Unwrap`.

Human messages remain actionable. In particular, project-not-found text includes “run from inside a Godot project or pass `--project`”. The existing CLI execution boundary maps any returned command error to exit `2` without stack traces; a focused CLI boundary test will exercise that behavior with a project error without wiring unfinished commands.

Alternative considered: assign a new `SB2xxx` code. Rejected because section 11.1 freezes the MVP diagnostic catalog and contains no project-discovery code; inventing one in an implementation issue would be a specification change.

Alternative considered: sentinel errors only. Rejected because callers and tests also need structured path and wrapped-cause context.

### 6. Preserve package boundaries and defer final CLI wiring

Only `internal/project` will import filesystem/path packages. It will return values and errors, never print or exit. Parser, metrics, diagnostics, presets, and budgets remain unchanged. No `--project` Cobra state will be introduced until issue #20 composes finder and resolver into real commands, avoiding dormant flags whose behavior cannot yet run end to end.

## Risks / Trade-offs

- [Lexical roots may contain symlink components] → Document the result as absolute/clean rather than canonical; issue #7 will evaluate symlinks and enforce containment once for every path operation.
- [A marker symlink is followed by `os.Stat` and can appear regular] → Treat it as a regular marker for discovery, then canonicalize the root in issue #7; do not duplicate partial symlink policy here.
- [Case-sensitive `.tscn` checking rejects `.TSCN` on every platform] → This is intentional, deterministic contract enforcement independent of host filesystem case behavior.
- [Injected stat functions can return inconsistent metadata] → Keep the seam narrow and test production behavior with real temporary files for all ordinary cases.
- [CLI test imports project before commands use it] → Limit that dependency to test code solely to verify the already-shipped fatal boundary; production composition remains owned by issue #20.

## Migration Plan

1. Add table-driven tests for explicit project validation, filesystem and `res://` start selection, nearest-root traversal, root termination, non-regular markers, and typed failures.
2. Add the project request/root/error models and finder with production and injected-stat constructors.
3. Verify project-not-found actionability, wrapped filesystem causes, and the generic CLI exit-2 boundary.
4. Audit package imports and run build, test, race, vet, lint, and strict OpenSpec validation.

Rollback is a normal Git revert of the focused implementation and test commits. No persisted data, configuration, external API, or runtime migration is involved.
