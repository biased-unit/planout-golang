package compiler

import (
	"io/ioutil"
	"testing"

	"github.com/biased-unit/planout-golang/compiler/lexer"
	"github.com/biased-unit/planout-golang/compiler/parser"
	"github.com/stretchr/testify/require"
)

// Test cases verified using: http://planout-editor.herokuapp.com/
func TestCompiler_Run(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"empty script",
			``,
			`{}`,
		},
		{
			"assign int",
			"x=5;",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":5}]}`,
		},
		{
			"assign float",
			"x=3.14;",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":3.14}]}`,
		},
		{
			"assign exponential",
			"my_exp = 3.14E-9;",
			`{"op":"seq","seq":[{"op":"set","var":"my_exp","value":3.14e-9}]}`,
		},
		{
			"return value",
			"return x;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"x"}}]}`,
		},
		{
			"assign identifier",
			"y = x;",
			`{"op":"seq","seq":[{"op":"set","var":"y","value":{"op":"get","var":"x"}}]}`,
		},
		{
			"prefix not identifier",
			"y=!x;",
			`{"op":"seq","seq":[{"op":"set","var":"y","value":{"op":"not","value":{"op":"get","var":"x"}}}]}`,
		},
		{
			"prefix neg number",
			"x = -5.5;",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":-5.5}]}`,
		},
		{
			"prefix neg identifier",
			"z = -y;",
			`{"op":"seq","seq":[{"op":"set","var":"z","value":{"op":"negative","value":{"op":"get","var":"y"}}}]}`,
		},
		{
			"infix plus",
			"return 5 + 5;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"sum","values":[5,5]}}]}`,
		},
		{
			"infix minus",
			"x = 5 - 5;",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"sum","values":[5,{"op":"negative","value":5}]}}]}`,
		},
		{
			"grouped expression",
			"x = (5 - 5) * 10",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"product","values":[{"op":"sum","values":[5,{"op":"negative","value":5}]},10]}}]}`,
		},
		{
			"prefix and infix",
			"return -a * b;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"negative","value":{"op":"product","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]}}}]}`,
		},

		{
			"double prefix",
			"return !-a;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"not","value":{"op":"negative","value":{"op":"get","var":"a"}}}}]}`,
		},
		{
			"associative",
			"return a + b + c;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"sum","values":[{"op":"sum","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]},{"op":"get","var":"c"}]}}]}`,
		},

		{
			"associative",
			"return a + b - c;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"sum","values":[{"op":"sum","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]},{"op":"negative","value":{"op":"get","var":"c"}}]}}]}`,
		},

		{
			"associative",
			"return a * b * c;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"product","values":[{"op":"product","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]},{"op":"get","var":"c"}]}}]}`,
		},

		{
			"associative",
			"return a * b / c;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"/","left":{"op":"product","values":[{"op":"get","var":"a"},{"op":"get","var":"b"}]},"right":{"op":"get","var":"c"}}}]}`,
		},

		{
			"associative",
			"return a + b * c + d / e - f;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"sum","values":[{"op":"sum","values":[{"op":"sum","values":[{"op":"get","var":"a"},{"op":"product","values":[{"op":"get","var":"b"},{"op":"get","var":"c"}]}]},{"op":"/","left":{"op":"get","var":"d"},"right":{"op":"get","var":"e"}}]},{"op":"negative","value":{"op":"get","var":"f"}}]}}]}`,
		},

		{
			"multiple expressions",
			"x = 3 + 4; return -x * 5;",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"sum","values":[3,4]}},{"op":"return","value":{"op":"negative","value":{"op":"product","values":[{"op":"get","var":"x"},5]}}}]}`,
		},
		{
			"multiple comparisons",
			"return 5 > 4 == 3 < 4;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"<","left":{"op":"equals","left":{"op":">","left":5,"right":4},"right":3},"right":4}}]}`,
		},

		{
			"multiple comparisons",
			"return 5 < 4 != 3 > 4",
			`{"op":"seq","seq":[{"op":"return","value":{"op":">","left":{"op":"not","value":{"op":"equals","left":{"op":"<","left":5,"right":4},"right":3}},"right":4}}]}`,
		},
		{
			"logical with math",
			"result = 3 + 4 * 5 == 3 * 1 + 4 *5;",
			`{"op":"seq","seq":[{"op":"set","var":"result","value":{"op":"equals","left":{"op":"sum","values":[3,{"op":"product","values":[4,5]}]},"right":{"op":"sum","values":[{"op":"product","values":[3,1]},{"op":"product","values":[4,5]}]}}}]}`,
		},
		{
			"logical operators",
			`x = true;
				y = false;
				z = x || y;`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":true},{"op":"set","var":"y","value":false},{"op":"set","var":"z","value":{"op":"or","values":[{"op":"get","var":"x"},{"op":"get","var":"y"}]}}]}`,
		},
		{
			"logical and comparison",
			"return 3 > 5 == false;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"equals","left":{"op":">","left":3,"right":5},"right":false}}]}`,
		},
		{
			"logical and comparison",
			"return 3 < 5 == true;",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"equals","left":{"op":"<","left":3,"right":5},"right":true}}]}`,
		},
		{
			"if expression",
			"if (x > 5) { return y; }",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":{"op":">","left":{"op":"get","var":"x"},"right":5},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"y"}}]}}]}]}`,
		},
		{
			"if else",
			"if (x > 5) { return y; } else { z=9; }",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":{"op":">","left":{"op":"get","var":"x"},"right":5},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"y"}}]}},{"if":true,"then":{"op":"seq","seq":[{"op":"set","var":"z","value":9}]}}]}]}`,
		},
		{
			"if else if",
			"if (x>5) { return y; } else if (x == 6 || x == 7) { z = 9; }",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":{"op":">","left":{"op":"get","var":"x"},"right":5},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"y"}}]}},{"if":{"op":"or","values":[{"op":"equals","left":{"op":"get","var":"x"},"right":6},{"op":"equals","left":{"op":"get","var":"x"},"right":7}]},"then":{"op":"seq","seq":[{"op":"set","var":"z","value":9}]}}]}]}`,
		},
		{
			"if else chained",
			"if (x == 5) { return y; } else if (x == 6) { return y + 1; } else if (x == 7) { return -y; } else { return x; }",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":5},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"y"}}]}},{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":6},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"sum","values":[{"op":"get","var":"y"},1]}}]}},{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":7},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"negative","value":{"op":"get","var":"y"}}}]}},{"if":true,"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"x"}}]}}]}]}`,
		},
		{
			"if else chained",
			"if (x == 5) { return y; } else if (x == 6) { return y + 1; } else if (x == 7) { return -y; } else if (x == true) { return x; }",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":5},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"y"}}]}},{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":6},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"sum","values":[{"op":"get","var":"y"},1]}}]}},{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":7},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"negative","value":{"op":"get","var":"y"}}}]}},{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":true},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"x"}}]}}]}]}`,
		},
		{
			"multiple if statements",
			"if (x==5) { return y; } if (z==7) { x = 9; }",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":{"op":"equals","left":{"op":"get","var":"x"},"right":5},"then":{"op":"seq","seq":[{"op":"return","value":{"op":"get","var":"y"}}]}}]},{"op":"cond","cond":[{"if":{"op":"equals","left":{"op":"get","var":"z"},"right":7},"then":{"op":"seq","seq":[{"op":"set","var":"x","value":9}]}}]}]}`,
		},
		{
			"empty if block",
			"if (true) {}",
			`{"op":"seq","seq":[{"op":"cond","cond":[{"if":true,"then":{"op":"seq","seq":[]}}]}]}`,
		},
		{
			"empty switch",
			"switch {}",
			`{"op":"seq","seq":[{"op":"switch","cases":[]}]}`,
		},
		{
			"switch case",
			"switch { x < 5 => if (true) { y = 6; }; x > 5 => return z; }",
			`{"op":"seq","seq":[{"op":"switch","cases":[{"op":"case","condidion":{"op":"<","left":{"op":"get","var":"x"},"right":5},"result":{"op":"cond","cond":[{"if":true,"then":{"op":"seq","seq":[{"op":"set","var":"y","value":6}]}}]}},{"op":"case","condidion":{"op":">","left":{"op":"get","var":"x"},"right":5},"result":{"op":"return","value":{"op":"get","var":"z"}}}]}]}`,
		},
		{
			"switch case then return",
			"switch { x < 5 => if (true) { y = 6; }; x > 5 => return z; } return 9;",
			`{"op":"seq","seq":[{"op":"switch","cases":[{"op":"case","condidion":{"op":"<","left":{"op":"get","var":"x"},"right":5},"result":{"op":"cond","cond":[{"if":true,"then":{"op":"seq","seq":[{"op":"set","var":"y","value":6}]}}]}},{"op":"case","condidion":{"op":">","left":{"op":"get","var":"x"},"right":5},"result":{"op":"return","value":{"op":"get","var":"z"}}}]},{"op":"return","value":9}]}`,
		},
		{
			"switch case then return with semicolon",
			"switch { x < 5 => if (true) { y = 6; }; x > 5 => return z; }; return 9;",
			`{"op":"seq","seq":[{"op":"switch","cases":[{"op":"case","condidion":{"op":"<","left":{"op":"get","var":"x"},"right":5},"result":{"op":"cond","cond":[{"if":true,"then":{"op":"seq","seq":[{"op":"set","var":"y","value":6}]}}]}},{"op":"case","condidion":{"op":">","left":{"op":"get","var":"x"},"right":5},"result":{"op":"return","value":{"op":"get","var":"z"}}}]},{"op":"return","value":9}]}`,
		},
		{
			"empty array",
			"x = [];",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"array","values":[]}}]}`,
		},
		{
			"assign array",
			"x = [1, 2, '3', four]",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"array","values":[1,2,"3",{"op":"get","var":"four"}]}}]}`,
		},
		{
			"null",
			"x = null;",
			`{"op":"seq","seq":[{"op":"set","var":"x","value":null}]}`,
		},
		{
			"empty json literal",
			`x = @{};`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"literal","value":{}}}]}`,
		},
		{
			"json literal",
			`x = @{"a": 1};`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"literal","value":{"a":1}}}]}`,
		},
		{
			"json literal",
			`x = @{"a": {"b": 2}, "c": [3, 4, 5.5]};`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"literal","value":{"a":{"b":2},"c":[3,4,5.5]}}}]}`,
		},
		{
			"json literal scientific notation",
			`x = @{"my_var": 3.14E-09};`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"literal","value":{"my_var":3.14e-9}}}]}`,
		},
		{
			"json literal string",
			`x = @"i am a JSON string";`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"literal","value":"i am a JSON string"}}]}`,
		},
		{
			"json literal number",
			`x = @123.4e-09;`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"literal","value":1.234e-07}}]}`,
		},
		{
			"json literal array",
			`return @[1, 2, 3];`,
			`{"op":"seq","seq":[{"op":"return","value":{"op":"literal","value":[1,2,3]}}]}`,
		},
		{
			"empty JSON literal array",
			"return @[];",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"literal","value":[]}}]}`,
		},
		{
			"JSON null",
			"return @null",
			`{"op":"seq","seq":[{"op":"return","value":{"op":"literal","value":null}}]}`,
		},
		{
			"index expression",
			`x = [1,2,3][0]`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"op":"index","base":{"op":"array","values":[1,2,3]},"index":0}}]}`,
		},
		{
			"function call no args",
			`return myFunc();`,
			`{"op":"seq","seq":[{"op":"return","value":{"op":"myFunc"}}]}`,
		},
		{
			"function call one arg",
			`y = myFunc(3);`,
			`{"op":"seq","seq":[{"op":"set","var":"y","value":{"value":3,"op":"myFunc"}}]}`,
		},
		{
			"function call several args",
			`x =  myFunc(1, "2", x);`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"values":[1,"2",{"op":"get","var":"x"}],"op":"myFunc"}}]}`,
		},
		{
			"function call complicated args",
			`x =  myFunc(@{"a": 1}, 3 - 4 / 2);`,
			`{"op":"seq","seq":[{"op":"set","var":"x","value":{"values":[{"op":"literal","value":{"a":1}},{"op":"sum","values":[3,{"op":"negative","value":{"op":"/","left":4,"right":2}}]}],"op":"myFunc"}}]}`,
		},
		{
			"function call one named arg",
			`return hello(str="world");`,
			`{"op":"seq","seq":[{"op":"return","value":{"str":"world","op":"hello"}}]}`,
		},
		{
			"function call named args",
			`result = myFunc(a=c, x="y");`,
			`{"op":"seq","seq":[{"op":"set","var":"result","value":{"a":{"op":"get","var":"c"},"x":"y","op":"myFunc"}}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := lexer.New(tt.input)
			p := parser.New(s)
			c := New(p)
			b, err := c.Run()
			if err != nil {
				t.Error(err)
			}
			require.JSONEq(t, tt.expected, string(b))
		})
	}
}

func TestCompiler_RunFixtures(t *testing.T) {
	tests := []struct {
		inputFile    string
		expectedFile string
	}{
		{"testdata/exp1.planout", "testdata/exp1.json"},
		{"testdata/exp2.planout", "testdata/exp2.json"},
		{"testdata/exp3.planout", "testdata/exp3.json"},
		{"testdata/exp4.planout", "testdata/exp4.json"},
		{"testdata/exp5.planout", "testdata/exp5.json"},
		{"testdata/exp6.planout", "testdata/exp6.json"},
		{"testdata/exp7.planout", "testdata/exp7.json"},
		{"testdata/exp8.planout", "testdata/exp8.json"},
	}

	for _, tc := range tests {
		t.Run(tc.inputFile, func(t *testing.T) {
			input, err := ioutil.ReadFile(tc.inputFile)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := ioutil.ReadFile(tc.expectedFile)
			if err != nil {
				t.Fatal(err)
			}

			lx := lexer.New(string(input))
			ps := parser.New(lx)
			comp := New(ps)
			actual, err := comp.Run()
			if err != nil {
				t.Fatal(err)
			}
			require.JSONEq(t, string(expected), string(actual))
		})

	}
}
