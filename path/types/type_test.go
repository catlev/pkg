package types

import (
	"reflect"
	"testing"
)

func alt(from, to string) Alternative {
	return Alternative{
		Type{Kind: Entity, Name: from},
		Type{Kind: Entity, Name: to},
	}
}

func ent(typ string) Alternative {
	return Alternative{
		absoluteType,
		Type{Kind: Entity, Name: typ},
	}
}

func TestNormalization(t *testing.T) {
	for _, test := range []struct {
		name   string
		alts   []Alternative
		normal []Alternative
	}{
		{
			name: "AlreadyNormal",
			alts: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
			},
			normal: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
			},
		},
		{
			name: "KindOutOfOrder",
			alts: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
				ent("a"),
			},
			normal: []Alternative{
				ent("a"),
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
			},
		},
		{
			name: "SourceOutOfOrder",
			alts: []Alternative{
				alt("a", "a"),
				alt("b", "a"),
				alt("a", "b"),
			},
			normal: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
			},
		},
		{
			name: "TargetOutOfOrder",
			alts: []Alternative{
				alt("a", "b"),
				alt("a", "a"),
				alt("b", "a"),
			},
			normal: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
			},
		},
		{
			name: "Duplicated",
			alts: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
				alt("a", "b"),
			},
			normal: []Alternative{
				alt("a", "a"),
				alt("a", "b"),
				alt("b", "a"),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := &Path{Alternatives: test.alts}
			p.normalize()
			if !reflect.DeepEqual(p.Alternatives, test.normal) {
				t.Error(p.Alternatives)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	p := Path{Alternatives: []Alternative{
		alt("a", "a"),
		alt("a", "b"),
		alt("a", "c"),
		alt("b", "a"),
	}}
	sub := p.filterAlternatives(Type{Kind: Entity, Name: "a"})
	expect := []Alternative{
		alt("a", "a"),
		alt("a", "b"),
		alt("a", "c"),
	}
	if !reflect.DeepEqual(sub, expect) {
		t.Error(sub)
	}
}
