# Cyclic fixture

`A.tscn → B.tscn → C.tscn → A.tscn` is a deterministic three-scene dependency cycle. Inspecting `A.tscn` must fail with exit code 2, diagnostic `SB2002`, and the complete closed chain in that order.
