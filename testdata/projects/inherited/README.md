# Inherited fixture

`zombie.tscn` inherits `enemy.tscn`, retains an override stub for `Body`, and adds one mesh. The supported base contribution is expanded, but inherited overrides make the result `PARTIAL approximate` with `SB1003`.

Expected metrics: three nodes, depth three, one mesh, one external resource, one scene dependency, and two parsed scene files.
