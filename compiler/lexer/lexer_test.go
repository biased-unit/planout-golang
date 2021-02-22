package lexer

import (
	"testing"

	"github.com/stitchfix/planout-golang/compiler/token"
)

func TestLexer_NextItem(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []token.Token
	}{
		{
			"empty string",
			"",
			[]token.Token{
				{Type: token.EOF},
			},
		},
		{
			name:  "assignment",
			input: `b = weightedChoice(choices=colors, weights=[0.2, 0.8], unit=userid);`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "b"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "weightedChoice"},
				{Type: token.LPAREN, Val: "("},
				{Type: token.IDENT, Val: "choices"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "colors"},
				{Type: token.COMMA, Val: ","},
				{Type: token.IDENT, Val: "weights"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "0.2"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "0.8"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.COMMA, Val: ","},
				{Type: token.IDENT, Val: "unit"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "userid"},
				{Type: token.RPAREN, Val: ")"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "assign array",
			input: `colors = ['#aa2200', '#22aa00', '#0022aa'];`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "colors"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.STRING, Val: "#aa2200"},
				{Type: token.COMMA, Val: ","},
				{Type: token.STRING, Val: "#22aa00"},
				{Type: token.COMMA, Val: ","},
				{Type: token.STRING, Val: "#0022aa"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "numbers",
			input: `my_list_2 = [5, 25, 2.3, 4.6e9, 0.3, 1E-4]`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "my_list_2"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "5"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "25"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "2.3"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "4.6e9"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "0.3"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "1E-4"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "strings",
			input: `"foo", 'bar'`,
			expected: []token.Token{
				{Type: token.STRING, Val: "foo"},
				{Type: token.COMMA, Val: ","},
				{Type: token.STRING, Val: "bar"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "bad string",
			input: `"foo" 'bar "baz"`,
			expected: []token.Token{
				{Type: token.STRING, Val: "foo"},
				{Type: token.ERROR, Val: "EOF while scanning string"},
			},
		},
		{
			name:  "bad number",
			input: `1 2 3 4a 5`,
			expected: []token.Token{
				{Type: token.NUMBER, Val: "1"},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.NUMBER, Val: "3"},
				{Type: token.ERROR, Val: `bad number syntax: "4a"`},
			},
		},
		{
			name: "arithmetic",
			input: `
				a = 2 + 3 + 4;
				b = 2 * 3 * 4;
				c = -2;
				d = 2 + 3 - 4;
				e = 4 % 2;
				f = 4 / 2;
				g = round(2.3);
				x = min(1, 2, -4);
				`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "a"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.ADD, Val: "+"},
				{Type: token.NUMBER, Val: "3"},
				{Type: token.ADD, Val: "+"},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "b"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.MUL, Val: "*"},
				{Type: token.NUMBER, Val: "3"},
				{Type: token.MUL, Val: "*"},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "c"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.SUB, Val: "-"},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "d"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.ADD, Val: "+"},
				{Type: token.NUMBER, Val: "3"},
				{Type: token.SUB, Val: "-"},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "e"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.REM, Val: "%"},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "f"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.QUO, Val: "/"},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "g"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "round"},
				{Type: token.LPAREN, Val: "("},
				{Type: token.NUMBER, Val: "2.3"},
				{Type: token.RPAREN, Val: ")"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "x"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "min"},
				{Type: token.LPAREN, Val: "("},
				{Type: token.NUMBER, Val: "1"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.COMMA, Val: ","},
				{Type: token.SUB, Val: "-"},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.RPAREN, Val: ")"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name: "returns",
			input: `
x = 5;
return x;
`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "x"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "5"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.RETURN, Val: "return"},
				{Type: token.IDENT, Val: "x"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name: "comment",
			input: `
####### MY PLANOUT SCRIPT ########
    a = 5; # assign a to 5
b = 'b';
# this is not read
`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "a"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.NUMBER, Val: "5"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.IDENT, Val: "b"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.STRING, Val: "b"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name: "logical operators",
			input: `
x = a && b;
y = a || b || c;
y <- !b;
`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "x"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "a"},
				{Type: token.AND, Val: "&&"},
				{Type: token.IDENT, Val: "b"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.IDENT, Val: "y"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "a"},
				{Type: token.OR, Val: "||"},
				{Type: token.IDENT, Val: "b"},
				{Type: token.OR, Val: "||"},
				{Type: token.IDENT, Val: "c"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.IDENT, Val: "y"},
				{Type: token.ARROW, Val: "<-"},
				{Type: token.NOT, Val: "!"},
				{Type: token.IDENT, Val: "b"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "empty json map literal",
			input: `a = @{};`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "a"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.JSON, Val: "{}"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "empty json array literal",
			input: `a = @[];`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "a"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.JSON, Val: "[]"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "json map literal",
			input: `a = @{"foo":1, "bar": [2,3]};`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "a"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.JSON, Val: "{\"foo\":1, \"bar\": [2,3]}"},
				{Type: token.SEMICOLON, Val: ";"},
				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "json array literal",
			input: `@[1, 2, 3]`,
			expected: []token.Token{
				{Type: token.JSON, Val: "[1, 2, 3]"},
			},
		},
		{
			name:  "nested json array literal",
			input: `@[1, 2, [3], {"four": 5.5}]`,
			expected: []token.Token{
				{Type: token.JSON, Val: `[1, 2, [3], {"four": 5.5}]`},
			},
		},
		{
			name: "arrays",
			input: `
a = [4, 3.14, 'foo'];
b = [a, 0.0, 3];      
x = a[0];
y = b[0][2];
z = b[22];
`,
			expected: []token.Token{
				{Type: token.IDENT, Val: "a"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "4"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "3.14"},
				{Type: token.COMMA, Val: ","},
				{Type: token.STRING, Val: "foo"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "b"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.IDENT, Val: "a"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "0.0"},
				{Type: token.COMMA, Val: ","},
				{Type: token.NUMBER, Val: "3"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "x"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "a"},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "0"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "y"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "b"},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "0"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "2"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.IDENT, Val: "z"},
				{Type: token.ASSIGN, Val: "="},
				{Type: token.IDENT, Val: "b"},
				{Type: token.LBRACKET, Val: "["},
				{Type: token.NUMBER, Val: "22"},
				{Type: token.RBRACKET, Val: "]"},
				{Type: token.SEMICOLON, Val: ";"},

				{Type: token.EOF, Val: ""},
			},
		},
		{
			name:  "invalid number",
			input: `5.5.5; .2 4. 19.3_`,
			expected: []token.Token{
				{Type: token.NUMBER, Val: "5.5"},
				{Type: token.ERROR, Val: "unexpected character: '.'"},
			},
		},
		{
			name:  "invalid number",
			input: `.2`,
			expected: []token.Token{
				{Type: token.ERROR, Val: "unexpected character: '.'"},
			},
		},
		{
			name:  "invalid number",
			input: `19.3_`,
			expected: []token.Token{
				{Type: token.ERROR, Val: "bad number syntax: \"19.3_\""},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			l := New(testCase.input)

			for i := range testCase.expected {
				item := l.NextToken()

				if item.Type != testCase.expected[i].Type {
					t.Errorf("expected[%d] - type wrong. expected=%q, got=%q", i, testCase.expected[i].Type, item.Type)
				}

				if item.Val != testCase.expected[i].Val {
					t.Errorf("expected[%d] - value wrong. expected=%q, got=%q", i, testCase.expected[i].Val, item.Val)
				}
			}
		})
	}
}
