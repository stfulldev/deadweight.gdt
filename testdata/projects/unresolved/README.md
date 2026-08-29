# Unresolved fixture

- `missing-tscn.tscn` mounts a missing text scene and must report `PARTIAL lower-bound` with `SB1004`.
- `imported-glb.tscn` mounts an imported scene format and must report `PARTIAL lower-bound` with `SB1002`.
- `placeholder.tscn` uses `instance_placeholder` and must report `PARTIAL lower-bound` with `SB1005`.

Each entry keeps the known root/mount contribution and never claims false `COMPLETE` analysis.
