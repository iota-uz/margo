package lexer

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "basic component",
			input: `\HeroV2`,
			expected: []Token{
				{Type: Component, Value: "HeroV2"},
				{Type: EOF},
			},
		},
		{
			name: "component with property",
			input: `
\Link
	Title: "Hello, Marshal"
	Href: "https://example.com"
`,
			expected: []Token{
				{Type: Component, Value: "Link"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Title"},
				{Type: Colon, Value: ":"},
				{Type: Quote, Value: "\""},
				{Type: Text, Value: "Hello, Marshal"},
				{Type: Quote, Value: "\""},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Href"},
				{Type: Colon, Value: ":"},
				{Type: Quote, Value: "\""},
				{Type: Text, Value: "https://example.com"},
				{Type: Quote, Value: "\""},
				//{Type: LineBreak, Value: "\n"},
				//{Type: Indent, Value: "\t"},
				//{Type: Property, Value: "Size"},
				//{Type: Colon, Value: ":"},
				//{Type: Number, Value: "12"},
				{Type: EOF},
			},
		},
		{
			name: "component as property",
			input: `
\HeroV2
	Title: 
		\Button
`,
			expected: []Token{
				{Type: Component, Value: "HeroV2"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Title"},
				{Type: Colon, Value: ":"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Indent, Value: "\t"},
				{Type: Component, Value: "Button"},
				{Type: EOF},
			},
		},
		{
			name: "nested components",
			input: `
\HeroV2
	Title:  "Congratulations!"
	!Visible
    \ButtonPrimary
        Href: "https://example.com"
        See [demo](https://example.com)
`,
			expected: []Token{
				{Type: Component, Value: "HeroV2"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Title"},
				{Type: Colon, Value: ":"},
				{Type: Quote, Value: "\""},
				{Type: Text, Value: "Congratulations!"},
				{Type: Quote, Value: "\""},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Bool, Value: "Visible"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Component, Value: "ButtonPrimary"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Href"},
				{Type: Colon, Value: ":"},
				{Type: Quote, Value: "\""},
				{Type: Text, Value: "https://example.com"},
				{Type: Quote, Value: "\""},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Indent, Value: "\t"},
				{Type: Text, Value: "See [demo](https://example.com)"},
				{Type: EOF},
			},
		},
		{
			name: "multiline props",
			input: `
\HeroV2
	Title:
		Title string
	Description: 
		Hero is a "component" that displays a hero 
		image with a title and description.
`,
			expected: []Token{
				{Type: Component, Value: "HeroV2"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Title"},
				{Type: Colon, Value: ":"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Indent, Value: "\t"},
				{Type: Text, Value: "Title string"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Property, Value: "Description"},
				{Type: Colon, Value: ":"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Indent, Value: "\t"},
				{Type: Text, Value: "Hero is a \"component\" that displays a hero "},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Indent, Value: "\t"},
				{Type: Text, Value: "image with a title and description."},
				{Type: EOF},
			},
		},
		{
			name: "boolean flag",
			input: `
\HeroV2
    !Visible`,
			expected: []Token{
				{Type: Component, Value: "HeroV2"},
				{Type: LineBreak, Value: "\n"},
				{Type: Indent, Value: "\t"},
				{Type: Bool, Value: "Visible"},
				{Type: EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := New(strings.TrimSpace(tt.input))
			var tokens []Token

			for {
				token := lexer.Next()
				tokens = append(tokens, token)
				if token.Type == EOF {
					break
				}
			}

			var maxLen int
			if len(tokens) < len(tt.expected) {
				maxLen = len(tt.expected)
			} else {
				maxLen = len(tokens)
			}

			for i := 0; i < maxLen; i++ {
				if i >= len(tokens) {
					t.Errorf("missing token at position %d, expected %v", i, tt.expected[i])
					continue
				}
				if i >= len(tt.expected) {
					t.Errorf("unexpected token at position %d, got %v", i, tokens[i])
					continue
				}
				expected := tt.expected[i]
				token := tokens[i]
				if token.Type != expected.Type {
					t.Errorf("Position %d, got %v want %v", i, token.Type, expected.Type)
				}
				if token.Value != expected.Value {
					t.Errorf("Position %d, got %q want %q", i, token.Value, expected.Value)
				}
			}
		})
	}
}

func TestLineAndColumn(t *testing.T) {
	input := `\HeroV2
    Title: "Hello"`

	lexer := New(strings.TrimSpace(input))
	expected := []struct {
		tokenType TokenType
		line      int
		column    int
	}{
		{Component, 1, 1},
		{LineBreak, 1, 7},
		{Indent, 2, 1},
		{Property, 2, 5},
		{Colon, 2, 10},
		{Quote, 2, 12},
		{Text, 2, 13},
		{Quote, 2, 18},
	}

	for _, exp := range expected {
		token := lexer.Next()
		if token.Type != exp.tokenType {
			t.Errorf("wrong token type, got %v want %v", token.Type, exp.tokenType)
		}
		if token.Line != exp.line {
			t.Errorf("wrong line number for %v, got %d want %d", token.Type, token.Line, exp.line)
		}
		if token.Column != exp.column {
			t.Errorf("wrong column number for %v, got %d want %d", token.Type, token.Column, exp.column)
		}
	}
}
