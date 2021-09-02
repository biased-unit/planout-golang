package compiler

import (
	"bytes"
	"encoding/json"
	"github.com/biased-unit/planout-golang/compiler/parser"
)

type Compiler struct {
	p *parser.Parser
}

func New(p *parser.Parser) *Compiler {
	return &Compiler{
		p: p,
	}
}

type ParserErrors []error

func (pe ParserErrors) Error() string {
	var out bytes.Buffer
	for _, err := range pe {
		out.WriteString(err.Error())
		out.WriteString("\n")
	}
	return out.String()
}

func (c *Compiler) Run() ([]byte, error) {
	program := c.p.ParseProgram()

	if c.p.HadError() {
		return nil, ParserErrors(c.p.Errors())
	}

	if len(program.Seq) == 0 {
		return []byte(`{}`), nil
	}

	buffer := new(bytes.Buffer)
	enc := json.NewEncoder(buffer)
	enc.SetEscapeHTML(false) // so angle brackets are properly encoded (https://www.alexedwards.net/blog/json-surprises-and-gotchas#5)
	enc.SetIndent("", "  ")

	if err := enc.Encode(program); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
