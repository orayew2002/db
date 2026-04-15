package db

import (
	lb "github.com/orayew2002/db/src/proto"
	"google.golang.org/protobuf/proto"
)

type CreateTable struct {
	Vls []string
}

func (c CreateTable) Raw() []byte {
	s := lb.CreateTable{Vals: c.Vals()}

	b, err := proto.Marshal(&s)
	if err != nil {
		panic(err.Error())
	}

	return b
}

func (c CreateTable) Vals() []string {
	return c.Vls
}

type Delete struct {
	Col string
	Val any
}

func (d Delete) Raw() []byte {
	return nil
}

func (d Delete) Vals() []string {
	return nil
}

type Insert struct {
	Val map[string]any
}

func (i Insert) Raw() []byte {
	return nil
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
	return nil
}

func (u Update) Vals() []string {
	return nil
}
