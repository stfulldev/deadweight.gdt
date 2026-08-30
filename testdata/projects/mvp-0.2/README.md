# MVP 0.2 cross-feature fixture

`root.tscn` is the release-acceptance baseline. It instantiates `child.tscn`
twice, so additive contributions and dependency occurrences can be verified
while both scenes refer to the same portable resource identity.

`root.candidate.tscn` is copied over `root.tscn` by the CLI integration test to
produce a same-root semantic diff. The project-local strict configuration adds
the `shipping` profile on top of the built-in `steam-deck` heuristic and keeps
the fixture below its effective budgets.

The `.png` file is deliberately a text placeholder: deadweight.gdt resolves
static resource identity and never decodes or imports the external asset.
