# Complete fixture

- `simple.tscn` is a single-file exact scene with three nodes, depth two, one mesh, one light, and one shadow-casting light.
- `nested.tscn` mounts `deps/child.tscn`; both scenes reference the same `assets/shared.png` so unique resource semantics can be asserted.
- Expected nested result: `COMPLETE exact`, four nodes, depth three, one scene instance, one mesh, one light, one shadow light, two external resources, one scene dependency, and two parsed scene files.

All assets are inert test files. Godot is not required to read or analyze this project.
