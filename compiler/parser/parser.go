// Package parser implements top-down operator precedence (Pratt) parsing with rules derived from
// https://github.com/facebook/planout/blob/master/compiler/planout.jison
package parser

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/biased-unit/planout-golang/compiler/ast"
	"github.com/biased-unit/planout-golang/compiler/token"
)

// precedence levels
const (
	_ int = iota
	LOWEST
	NOT        // !
	LOGICAL    // OR, AND, COALESCE
	COMPARISON // ==, !=, <=, >=, >, <
	SUM        // +, -
	PROD       // *, /, %
	CALL       // (
	INDEX      // [
)

// operator precedence table
var precedences = map[token.Type]int{
	token.NOT:      NOT,
	token.OR:       LOGICAL,
	token.AND:      LOGICAL,
	token.COALESCE: LOGICAL,
	token.EQUALS:   COMPARISON,
	token.NEQ:      COMPARISON,
	token.LTE:      COMPARISON,
	token.GTE:      COMPARISON,
	token.GTR:      COMPARISON,
	token.LSS:      COMPARISON,
	token.ADD:      SUM,
	token.SUB:      SUM,
	token.MUL:      PROD,
	token.QUO:      PROD,
	token.REM:      PROD,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

func New(lx tokenLexer) *Parser {
	p := &Parser{
		lx: lx,
	}

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.NUMBER, p.parseNumericLiteral)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.SUB, p.parsePrefixSub)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.NULL, p.parseNull)
	p.registerPrefix(token.JSON, p.parseJSONLiteral)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.REM, p.parseInfixExpression)
	p.registerInfix(token.QUO, p.parseInfixExpression)
	p.registerInfix(token.GTR, p.parseInfixExpression)
	p.registerInfix(token.LSS, p.parseInfixExpression)
	p.registerInfix(token.EQUALS, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.ADD, p.parseInfixExpression)
	p.registerInfix(token.SUB, p.parseInfixExpression)
	p.registerInfix(token.MUL, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.COALESCE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.nextToken()
	p.nextToken()
	return p
}

type Parser struct {
	lx tokenLexer

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.Type]prefixParseFn
	infixParseFns  map[token.Type]infixParseFn

	errors []error
}

type tokenLexer interface {
	NextToken() token.Token
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type LexingError token.Token

func (le LexingError) Error() string {
	return fmt.Sprintf("%d: %s", le.Line, le.Val)
}

type ParsingError struct {
	Tok token.Token
	Err error
}

func (pe ParsingError) Error() string {
	return fmt.Sprintf("%d: %s", pe.Tok.Line, pe.Err.Error())
}

func (p *Parser) ParseProgram() *ast.Program {
	program := ast.NewProgram()

	for !p.curTokenIs(token.EOF) {
		exp := p.parseStatement()
		if exp == nil || len(p.errors) > 0 {
			return program
		}
		program.Seq = append(program.Seq, exp)
		p.nextToken()
		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	return program
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lx.NextToken()
}

func (p *Parser) registerPrefix(itemType token.Type, fn prefixParseFn) {
	p.prefixParseFns[itemType] = fn
}

func (p *Parser) registerInfix(itemType token.Type, fn infixParseFn) {
	p.infixParseFns[itemType] = fn
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.IF:
		return p.parseIfStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.IDENT:
		return p.parseAssignmentStatement()
	default:
		p.tokTypeError(p.curToken, token.IDENT, token.IF, token.RETURN, token.SWITCH)
		return nil
	}
}

func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
	var cases []ast.Case

	if !p.accept(token.LBRACE) {
		p.tokTypeError(p.curToken, token.LBRACE)
		return nil
	}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		cond := p.parseSimpleExpression(LOWEST)

		if !p.accept(token.THEN) {
			p.tokTypeError(p.peekToken, token.THEN)
			return nil
		}
		p.nextToken()

		res := p.parseStatement()
		cases = append(cases, ast.Case{Condition: cond, Result: res})

		p.nextToken()

		if p.curTokenIs(token.SEMICOLON) {
			p.nextToken()
		}
	}

	return ast.NewSwitchStatement(cases...)
}

// parseIfStatement iteratively parses a chain of if/else statements
func (p *Parser) parseIfStatement() *ast.IfStatement {

	var conds []ast.Conditional

	for !p.curTokenIs(token.EOF) {

		// loop should always start on the IF token so next token must be open paren
		if !p.accept(token.LPAREN) {
			p.tokTypeError(p.peekToken, token.LPAREN)
			return nil
		}

		// move into the predicate and parse the expression
		p.nextToken()
		cond := p.parseSimpleExpression(LOWEST)
		if !p.accept(token.RPAREN) {
			p.tokTypeError(p.peekToken, token.RPAREN)
			return nil
		}

		// move into the consequence and parse the block expression
		if !p.accept(token.LBRACE) {
			p.tokTypeError(p.peekToken, token.LBRACE)
			return nil
		}
		cons := p.parseBlockStatement()
		conds = append(conds, ast.Conditional{Condition: cond, Consequence: cons})

		// If there's no else block, we are done
		if !p.accept(token.ELSE) {
			break
		}

		// If we find an "else if", move to the "if" token and continue
		// If we find a "{" it's the start of the terminal else block
		switch p.peekToken.Type {
		default:
			p.tokTypeError(p.peekToken, token.IF, token.LBRACE)
		case token.IF:
			p.nextToken()
		case token.LBRACE:
			p.nextToken()
			alt := p.parseBlockStatement()
			conds = append(conds, ast.Conditional{Condition: ast.Boolean(true), Consequence: alt})
			return ast.NewIfStatement(conds...)
		}
	}

	return ast.NewIfStatement(conds...)
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	if p.accept(token.EOF) {
		p.errors = append(p.errors, ParsingError{
			Tok: p.curToken,
			Err: fmt.Errorf("EOF while parsing block statement"),
		})
		return nil
	}

	p.nextToken()

	var seq []ast.Statement
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		exp := p.parseStatement()
		if exp != nil {
			seq = append(seq, exp)
		}
		switch {
		default:
			p.tokTypeError(p.curToken, token.SEMICOLON, token.RBRACE)
			return nil
		case p.curTokenIs(token.SEMICOLON):
			if p.accept(token.EOF) {
				p.errors = append(p.errors, ParsingError{
					Tok: p.curToken,
					Err: fmt.Errorf("EOF while parsing block statement"),
				})
				return nil
			}
			p.nextToken()
		case p.curTokenIs(token.RBRACE):
			p.nextToken()
		}
	}
	return ast.NewBlockStatement(seq...)
}

func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {

	varName := p.curToken.Val

	if !p.accept(token.ASSIGN, token.ARROW) {
		p.tokTypeError(p.peekToken, token.ASSIGN, token.ARROW)
		return nil
	}

	p.nextToken() // skip the ASSIGN or ARROW token

	// Now parse the expression to the right of the assignment operator
	value := p.parseSimpleExpression(LOWEST)

	if !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return ast.NewAssignmentStatement(varName, value)
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	p.nextToken()
	value := p.parseSimpleExpression(LOWEST)

	if !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return ast.NewReturnStatement(value)
}

// parseSimpleExpression implements Pratt parsing to correctly
// handle expressions with multiple operators of specified precedence
func (p *Parser) parseSimpleExpression(precedence int) ast.Expression {

	if p.curTokenIs(token.ERROR) {
		p.errors = append(p.errors, LexingError(p.curToken))
		return nil
	}

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	tokType := p.curToken.Type
	tokPrecedence := p.curPrecedence()
	p.nextToken()
	right := p.parseSimpleExpression(tokPrecedence)
	return ast.NewPrefixExpression(tokType, right)
}

// parsePrefixSub a special case of parsePrefixExpression, because
// a prefix SUB token (the '-' operator) is parsed differently depending on the type of the operand
func (p *Parser) parsePrefixSub() ast.Expression {
	tokType := p.curToken.Type
	tokPrecedence := p.curPrecedence()
	p.nextToken()
	right := p.parseSimpleExpression(tokPrecedence)
	switch v := right.(type) {
	default:
		return ast.NewPrefixExpression(tokType, right)
	case ast.IntegerLiteral:
		return ast.IntegerLiteral(-1 * int64(v))
	case ast.FloatLiteral:
		return ast.FloatLiteral(-1 * float64(v))
	}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	tokType := p.curToken.Type
	precedence := p.curPrecedence()
	p.nextToken()
	right := p.parseSimpleExpression(precedence)

	return ast.NewInfixExpression(tokType, left, right)
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseSimpleExpression(LOWEST)

	if !p.accept(token.RPAREN) {
		p.tokTypeError(p.peekToken, token.RPAREN)
		return nil
	}

	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	p.nextToken()
	vals := p.parseSimpleExpressionList(token.RBRACKET)
	return ast.NewArrayLiteral(vals)
}

func (p *Parser) parseSimpleExpressionList(closeToken token.Type) []ast.Expression {
	var exps []ast.Expression

	if p.curTokenIs(closeToken) {
		return exps
	}

	exps = append(exps, p.parseSimpleExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		exps = append(exps, p.parseSimpleExpression(LOWEST))
	}

	if !p.accept(closeToken) {
		p.tokTypeError(p.peekToken, closeToken)
		return nil
	}
	return exps
}

// parseNumericLiteral tries to convert the current item'lx value
// to an int or a float. If it can't be cast to either, appends an error message and returns nil
func (p *Parser) parseNumericLiteral() ast.Expression {
	if iVal, err := p.convIntLiteral(p.curToken.Val); err == nil {
		return iVal
	}

	if fVal, err := p.convFloatLiteral(p.curToken.Val); err == nil {
		return fVal
	}

	p.errors = append(p.errors, ParsingError{
		Tok: p.curToken,
		Err: fmt.Errorf("not a valid number: %s", p.curToken.Val),
	})
	return nil
}

func (p *Parser) convIntLiteral(strVal string) (ast.Expression, error) {
	value, err := strconv.ParseInt(strVal, 0, 64)
	if err != nil {
		return nil, err
	}

	return ast.IntegerLiteral(value), nil
}

func (p *Parser) convFloatLiteral(strVal string) (ast.Expression, error) {
	value, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return nil, err
	}

	return ast.FloatLiteral(value), nil
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return ast.StringLiteral(p.curToken.Val)
}

func (p *Parser) parseBoolean() ast.Expression {
	return ast.Boolean(p.curTokenIs(token.TRUE))
}

func (p *Parser) parseIdentifier() ast.Expression {
	return ast.NewIdentifierExpression(p.curToken.Val)
}

func (p *Parser) parseNull() ast.Expression {
	return ast.NewNull()
}

func (p *Parser) parseJSONLiteral() ast.Expression {
	var val interface{}
	err := json.Unmarshal([]byte(p.curToken.Val), &val)
	if err != nil {
		parseError := ParsingError{
			Tok: p.curToken,
			Err: fmt.Errorf("failed to parse JSON literal %s: %w", p.curToken.Val, err),
		}
		p.errors = append(p.errors, parseError)
		return nil
	}
	return ast.NewJSONLiteral(val)
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	p.nextToken()
	ind := p.parseSimpleExpression(LOWEST)
	if !p.accept(token.RBRACKET) {
		p.tokTypeError(p.peekToken, token.RBRACKET)
		return nil
	}
	return ast.NewIndexExpression(left, ind)
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	funcIdentifier, ok := left.(*ast.Identifier)
	if !ok {
		parseError := ParsingError{
			Tok: p.curToken,
			Err: fmt.Errorf("%s ( %s", left, "function-call syntax with non-identifier expression"),
		}
		p.errors = append(p.errors, parseError)
		return nil
	}

	if p.peekTokenIs(token.RPAREN) {
		return ast.NewFunctionCall(funcIdentifier.Var)
	}
	p.nextToken()

	if p.peekTokenIs(token.ASSIGN) {
		args := p.parseNamedArgsList()
		return ast.NewFunctionCallNamedArgs(funcIdentifier.Var, args)
	}

	args := p.parseSimpleExpressionList(token.RPAREN)
	return ast.NewFunctionCall(funcIdentifier.Var, args...)
}

func (p *Parser) parseNamedArgsList() ast.NamedArgs {
	args := make(map[string]ast.Expression)
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		if p.curToken.Type != token.IDENT {
			p.tokTypeError(p.curToken, token.IDENT)
			return nil
		}
		varName := p.curToken.Val
		if !p.accept(token.ASSIGN) {
			p.tokTypeError(p.peekToken, token.ASSIGN)
			return nil
		}
		p.nextToken()
		value := p.parseSimpleExpression(LOWEST)
		args[varName] = value
		p.nextToken()
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}
	return args
}

func (p *Parser) noPrefixParseFnError(t token.Type) {
	p.errors = append(p.errors, ParsingError{
		Tok: p.curToken,
		Err: fmt.Errorf("no prefix parse function for token type %s", t),
	})
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// consumes the next token if it is of any of the allowed types
// returns true if any of the alllowed types were consumed, false otherwise
func (p *Parser) accept(allowed ...token.Type) bool {

	for _, t := range allowed {
		if p.peekTokenIs(t) {
			p.nextToken()
			return true
		}
	}

	return false
}

func (p *Parser) Errors() []error {
	return p.errors
}

func (p *Parser) HadError() bool {
	return len(p.errors) > 0
}

func (p *Parser) tokTypeError(got token.Token, expected ...token.Type) {
	switch got.Type {
	case token.ERROR:
		p.errors = append(p.errors, LexingError(got))
	default:
		err := ParsingError{
			Tok: got,
			Err: fmt.Errorf("expecting %q, got %q", expected, got),
		}
		if len(expected) == 1 {
			err.Err = fmt.Errorf("expecting %q, got %q", expected[0], got)
		}
		p.errors = append(p.errors, err)
	}
}
