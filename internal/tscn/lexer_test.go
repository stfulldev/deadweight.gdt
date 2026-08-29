package tscn

import (
	"errors"
	"strings"
	"testing"
)

func TestLexerRecognizesStringsCommentsCRLFAndScalars(t *testing.T) {
	t.Parallel()

	scanner := newLexer(strings.NewReader("\"alpha;beta\\n\\u263a\" ; ignored\r\ntrue -12 [ ]"), "lexer.tscn")
	want := []token{
		{typeID: tokenString, literal: "alpha;beta\n☺", position: Position{Line: 1, Column: 1}},
		{typeID: tokenNewline, position: Position{Line: 1, Column: 31}},
		{typeID: tokenBool, literal: "true", position: Position{Line: 2, Column: 1}},
		{typeID: tokenInteger, literal: "-12", position: Position{Line: 2, Column: 6}},
		{typeID: tokenLeftBracket, literal: "[", position: Position{Line: 2, Column: 10}},
		{typeID: tokenRightBracket, literal: "]", position: Position{Line: 2, Column: 12}},
		{typeID: tokenEOF, position: Position{Line: 2, Column: 13}},
	}

	for index, expected := range want {
		actual, err := scanner.nextToken()
		if err != nil {
			t.Fatalf("token %d: nextToken() error = %v", index, err)
		}
		if actual != expected {
			t.Errorf("token %d = %#v, want %#v", index, actual, expected)
		}
	}
}

func TestLexerRejectsUnterminatedAndUnsupportedStringEscapes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "unterminated", input: `"value`, want: "unterminated string"},
		{name: "newline", input: "\"value\nnext\"", want: "unescaped newline"},
		{name: "escape", input: `"value\q"`, want: "unsupported string escape"},
		{name: "unicode", input: `"value\u12xz"`, want: "invalid Unicode escape"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			scanner := newLexer(strings.NewReader(test.input), "bad.tscn")
			_, err := scanner.nextToken()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("nextToken() error = %v, want containing %q", err, test.want)
			}

			var parseError *ParseError
			if !errors.As(err, &parseError) || parseError.Code != invalidSceneCode {
				t.Fatalf("error = %#v, want *ParseError with code %s", err, invalidSceneCode)
			}
		})
	}
}
