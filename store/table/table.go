package table

import (
	"github.com/catlev/pkg/store/block"
)

type Table struct {
	schema *Schema
	store  block.Reader
	start  block.Word
	depth  int
}

type Query []QueryClause

type QueryClause struct {
	ColumnOffset int
	Comparison   QueryComparison
	Value        block.Word
}

type QueryComparison int

const (
	EQ QueryComparison = iota
	GT
	LT
)

type QueryResult struct {}

func NewTable(schema *Schema, store block.Reader, start block.Word, depth int) *Table {
	return &Table{schema: schema, store: store, depth: depth, start: start}
}

func (r QueryResult) Next() bool {
	return false
}

func (r QueryResult) Get(i int) block.Word {
	return 0
}

func (t *Table) Query(q Query) (*QueryResult, error) {
	return nil, nil
}
