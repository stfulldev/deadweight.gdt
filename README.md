# deadweight.gdt

`deadweight.gdt` is a standalone Go CLI for static complexity analysis of Godot 4 `.tscn` scenes.

The finished MVP will expand nested text scenes, calculate effective scene metrics, and check them against project-specific or built-in heuristic budgets. Godot will not be required to run the analyzer.

> [!IMPORTANT]
> Built-in presets are heuristic guardrails, not performance guarantees. Always profile your game on target hardware.

## Project status

MVP `0.1.0` is under active development.

The initial repository slice currently provides:

- the Go module and CLI entry point;
- stable metric and budget domain types;
- a streaming Godot 4 TSCN subset lexer and parser with source-aware errors;
- the experimental built-in presets `mobile`, `steam-deck`, and `desktop`;
- working `presets` and `presets show` commands;
- unit tests and a cross-platform CI skeleton.

Project-root/path resolution and the recursive scene analyzer are the next implementation milestones. `inspect` and `check` are present in CLI help but intentionally return a clear not-implemented error until their vertical slices land.

The implementation contract is documented in [docs/MVP_0.1_SPEC.md](docs/MVP_0.1_SPEC.md).

Contributor changes use a lightweight
[OpenSpec + Codex workflow](docs/SPEC_DRIVEN_WORKFLOW.md) alongside GitHub Issues
and Draft PRs. OpenSpec is development-only and is not required to build or run
the Go CLI.

## Try the current slice

Requirements: Go 1.24 or newer.

```bash
go run ./cmd/deadweight.gdt presets
go run ./cmd/deadweight.gdt presets show steam-deck
```

Run quality checks:

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Planned MVP usage

```bash
deadweight.gdt inspect res://levels/city.tscn
deadweight.gdt check res://levels/city.tscn --preset steam-deck
deadweight.gdt check res://levels/city.tscn --profile shipping
```

The analyzer will report `PARTIAL` whenever imported, binary, missing, or inherited scene data cannot be represented accurately by the standalone static analyzer.

## What this tool is not

`deadweight.gdt` is not an FPS predictor, runtime profiler, or official hardware certification tool. Static node counts cannot account for scripts, physics behavior, shaders, visibility, imported scene internals, or nodes created at runtime.

## License

[MIT](LICENSE)
