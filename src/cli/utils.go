package cli

import (
	"github.com/orayew2002/db/src/lexer"
	"github.com/orayew2002/db/src/parser"
)

type Action string

func parseCMD(cmd string) parser.Statement {
	s, err := (parser.New(lexer.New(cmd))).Parse()
	if err != nil {
		panic(err.Error())
	}

	return s
}
