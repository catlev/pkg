package types

import (
	"reflect"
	"testing"

	"github.com/catlev/pkg/path/syntax"
)

type testModel struct {
	items map[string][]Alternative
}

func (m *testModel) Lookup(name string) (Path, error) {
	alts, ok := m.items[name]
	if !ok {
		return Path{}, ErrUnknown
	}
	return Path{
		syntax.Must(syntax.ParseString(name)),
		alts,
	}, nil
}

func TestAnalysis(t *testing.T) {
	for _, test := range []struct {
		name     string
		setting  map[string][]Alternative
		path     string
		resolved []Alternative
	}{
		{
			"Value",
			nil,
			`"a"`,
			[]Alternative{{absoluteType, attributeType}},
		},
		{
			"Term",
			map[string][]Alternative{
				"f": {alt("a", "b")},
			},
			"f",
			[]Alternative{alt("a", "b")},
		},
		{
			"Inverse",
			map[string][]Alternative{
				"f": {alt("a", "b")},
			},
			"~f",
			[]Alternative{alt("b", "a")},
		},
		{
			"Join",
			map[string][]Alternative{
				"f": {alt("a", "b")},
				"g": {alt("b", "c")},
			},
			"f/g",
			[]Alternative{alt("a", "c")},
		},
		{
			"JoinTwoPaths",
			map[string][]Alternative{
				"f": {alt("a", "b"), alt("a", "c")},
				"g": {alt("b", "d"), alt("c", "d")},
			},
			"f/g",
			[]Alternative{alt("a", "d")},
		},
		{
			"Union",
			map[string][]Alternative{
				"f": {alt("a", "b")},
				"g": {alt("a", "c")},
			},
			"f|g",
			[]Alternative{alt("a", "b"), alt("a", "c")},
		},
		{
			"UnionDup",
			map[string][]Alternative{
				"f": {alt("a", "b")},
				"g": {alt("a", "c")},
			},
			"f|g|f",
			[]Alternative{alt("a", "b"), alt("a", "c")},
		},
		{
			"Intersection",
			map[string][]Alternative{
				"f": {alt("a", "c"), alt("a", "b")},
				"g": {alt("a", "c")},
			},
			"f&g",
			[]Alternative{alt("a", "c")},
		},
		{
			"Combination",
			map[string][]Alternative{
				"id":   {alt("a", "int"), alt("b", "int")},
				"b_id": {alt("a", "int")},
				"a":    {ent("a")},
				"b":    {ent("b")},
			},
			"(~a/b)&(b_id/~id)",
			[]Alternative{alt("a", "b")},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			m := &testModel{test.setting}
			p, err := Analyze(m, syntax.Must(syntax.ParseString(test.path)))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(p.Alternatives, test.resolved) {
				t.Error(p.Alternatives)
			}
		})
	}
}

func TestError(t *testing.T) {
	m := &testModel{
		map[string][]Alternative{
			"id": {alt("a", "int")},
		},
	}
	for _, test := range []struct {
		name string
		path string
	}{
		{
			"Direct",
			"f",
		},
		{
			"Inverse",
			"~f",
		},
		{
			"Join",
			"f/g",
		},
		{
			"JoinRight",
			"id/f",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := Analyze(m, syntax.Must(syntax.ParseString(test.path)))
			if err == nil {
				t.Error("expecting error")
			}
		})
	}
}
