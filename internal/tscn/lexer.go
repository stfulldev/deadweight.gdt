package tscn

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type runeLookahead struct {
	value rune
	err   error
}

type lexer struct {
	reader     *bufio.Reader
	source     string
	line       int
	column     int
	previousCR bool
	lookahead  *runeLookahead
}

func newLexer(reader io.Reader, source string) *lexer {
	return &lexer{
		reader: bufio.NewReader(reader),
		source: source,
		line:   1,
		column: 1,
	}
}

func (scanner *lexer) nextToken() (token, error) {
	for {
		current, err := scanner.peekRune()
		if err == io.EOF {
			return token{typeID: tokenEOF, position: scanner.position()}, nil
		}
		if err != nil {
			return token{}, newParseError(scanner.source, scanner.position(), "read input: %v", err)
		}

		if current == ' ' || current == '\t' || current == '\v' || current == '\f' || (unicode.IsSpace(current) && current != '\n' && current != '\r') {
			_, _, _ = scanner.consumeRune()
			continue
		}

		if current == ';' {
			scanner.skipComment()
			continue
		}

		break
	}

	current, position, err := scanner.consumeRune()
	if err != nil {
		return token{}, newParseError(scanner.source, scanner.position(), "read input: %v", err)
	}

	switch current {
	case '\n':
		return token{typeID: tokenNewline, position: position}, nil
	case '\r':
		if next, peekErr := scanner.peekRune(); peekErr == nil && next == '\n' {
			_, _, _ = scanner.consumeRune()
		}
		return token{typeID: tokenNewline, position: position}, nil
	case '[':
		return token{typeID: tokenLeftBracket, literal: "[", position: position}, nil
	case ']':
		return token{typeID: tokenRightBracket, literal: "]", position: position}, nil
	case '(':
		return token{typeID: tokenLeftParen, literal: "(", position: position}, nil
	case ')':
		return token{typeID: tokenRightParen, literal: ")", position: position}, nil
	case '{':
		return token{typeID: tokenLeftBrace, literal: "{", position: position}, nil
	case '}':
		return token{typeID: tokenRightBrace, literal: "}", position: position}, nil
	case '=':
		return token{typeID: tokenEquals, literal: "=", position: position}, nil
	case ',':
		return token{typeID: tokenComma, literal: ",", position: position}, nil
	case '"':
		return scanner.readString(position)
	default:
		return scanner.readBare(current, position)
	}
}

func (scanner *lexer) readString(position Position) (token, error) {
	var value strings.Builder

	for {
		current, currentPosition, err := scanner.consumeRune()
		if err == io.EOF {
			return token{}, newParseError(scanner.source, position, "unterminated string")
		}
		if err != nil {
			return token{}, newParseError(scanner.source, scanner.position(), "read string: %v", err)
		}

		switch current {
		case '"':
			return token{typeID: tokenString, literal: value.String(), position: position}, nil
		case '\n', '\r':
			value.WriteRune(current)
		case '\\':
			escaped, escapePosition, escapeErr := scanner.consumeRune()
			if escapeErr == io.EOF {
				return token{}, newParseError(scanner.source, currentPosition, "unterminated string escape")
			}
			if escapeErr != nil {
				return token{}, newParseError(scanner.source, currentPosition, "read string escape: %v", escapeErr)
			}

			switch escaped {
			case '"', '\\', '/':
				value.WriteRune(escaped)
			case 'n':
				value.WriteByte('\n')
			case 'r':
				value.WriteByte('\r')
			case 't':
				value.WriteByte('\t')
			case 'b':
				value.WriteByte('\b')
			case 'f':
				value.WriteByte('\f')
			case 'u':
				decoded, decodeErr := scanner.readUnicodeEscape(4, escapePosition)
				if decodeErr != nil {
					return token{}, decodeErr
				}
				value.WriteRune(decoded)
			case 'U':
				decoded, decodeErr := scanner.readUnicodeEscape(8, escapePosition)
				if decodeErr != nil {
					return token{}, decodeErr
				}
				value.WriteRune(decoded)
			default:
				return token{}, newParseError(scanner.source, escapePosition, "unsupported string escape \\%c", escaped)
			}
		default:
			value.WriteRune(current)
		}
	}
}

func (scanner *lexer) readUnicodeEscape(length int, position Position) (rune, error) {
	digits := make([]rune, 0, length)
	for range length {
		current, _, err := scanner.consumeRune()
		if err != nil {
			return 0, newParseError(scanner.source, position, "incomplete Unicode escape")
		}
		digits = append(digits, current)
	}

	value, err := strconv.ParseUint(string(digits), 16, 32)
	if err != nil {
		return 0, newParseError(scanner.source, position, "invalid Unicode escape %q", string(digits))
	}
	if value > unicode.MaxRune {
		return 0, newParseError(scanner.source, position, "Unicode escape is outside the valid range")
	}

	return rune(value), nil
}

func (scanner *lexer) readBare(first rune, position Position) (token, error) {
	var value strings.Builder
	value.WriteRune(first)

	for {
		current, err := scanner.peekRune()
		if err == io.EOF || isTokenDelimiter(current) {
			break
		}
		if err != nil {
			return token{}, newParseError(scanner.source, scanner.position(), "read token: %v", err)
		}

		consumed, _, consumeErr := scanner.consumeRune()
		if consumeErr != nil {
			return token{}, newParseError(scanner.source, scanner.position(), "read token: %v", consumeErr)
		}
		value.WriteRune(consumed)
	}

	literal := value.String()
	if literal == "true" || literal == "false" {
		return token{typeID: tokenBool, literal: literal, position: position}, nil
	}
	if _, err := strconv.ParseInt(literal, 10, 64); err == nil {
		return token{typeID: tokenInteger, literal: literal, position: position}, nil
	}

	return token{typeID: tokenIdentifier, literal: literal, position: position}, nil
}

func (scanner *lexer) skipComment() {
	for {
		current, err := scanner.peekRune()
		if err != nil || current == '\n' || current == '\r' {
			return
		}
		_, _, _ = scanner.consumeRune()
	}
}

func (scanner *lexer) peekRune() (rune, error) {
	if scanner.lookahead == nil {
		value, _, err := scanner.reader.ReadRune()
		scanner.lookahead = &runeLookahead{value: value, err: err}
	}

	return scanner.lookahead.value, scanner.lookahead.err
}

func (scanner *lexer) consumeRune() (rune, Position, error) {
	position := scanner.position()

	var value rune
	var err error
	if scanner.lookahead != nil {
		value = scanner.lookahead.value
		err = scanner.lookahead.err
		scanner.lookahead = nil
	} else {
		value, _, err = scanner.reader.ReadRune()
	}
	if err != nil {
		return 0, position, err
	}

	switch value {
	case '\r':
		scanner.line++
		scanner.column = 1
		scanner.previousCR = true
	case '\n':
		if !scanner.previousCR {
			scanner.line++
			scanner.column = 1
		}
		scanner.previousCR = false
	default:
		scanner.column++
		scanner.previousCR = false
	}

	return value, position, nil
}

func (scanner *lexer) position() Position {
	return Position{Line: scanner.line, Column: scanner.column}
}

func isTokenDelimiter(value rune) bool {
	if unicode.IsSpace(value) {
		return true
	}

	switch value {
	case ';', '"', '[', ']', '(', ')', '{', '}', '=', ',':
		return true
	default:
		return false
	}
}
