package syntax

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpr(t *testing.T) {
	for _, test := range []struct {
		name string
		expr string
		tree Tree
	}{
		{
			name: "String",
			expr: `"a"`,
			tree: Tree{Kind: String, Value: `"a"`},
		},
		{
			name: "Integer",
			expr: "12",
			tree: Tree{Kind: Integer, Value: "12"},
		},
		{
			name: "NamedElement",
			expr: "a",
			tree: Tree{Kind: Rel, Value: "a"},
		},
		{
			name: "NamedOp",
			expr: "a()",
			tree: Tree{Kind: Op, Value: "a"},
		},
		{
			name: "NamedOpArgs",
			expr: "a(1,2)",
			tree: Tree{Kind: Op, Value: "a", Children: []Tree{
				{Kind: Integer, Value: "1"},
				{Kind: Integer, Value: "2"},
			}},
		},
		{
			name: "Inv",
			expr: "~a",
			tree: Tree{Kind: Op, Value: "inverse", Children: []Tree{
				{Kind: Rel, Value: "a"},
			}},
		},
		{
			name: "Union",
			expr: "a|b",
			tree: Tree{Kind: Op, Value: "union", Children: []Tree{
				{Kind: Rel, Value: "a"},
				{Kind: Rel, Value: "b"},
			}},
		},
		{
			name: "Intersection",
			expr: "a&b",
			tree: Tree{Kind: Op, Value: "intersection", Children: []Tree{
				{Kind: Rel, Value: "a"},
				{Kind: Rel, Value: "b"},
			}},
		},
		{
			name: "Join",
			expr: "a/b",
			tree: Tree{Kind: Op, Value: "join", Children: []Tree{
				{Kind: Rel, Value: "a"},
				{Kind: Rel, Value: "b"},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := Parse(strings.NewReader(test.expr))
			require.Nil(t, err)
			assert.Equal(t, test.tree, parsed)
		})
	}
}

func TestFailure(t *testing.T) {
	_, err := Parse(strings.NewReader(`(`))
	if !errors.Is(err, ErrFailedToMatch) {
		t.Error(err)
	}
	_, err = Parse(strings.NewReader("\x00"))
	if !errors.Is(err, ErrFailedToMatch) {
		t.Error(err)
	}
}
