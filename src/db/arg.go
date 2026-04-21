package db

import (
	lp "github.com/orayew2002/db/src/proto"
	"github.com/orayew2002/db/src/shared"
	"google.golang.org/protobuf/proto"
)

type Arg interface {
	AppendRaw([]byte) []byte
	Vals() []string
}

type CreateTable struct {
	Cols []ColDef
}

type ColDef struct {
	Name string
	Type string
}

func (c CreateTable) AppendRaw(dst []byte) []byte {
	tc := make([]*lp.TableCol, len(c.Cols))
	for i, col := range c.Cols {
		tc[i] = &lp.TableCol{Name: col.Name, Type: col.Type}
	}
	b, _ := proto.Marshal(&lp.CreateTable{Cols: tc})
	return append(dst, b...)
}

func (c CreateTable) Vals() []string {
	v := make([]string, len(c.Cols))
	for i, val := range c.Cols {
		v[i] = val.Name
	}
	return v
}

type Delete struct {
	Col string
	Val any
}

func (d Delete) AppendRaw(dst []byte) []byte {
	v, _ := shared.Marshal(d.Val)
	b, _ := proto.Marshal(&lp.Delete{Col: d.Col, Val: v})
	return append(dst, b...)
}

func (d Delete) Vals() []string {
	return nil
}

type Insert struct {
	Val map[string]any
}

func (i Insert) AppendRaw(dst []byte) []byte {
	return appendMapJSON(dst, i.Val)
}

func (i Insert) Vals() []string {
	return nil
}

type Update struct {
	Col  string
	Val  any
	Args map[string]any
}

func (u Update) AppendRaw(dst []byte) []byte {
	v, _ := shared.Marshal(u.Val)
	vs, _ := shared.MarshalMap(u.Args)
	b, _ := proto.Marshal(&lp.Update{Col: u.Col, Val: v, Args: vs})
	return append(dst, b...)
}

func (u Update) Vals() []string {
	return nil
}
