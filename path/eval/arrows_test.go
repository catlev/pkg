package eval

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/catlev/pkg/model"
	"github.com/catlev/pkg/path/syntax"
	"github.com/catlev/pkg/store/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	m := &model.EntityModel{
		Types: []model.Type{
			{
				Name: "person",
				ID:   5,
				Attributes: model.Attributes{
					{
						Name: "age",
						Type: model.IntegerID,
					},
					{
						Name: "rank",
						Type: model.IntegerID,
					},
				},
			},
		},
	}
	c := NewCompiler(m, StandardOps)
	s, err := syntax.ParseString(`123|~(person/age)&rank`)
	require.Nil(t, err)
	p, err := c.Compile(s)
	require.Nil(t, err)
	assert.Equal(t, &unionPath{
		left: &intPath{valueID: model.IntegerID, value: 123},
		right: &intersectionPath{
			left: &joinPath{
				left: &attrFilter{
					entityID: 5,
					valueID:  model.IntegerID,
					column:   0,
				},
				right: &entityFilter{
					entityID: 5,
				},
			},
			right: &attrPath{
				entityID: 5,
				valueID:  model.IntegerID,
				column:   1,
			},
		},
	}, p)
}

type testStore struct {
	model *model.EntityModel
	arms  []storeArm
}

type storeArm struct {
	entityID block.Word
	fields   []block.Word
}

type testCursor struct {
	arm   storeArm
	width int
	pos   int
}

func (*testCursor) Err() error {
	return nil
}

func (c *testCursor) Next() bool {
	if c.pos+c.width >= len(c.arm.fields) {
		return false
	}
	c.pos += c.width
	return true
}

func (c *testCursor) This() Object {
	return Object{
		EntityID: c.arm.entityID,
		Fields:   c.arm.fields[c.pos : c.pos+c.width],
	}
}

func (s *testStore) FindEntities(entityID block.Word, key []block.Word) Cursor {
	for _, a := range s.arms {
		if a.entityID != entityID {
			continue
		}
		width := len(s.model.Types[a.entityID].Attributes)

		return &testCursor{
			arm:   a,
			width: width,
			pos:   -width,
		}
	}
	return &errCursor{nil}
}

func (*testStore) ParseValue(valueID block.Word, value string) (block.Word, error) {
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return block.Word(i), nil
}

func cursorEquals(c Cursor, xs []Object) bool {
	for _, x := range xs {
		if !c.Next() {
			return false
		}
		if !reflect.DeepEqual(c.This(), x) {
			return false
		}
	}
	return !c.Next()
}

func TestEval(t *testing.T) {
	m := &model.EntityModel{
		Types: []model.Type{
			model.IntegerID: {
				ID:   model.IntegerID,
				Kind: model.Value,
			},
			3: {
				ID:   3,
				Name: "person",
				Attributes: model.Attributes{
					{
						Name: "rank",
						Type: model.IntegerID,
					},
					{
						Name: "size",
						Type: model.IntegerID,
					},
				},
			},
		},
	}
	s := &testStore{
		model: m,
		arms: []storeArm{
			{
				entityID: 3,
				fields: []block.Word{
					1, 2,
					2, 10,
				},
			},
		},
	}
	h := NewHost(m, s)

	b := h.Eval(h.Absolute(), "1|(person&2/~rank)/size")

	assert.True(t, cursorEquals(b.Enumerate(), []Object{
		{EntityID: model.IntegerID, Fields: []block.Word{1}},
		{EntityID: model.IntegerID, Fields: []block.Word{10}},
	}))
}
