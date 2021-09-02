package planout

import (
	"encoding/json"

	"github.com/biased-unit/planout-golang/compiler"
	"github.com/biased-unit/planout-golang/compiler/lexer"
	"github.com/biased-unit/planout-golang/compiler/parser"
)

func Compile(script string) (map[string]interface{}, error) {
	lx := lexer.New(script)
	ps := parser.New(lx)
	comp := compiler.New(ps)

	marshalled, err := comp.Run()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(marshalled, &result); err != nil {
		return nil, err
	}

	return result, nil
}
