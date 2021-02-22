package ast

import (
	"github.com/stitchfix/planout-golang/compiler/token"
)

// Statement denotes a top-level node such as an if statement, assignment statement, or return statement
type Statement interface {
	statementNode()
}

// Expression denotes an internal node, such as numeric or string
// literals, infix operator expressions, or identifier expressions
type Expression interface {
	expressionNode()
}

func NewProgram() *Program {
	return &Program{
		Op:  "seq",
		Seq: make([]Statement, 0),
	}
}

type Program struct {
	Op  string      `json:"op"`
	Seq []Statement `json:"seq"`
}

func NewSwitchStatement(cases ...Case) *SwitchStatement {
	result := &SwitchStatement{
		Op:    "switch",
		Cases: make([]Case, 0),
	}

	for _, cs := range cases {
		result.Cases = append(result.Cases, Case{
			Op:        "case",
			Condition: cs.Condition,
			Result:    cs.Result,
		})
	}

	return result
}

type SwitchStatement struct {
	Op    string `json:"op"`
	Cases []Case `json:"cases"`
}

func (se *SwitchStatement) statementNode() {}

type Case struct {
	Op        string     `json:"op"`
	Condition Expression `json:"condidion"` // This typo is present in the language definition
	Result    Statement  `json:"result"`
}

func NewAssignmentStatement(varName string, value Expression) *AssignmentStatement {
	return &AssignmentStatement{
		Op:    "set",
		Var:   varName,
		Value: value,
	}
}

type AssignmentStatement struct {
	Op    string     `json:"op"`
	Var   string     `json:"var"`
	Value Expression `json:"value"`
}

func (s *AssignmentStatement) statementNode() {}

func NewReturnStatement(value Expression) *ReturnStatement {
	return &ReturnStatement{
		Op:    "return",
		Value: value,
	}
}

type ReturnStatement struct {
	Op    string     `json:"op"`
	Value Expression `json:"value"`
}

func (s *ReturnStatement) statementNode() {}

func NewIfStatement(cond ...Conditional) *IfStatement {
	if cond == nil {
		cond = make([]Conditional, 0)
	}
	return &IfStatement{
		Op:   "cond",
		Cond: cond,
	}
}

type IfStatement struct {
	Op   string        `json:"op"`
	Cond []Conditional `json:"cond"`
}

func (ie *IfStatement) statementNode() {}

type Conditional struct {
	Condition   Expression      `json:"if"`
	Consequence *BlockStatement `json:"then"`
}

func NewBlockStatement(seq ...Statement) *BlockStatement {
	if seq == nil {
		seq = make([]Statement, 0)
	}
	return &BlockStatement{
		Op:  "seq",
		Seq: seq,
	}
}

type BlockStatement struct {
	Op  string      `json:"op"`
	Seq []Statement `json:"seq"`
}

func NewPrefixExpression(tok token.Type, right Expression) Expression {
	switch tok {
	default:
		return nil
	case token.SUB:
		return &PrefixExpression{
			Op:    "negative",
			Value: right,
		}
	case token.NOT:
		return &PrefixExpression{
			Op:    "not",
			Value: right,
		}
	}
}

type PrefixExpression struct {
	Op    string     `json:"op"`
	Value Expression `json:"value"`
}

func (s *PrefixExpression) expressionNode() {}

func NewInfixExpression(tok token.Type, left, right Expression) Expression {
	switch tok {
	default:
		return nil
	case token.REM:
		return &InfixExpressionLeftRight{
			Op:    "%",
			Left:  left,
			Right: right,
		}
	case token.QUO:
		return &InfixExpressionLeftRight{
			Op:    "/",
			Left:  left,
			Right: right,
		}
	case token.GTR:
		return &InfixExpressionLeftRight{
			Op:    ">",
			Left:  left,
			Right: right,
		}
	case token.LSS:
		return &InfixExpressionLeftRight{
			Op:    "<",
			Left:  left,
			Right: right,
		}
	case token.EQUALS:
		return &InfixExpressionLeftRight{
			Op:    "equals",
			Left:  left,
			Right: right,
		}
	case token.NEQ:
		return &PrefixExpression{
			Op:    "not",
			Value: NewInfixExpression(token.EQUALS, left, right),
		}
	case token.LTE:
		return &InfixExpressionLeftRight{
			Op:    "<=",
			Left:  left,
			Right: right,
		}
	case token.GTE:
		return &InfixExpressionLeftRight{
			Op:    ">=",
			Left:  left,
			Right: right,
		}
	case token.ADD:
		return &InfixExpressionValues{
			Op:     "sum",
			Values: [2]Expression{left, right},
		}
	case token.SUB:
		return &InfixExpressionValues{
			Op: "sum",
			Values: [2]Expression{
				left,
				NewPrefixExpression(token.SUB, right)},
		}
	case token.MUL:
		return &InfixExpressionValues{
			Op:     "product",
			Values: [2]Expression{left, right},
		}
	case token.OR:
		return &InfixExpressionValues{
			Op:     "or",
			Values: [2]Expression{left, right},
		}
	case token.AND:
		return &InfixExpressionValues{
			Op:     "and",
			Values: [2]Expression{left, right},
		}
	case token.COALESCE:
		return &InfixExpressionValues{
			Op:     "coalesce",
			Values: [2]Expression{left, right},
		}
	}
}

type InfixExpressionLeftRight struct {
	Op    string     `json:"op"`
	Left  Expression `json:"left"`
	Right Expression `json:"right"`
}

func (s *InfixExpressionLeftRight) expressionNode() {}

type InfixExpressionValues struct {
	Op     string        `json:"op"`
	Values [2]Expression `json:"values"`
}

func (s *InfixExpressionValues) expressionNode() {}

func NewIdentifierExpression(varName string) *Identifier {
	return &Identifier{
		Op:  "get",
		Var: varName,
	}
}

type Identifier struct {
	Op  string `json:"op"`
	Var string `json:"var"`
}

func (s *Identifier) expressionNode() {}

type IntegerLiteral int64

func (il IntegerLiteral) expressionNode() {}

type FloatLiteral float64

func (il FloatLiteral) expressionNode() {}

type StringLiteral string

func (il StringLiteral) expressionNode() {}

type Boolean bool

func (bl Boolean) expressionNode() {}

func NewArrayLiteral(vals []Expression) *ArrayLiteral {
	if vals == nil {
		vals = make([]Expression, 0)
	}
	return &ArrayLiteral{
		Op:     "array",
		Values: vals,
	}
}

type ArrayLiteral struct {
	Op     string       `json:"op"`
	Values []Expression `json:"values"`
}

func (al *ArrayLiteral) expressionNode() {}

func NewNull() *NullLiteral {
	return &NullLiteral{}
}

type NullLiteral struct{}

func (nl *NullLiteral) expressionNode() {}
func (nl *NullLiteral) MarshalJSON() ([]byte, error) {
	return []byte(`null`), nil
}

func NewJSONLiteral(val interface{}) *JSONLiteral {
	return &JSONLiteral{
		Op:    "literal",
		Value: val,
	}
}

// JSONLiteral can be a JSON array, map, string, or number
type JSONLiteral struct {
	Op    string      `json:"op"`
	Value interface{} `json:"value"`
}

func (jl *JSONLiteral) expressionNode() {}

func NewIndexExpression(base, index Expression) *IndexExpression {
	return &IndexExpression{
		Op:    "index",
		Base:  base,
		Index: index,
	}
}

type IndexExpression struct {
	Op    string     `json:"op"`
	Base  Expression `json:"base"`
	Index Expression `json:"index"`
}

func (ie *IndexExpression) expressionNode() {}

func NewFunctionCall(fName string, args ...Expression) Expression {
	switch len(args) {
	default:
		return &FunctionCallManyArgs{Op: fName, Values: args}
	case 0:
		return &FunctionCallNoArgs{Op: fName}
	case 1:
		return &FunctionCallOneArg{Op: fName, Value: args[0]}
	}
}

func NewFunctionCallNamedArgs(fName string, args map[string]Expression) Expression {
	if args == nil {
		args = make(map[string]Expression)
	}
	x := NamedArgs(args)
	x["op"] = StringLiteral(fName)
	return &x
}

type FunctionCallNoArgs struct {
	Op string `json:"op"`
}

func (fca *FunctionCallNoArgs) expressionNode() {}

type FunctionCallOneArg struct {
	Value Expression `json:"value"`
	Op    string     `json:"op"`
}

func (fco *FunctionCallOneArg) expressionNode() {}

type FunctionCallManyArgs struct {
	Values []Expression `json:"values"`
	Op     string       `json:"op"`
}

func (fcm *FunctionCallManyArgs) expressionNode() {}

type NamedArgs map[string]Expression

func (fcn *NamedArgs) expressionNode() {}
