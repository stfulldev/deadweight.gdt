# Malformed fixture

- `format2.tscn` is rejected because MVP 0.1 accepts only `format=3` roots.
- `unclosed-string.tscn` is rejected by the streaming lexer.
- `bad-ext-id.tscn` is rejected because an external-resource ID must be a non-empty scalar.
- `invalid-config.json` is valid JSON with an unknown root key and is rejected as `SB2003` before scene analysis.

Every failure is fatal and maps to CLI exit code 2.
