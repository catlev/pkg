package types

import (
	"github.com/catlev/pkg/path/syntax"
	"reflect"
	"testing"
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
			"\"a\"",
			[]Alternative{Alternative{absoluteType, attributeType}},
		},
		{
			"Term",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "b")},
			},
			"f",
			[]Alternative{alt("a", "b")},
		},
		{
			"Inverse",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "b")},
			},
			"~f",
			[]Alternative{alt("b", "a")},
		},
		{
			"Join",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "b")},
				"g": []Alternative{alt("b", "c")},
			},
			"f/g",
			[]Alternative{alt("a", "c")},
		},
		{
			"JoinTwoPaths",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "b"), alt("a", "c")},
				"g": []Alternative{alt("b", "d"), alt("c", "d")},
			},
			"f/g",
			[]Alternative{alt("a", "d")},
		},
		{
			"Union",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "b")},
				"g": []Alternative{alt("a", "c")},
			},
			"f|g",
			[]Alternative{alt("a", "b"), alt("a", "c")},
		},
		{
			"UnionDup",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "b")},
				"g": []Alternative{alt("a", "c")},
			},
			"f|g|f",
			[]Alternative{alt("a", "b"), alt("a", "c")},
		},
		{
			"Intersection",
			map[string][]Alternative{
				"f": []Alternative{alt("a", "c"), alt("a", "b")},
				"g": []Alternative{alt("a", "c")},
			},
			"f&g",
			[]Alternative{alt("a", "c")},
		},
		{
			"Combination",
			map[string][]Alternative{
				"id":   []Alternative{alt("a", "int"), alt("b", "int")},
				"b_id": []Alternative{alt("a", "int")},
				"a":    []Alternative{ent("a")},
				"b":    []Alternative{ent("b")},
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
			"id": []Alternative{alt("a", "int")},
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
