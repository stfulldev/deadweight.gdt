# Format-4 compatibility fixtures

- `root.tscn` expands a format-3 child that expands a format-4 leaf with opaque packed values.
- `derived.tscn` inherits a supported format-4 base and retains the existing approximate inheritance contract.
- `equivalent3.tscn` and `equivalent4.tscn` have identical metric-relevant content.
- `future-root.tscn` reaches a format-5 dependency and must fail rather than downgrade it to partial.
