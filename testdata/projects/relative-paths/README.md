# Relative-path fixture

`levels/city.tscn` declares `../props/lamp.tscn`. Resolution must start from the declaring scene directory, stay within the project, and display the canonical dependency as `res://props/lamp.tscn`.

Expected result: `COMPLETE exact`, two nodes, depth two, one scene instance, one light, one external resource, one scene dependency, and two parsed scene files.
