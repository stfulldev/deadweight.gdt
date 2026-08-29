package tscn

type tokenType uint8

const (
	tokenInvalid tokenType = iota
	tokenIdentifier
	tokenString
	tokenInteger
	tokenBool
	tokenLeftBracket
	tokenRightBracket
	tokenLeftParen
	tokenRightParen
	tokenLeftBrace
	tokenRightBrace
	tokenEquals
	tokenComma
	tokenNewline
	tokenEOF
)

type token struct {
	typeID   tokenType
	literal  string
	position Position
}

func (value token) describe() string {
	switch value.typeID {
	case tokenIdentifier:
		return value.literal
	case tokenString:
		return "string"
	case tokenInteger:
		return "integer"
	case tokenBool:
		return "boolean"
	case tokenLeftBracket:
		return "["
	case tokenRightBracket:
		return "]"
	case tokenLeftParen:
		return "("
	case tokenRightParen:
		return ")"
	case tokenLeftBrace:
		return "{"
	case tokenRightBrace:
		return "}"
	case tokenEquals:
		return "="
	case tokenComma:
		return ","
	case tokenNewline:
		return "newline"
	case tokenEOF:
		return "end of file"
	default:
		return "invalid token"
	}
}
