package syntax

import (
	"errors"
	"strings"
	"testing"

	"github.com/catlev/pkg/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpr(t *testing.T) {
	for _, test := range []struct {
		name string
		expr string
		tree path.Expr
	}{
		{
			name: "String",
			expr: `"a"`,
			tree: path.Expr{Kind: path.String, Value: `"a"`},
		},
		{
			name: "Integer",
			expr: "12",
			tree: path.Expr{Kind: path.Integer, Value: "12"},
		},
		{
			name: "NamedElement",
			expr: "a",
			tree: path.Expr{Kind: path.Rel, Value: "a"},
		},
		{
			name: "Namedpath.Op",
			expr: "a()",
			tree: path.Expr{Kind: path.Op, Value: "a"},
		},
		{
			name: "Namedpath.OpArgs",
			expr: "a(1,2)",
			tree: path.Expr{Kind: path.Op, Value: "a", Children: []path.Expr{
				{Kind: path.Integer, Value: "1"},
				{Kind: path.Integer, Value: "2"},
			}},
		},
		{
			name: "Inv",
			expr: "~a",
			tree: path.Expr{Kind: path.Op, Value: "inverse", Children: []path.Expr{
				{Kind: path.Rel, Value: "a"},
			}},
		},
		{
			name: "Union",
			expr: "a|b",
			tree: path.Expr{Kind: path.Op, Value: "union", Children: []path.Expr{
				{Kind: path.Rel, Value: "a"},
				{Kind: path.Rel, Value: "b"},
			}},
		},
		{
			name: "Intersection",
			expr: "a&b",
			tree: path.Expr{Kind: path.Op, Value: "intersection", Children: []path.Expr{
				{Kind: path.Rel, Value: "a"},
				{Kind: path.Rel, Value: "b"},
			}},
		},
		{
			name: "Join",
			expr: "a/b",
			tree: path.Expr{Kind: path.Op, Value: "join", Children: []path.Expr{
				{Kind: path.Rel, Value: "a"},
				{Kind: path.Rel, Value: "b"},
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
