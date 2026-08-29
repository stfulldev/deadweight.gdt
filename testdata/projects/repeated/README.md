# Repeated fixture

`city.tscn` mounts the same `lamp.tscn` three times. The lamp contribution is multiplied per occurrence while its canonical file is parsed once.

Expected result: `COMPLETE exact`, seven nodes, depth three, three scene instances, three meshes, three lights, three shadow lights, one external resource, one scene dependency, and two parsed scene files.
