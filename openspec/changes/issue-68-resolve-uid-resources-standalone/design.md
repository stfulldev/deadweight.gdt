## Context

See `proposal.md` for motivation and the delta specs for observable behavior. Today `internal/project.Resolver` rejects every `uid://` root or resource before filesystem access, while parsed external-resource records already retain both UID and path. Project discovery treats only `res://` as a context-dependent input, and the official-demo runner preclassifies eleven UID main scenes without invoking the binary.

Godot's own contracts provide four local evidence forms relevant to standalone analysis:

- `ResourceUID` stores the project cache at the configured project-data path and uses a count plus little-endian `(uint64 id, uint32 UTF-8 length, bytes)` entry layout: https://github.com/godotengine/godot/blob/4.5/core/io/resource_uid.cpp
- text scenes/resources carry ownership UIDs in their first resource header, and text external resources prefer a known UID mapping while falling back to their path when the UID is unknown: https://github.com/godotengine/godot/blob/master/scene/resources/resource_format_text.cpp
- loaders without custom UID support read an adjacent `.uid` file: https://github.com/godotengine/godot/blob/master/core/io/resource_loader.cpp
- imported resources read `uid` from the `[remap]` section of the adjacent `.import` file: https://github.com/godotengine/godot/blob/master/core/io/resource_importer.cpp

The selected project remains the trust boundary. Generated `.godot` state is commonly absent from source checkouts, so the design cannot depend on `uid_cache.bin`; conversely, a present cache may be stale and cannot override current resource-owned evidence.

## Goals / Non-Goals

**Goals:**

- resolve supported UID roots, scene mounts, inherited bases, and ordinary resources to the same canonical identities used by path-based analysis;
- build a deterministic cold-checkout index from bounded project-local metadata, with the generated cache as an optional fallback;
- make every ambiguity, malformed source, unsafe path, and fallback decision typed and testable;
- keep the UID index lazy, immutable after construction, and shared for one application invocation;
- preserve all frozen metrics, report schema-v1 shapes, exit meanings, path behavior, and the no-Godot default runtime.

**Non-Goals:**

- reproducing the editor's import pipeline, filesystem mutation, UID generation, resource moves, or cache repair;
- loading imported/binary scene contents, executing scripts, scanning remote registries, or invoking a Godot bridge;
- resolving UIDs embedded in arbitrary GDScript expressions or unparsed Variant payloads;
- changing contribution identities, metric definitions, presets, budgets, or report schema versions.

## Decisions

### 1. Separate UID syntax/metadata parsing from project path policy

Add a small pure Go UID codec and bounded metadata readers with `io.Reader` inputs. They decode canonical Godot UID text, the first supported `.tscn`/`.tres` header, one-line `.uid` files, `[remap] uid` from `.import`, and the supported binary cache layout. Readers return typed records and never perform path resolution, traversal, printing, or process actions.

The project layer owns directory traversal, source precedence, symlink containment, regular-file validation, display paths, and lazy lifecycle. This keeps Godot serialization details out of the secure path resolver while keeping trust policy in one package.

Alternatives rejected:

- Fully parse every `.tscn` with the analysis parser: this would make an unrelated malformed scene fatal to any UID lookup and would perform unnecessary whole-document work.
- Use regular expressions for resource headers or `.import`: this would create a second fragile parser path and mishandle quoting, comments, and malformed delimiters.
- Put filesystem scanning in the TSCN package: that would violate the parser/filesystem dependency boundary.

The header and metadata readers will use small streaming state machines and strict bounded reads. No Variant AST or decoded resource payload is retained.

### 2. Build a lazy immutable index through injected effects

The project resolver receives a lazy UID lookup dependency backed by an index builder. The builder exposes injected walk, open, stat, and symlink/canonicalization effects alongside the existing resolver effects. Its first query performs one deterministic project walk; subsequent queries reuse immutable owned maps.

The walk:

1. reads `project.godot` only for the declared Godot feature version and `application/config/use_hidden_project_data_directory`;
2. walks canonical project entries in normalized lexical order without following directory symlinks;
3. skips VCS internals and the selected project-data directory during the source scan;
4. reads only supported resource headers, applicable sidecars, and `.import` metadata;
5. reads the selected project-data `uid_cache.bin` separately when it exists;
6. validates every candidate with the existing segment-aware and symlink-aware containment boundary before making it usable.

Path-only operations never force the lazy value. Index state is scoped to the resolver/application instance and is neither global nor persisted.

Alternative rejected: prebuild the index during project discovery. That would penalize every command and mix content parsing into a finder whose current contract intentionally uses metadata only.

### 3. Version-gate evidence formats instead of guessing

Parse the highest declared Godot 4 feature version from `project.godot`. Text resource ownership headers and `.import` remap UIDs use their verified Godot 4 readers. General `.uid` sidecars are accepted only for project versions where that source form is supported. The binary cache reader accepts only the verified Godot 4 layout and rejects truncation, impossible counts/lengths, trailing structural corruption, or arithmetic overflow.

An absent version does not invent future behavior: sources whose format can be verified independently remain usable, while version-dependent or unknown future forms produce unsupported evidence only when relevant to a requested UID.

Alternative rejected: accept every file ending in `.uid` in every project version. That could turn unrelated user files or future incompatible metadata into authoritative mappings.

### 4. Use resource ownership before generated cache

Index entries retain UID, candidate canonical/display path, source kind, and source location. Resolution groups evidence by decoded UID and applies this order:

1. unique direct ownership from a supported text header, applicable `.uid` sidecar, or `.import` remap;
2. a unique cache-only entry whose target remains an existing contained regular file.

Direct claims from multiple canonical resources are ambiguous. A conflicting cache record is stale and cannot override a unique direct owner. Equivalent claims for the same canonical resource collapse without losing source evidence. All evidence is sorted by normalized source kind and canonical path before results are published.

Alternative rejected: last-writer-wins to mirror a hash map. Directory or map order would make duplicate UID behavior host-dependent and could redirect analysis silently.

### 5. Extend resolution with combined UID/path references

Keep the existing path-only resolver entry points as compatibility wrappers. Add a combined resource-reference operation that receives declaring scene, UID, and path from the already retained local external-resource record.

The decision table is:

| UID evidence | Path evidence | Result |
|---|---|---|
| unique secure mapping | any | use the UID-mapped canonical resource |
| absent/malformed/unknown | usable path | use the ordinary path and retain UID fallback evidence |
| ambiguous/stale/unsafe/unsupported | usable path | use the path conservatively and retain uncertainty evidence |
| unresolved | absent/unusable | return typed UID/path unresolved evidence |

Root UID input has no fallback and therefore converts every non-unique outcome into the existing fatal root-resolution boundary. A successful mapped root must still be a contained regular lowercase `.tscn`.

`ResolvedPath.Original` remains the user's or document's raw value. Secondary UID source/fallback evidence is carried separately so canonical graph and cache identities remain path-based and portable.

Alternative rejected: replace canonical graph identities with UID strings. Existing path identities already provide secure I/O, portable `res://` reporting, cycle detection, and cache reuse; UID aliases must converge on them, not form a parallel graph.

### 6. Reuse the existing analysis pipeline after resolution

Application composition recognizes `uid://` like `res://` for cwd-based project discovery, constructs one resolver/index, and resolves the root before analysis. Recursive graph discovery and occurrence expansion call the combined reference operation for each external-resource record. UID-resolved `.tscn` files enter the existing parser, invocation cache, cycle detector, inherited-scene handling, contributions, coverage, tree, check, and report paths unchanged.

Resolution reasons expand from the current single `uid_only` bucket to stable UID missing, malformed, ambiguous, stale, unsupported-version, and unsafe variants. Unresolved variants map to `SB1006` and existing confidence evidence; successful UID resolution alone creates no diagnostic or confidence downgrade. A path fallback retains the UID evidence required by the specs without changing the chosen canonical identity.

No JSON-v1 structural field is added. Existing diagnostic and unresolved-resource records carry the stable reason and original UID where uncertainty affects a published result.

### 7. Make corpus movement an executable acceptance gate

Remove `UNSUPPORTED_UID_ROOT` preclassification from the PowerShell runner. Every configured main scene, including all eleven `uid://` values, is passed to the standalone binary with its project root. The current path-based characterization proves ten of those roots are otherwise `COMPLETE` and one is `PARTIAL`; after implementation the full 139-scene run determines the committed exact counts. Expected zero unexpected fatal outcomes remains mandatory.

Focused fixtures cover cold checkouts, every source kind and precedence edge, UID/path aliases, duplicate direct claims, corrupt cache lengths, hidden/non-hidden project-data directories, symlink escapes, root/nested/inherited behavior, invocation reuse, and unchanged path-only effects. Existing frozen text and JSON goldens remain unchanged unless a fixture intentionally exercises new UID evidence.

## Risks / Trade-offs

- **[Large projects make the first UID lookup expensive]** → Keep indexing lazy and single-use per invocation, read only bounded headers/metadata, skip project-data/VCS trees, and add deterministic scan benchmarks before considering further optimization.
- **[Generated cache is stale or absent]** → Treat cache as lower-priority optional evidence and prove cold-checkout resolution from source-owned metadata.
- **[Duplicate UIDs make Godot editor behavior dependent on cache history]** → Return explicit ambiguity instead of selecting by traversal order; use a path fallback only with retained uncertainty.
- **[Future Godot changes a metadata format]** → Gate version-dependent sources, reject unsupported layouts, and keep readers isolated so a later compatibility slice can add a decoder without changing graph semantics.
- **[Corrupt cache lengths cause memory or arithmetic abuse]** → Bound file size, entry count, and individual lengths; use checked arithmetic and stream reads; never allocate directly from an unchecked field.
- **[Symlinked or cached paths escape the project]** → Reuse the canonical containment and nearest-existing-ancestor policy before opening target contents or returning an identity.
- **[Path fallback evidence can turn an otherwise usable report partial]** → Restrict confidence downgrade to ambiguity that can affect which resource Godot would load; an unknown UID with an independently verified ordinary path remains analyzable and its advisory evidence must not invent metric uncertainty.
- **[Windows path/case order differs from Unix]** → Normalize display paths, use canonical host semantics for existence, and sort published evidence by stable source kind plus normalized project-relative identity.

## Migration Plan

1. Add pure UID codec/metadata-reader tests and the project index model with injected effects; no existing resolver call site changes yet.
2. Add combined UID/path resolution plus finder and application root wiring while preserving path-only wrappers.
3. Route recursive and inherited external-resource records through combined resolution and map new unresolved reasons to existing diagnostics/confidence.
4. Add CLI/application fixtures, characterize all official UID roots, remove runner preclassification, and commit measured workflow expectations and compatibility docs.
5. Run targeted security/corruption tests, frozen text/JSON acceptance, build/test/race/vet/lint, strict OpenSpec validation, and hosted Linux/macOS/Windows/PowerShell gates.

Rollback is additive: remove UID index injection and combined-reference routing to restore the previous typed `uid_only` boundary. No persisted state, report migration, preset change, or cache cleanup is required.
