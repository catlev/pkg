package table

import "github.com/catlev/pkg/store/block"

type Schema struct {
	ID      block.Word
	Name    string
	Columns []Column
}

type Column struct {
	Name           string
	Offset, KeyPos int
}

func (s *Schema) RowsPerBlock() int {
	return 64 / len(s.Columns)
}
