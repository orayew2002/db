package db

import (
	lp "github.com/orayew2002/db/src/proto"
	"github.com/orayew2002/db/src/shared"
	"google.golang.org/protobuf/proto"
)

type CreateTable struct {
	Cols []ColDef
}

type ColDef struct {
	Name string
	Type string
}

func (c CreateTable) Raw() []byte {
	tc := make([]*lp.TableCol, len(c.Cols))
	for i, c := range c.Cols {
		tc[i] = &lp.TableCol{Name: c.Name, Type: c.Type}
	}

	b, err := proto.Marshal(&lp.CreateTable{Cols: tc})
	if err != nil {
		panic(err.Error())
	}

	return b
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

func (d Delete) Raw() []byte {
	v, _ := shared.Marshal(d.Val)
	raw, err := proto.Marshal(&lp.Delete{Col: d.Col, Val: v})
	if err != nil {
		panic(err.Error())
	}

	return raw
}

func (d Delete) Vals() []string {
	return nil
}

type Insert struct {
	Val map[string]any
}

func (i Insert) Raw() []byte {
	converted, err := shared.MarshalMap(i.Val)
	if err != nil {
		panic(err.Error())
	}

	raw, err := proto.Marshal(&lp.Insert{Val: converted})
	if err != nil {
		panic(err.Error())
	}

	return raw
}

func (i Insert) Vals() []string {
	return nil
}

type Update struct {
	Col  string
	Val  any
	Args map[string]any
}

func (u Update) Raw() []byte {
	v, err := shared.Marshal(u.Val)
	if err != nil {
		panic(err.Error())
	}

	vs, err := shared.MarshalMap(u.Args)
	if err != nil {
		panic(err.Error())
	}

	raw, err := proto.Marshal(&lp.Update{
		Col:  u.Col,
		Val:  v,
		Args: vs,
	})
	if err != nil {
		panic(err.Error())
	}

	return raw
}

func (u Update) Vals() []string {
	return nil
}
