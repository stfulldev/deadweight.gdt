## Purpose

Defines the streaming lexical and section-body behavior required to parse valid real-world Godot format-3 text scenes while retaining deterministic source-aware failures.

## ADDED Requirements

### Requirement: Quoted strings may span physical lines
The format-3 lexer SHALL accept literal LF, CRLF, and CR line endings between an opening and closing quote. The token value MUST retain the literal line-ending runes, and source positions for tokens and failures after the string MUST count each physical line ending once. A physical line ending inside a quoted string MUST NOT terminate the containing property value or produce an unescaped-newline failure.

#### Scenario: Multiline LF string
- **WHEN** a format-3 property contains a quoted string with one or more literal LF line endings before its closing quote
- **THEN** parsing succeeds and the string is consumed as one value rather than multiple properties
- **AND** the next token begins at the correct physical line and column

#### Scenario: Multiline CRLF string
- **WHEN** a format-3 property contains a quoted string with literal CRLF line endings before its closing quote
- **THEN** parsing succeeds, the token retains each CRLF pair, and the next token position advances by one line per pair

#### Scenario: Multiline CR string
- **WHEN** a format-3 property contains a quoted string with a literal CR line ending before its closing quote
- **THEN** parsing succeeds and the next token position advances by one physical line

### Requirement: Malformed quoted strings remain typed failures
The lexer MUST return a source-positioned `SB2001` parse error when EOF occurs before a quoted string closes, when a string escape is incomplete, or when an escape is outside the supported format-3 subset. Accepting physical line endings MUST NOT turn malformed strings into successful documents or remove their actionable source context.

#### Scenario: Unterminated multiline string
- **WHEN** a quoted string crosses a physical line and reaches EOF without a closing quote
- **THEN** parsing fails with `SB2001` at the opening string position and identifies the unterminated string

#### Scenario: Unsupported escape after a physical line
- **WHEN** a multiline quoted string later contains an unsupported escape sequence
- **THEN** parsing fails with `SB2001` at the escape position and identifies the unsupported escape

### Requirement: Section bodies accept quoted property names
A format-3 section body SHALL accept either an identifier token or a quoted-string token as a property name before `=`. Recognized properties MUST retain their existing typed semantics when expressed by their canonical name, while unknown quoted or identifier properties MUST be skipped with the same balanced streaming rules. A string token in any other top-level section-body position MUST remain a deterministic parse failure.

#### Scenario: Unknown quoted property name
- **WHEN** a section contains a property such as `"nodes/Animation 2/node" = SubResource("13")`
- **THEN** the parser accepts the quoted name, skips the balanced value, and continues with following properties and sections

#### Scenario: Canonical property behavior is unchanged
- **WHEN** a node contains ordinary identifier properties including `shadow_enabled = true`
- **THEN** recognized-property extraction and duplicate/value validation remain unchanged

#### Scenario: Quoted name without assignment
- **WHEN** a section body begins with a quoted string that is not followed by `=` and a property value
- **THEN** parsing fails deterministically with a source-positioned `SB2001` error
