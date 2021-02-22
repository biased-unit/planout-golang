package planout

import (
	"github.com/stitchfix/planout-golang/compiler"
	"github.com/stitchfix/planout-golang/compiler/lexer"
	"github.com/stitchfix/planout-golang/compiler/parser"
)

func Compile(code string) ([]byte, error) {
	lx := lexer.New(code)
	ps := parser.New(lx)
	comp := compiler.New(ps)

	return comp.Run()
}
