package types

import (
	"github.com/catlev/pkg/path/syntax"
	"sort"
)

type Path struct {
	Expr         syntax.Tree
	Alternatives []Alternative
}

type Alternative struct {
	Source, Target Type
}

type Kind int

const (
	Absolute Kind = iota
	Attribute
	Entity
)

type Type struct {
	Kind Kind
	Name string
}

type Model interface {
	Lookup(name string) (Path, error)
}

var (
	absoluteType  = Type{Kind: Absolute, Name: "^"}
	attributeType = Type{Kind: Attribute, Name: "$"}
)

func (p *Path) normalize() {
	sort.Slice(p.Alternatives, func(i, j int) bool {
		x, y := p.Alternatives[i], p.Alternatives[j]
		if x.Source != y.Source {
			return x.Source.lessThan(y.Source)
		}
		return x.Target.lessThan(y.Target)
	})
	for i := 1; i < len(p.Alternatives); i++ {
		if p.Alternatives[i] == p.Alternatives[i-1] {
			p.dedup()
			break
		}
	}
}

func (p *Path) dedup() {
	deduped := []Alternative{p.Alternatives[0]}
	for i, a := range p.Alternatives[1:] {
		if p.Alternatives[i] != a {
			deduped = append(deduped, a)
		}
	}
	p.Alternatives = deduped
}

func (p Path) filterAlternatives(source Type) []Alternative {
	start := sort.Search(len(p.Alternatives), func(idx int) bool {
		candidate := p.Alternatives[idx].Source
		if candidate == source {
			return true
		}
		return source.lessThan(candidate)
	})
	end := sort.Search(len(p.Alternatives), func(idx int) bool {
		candidate := p.Alternatives[idx].Source
		return source.lessThan(candidate)
	})
	return p.Alternatives[start:end]
}

func (t Type) lessThan(u Type) bool {
	if t.Kind != u.Kind {
		return t.Kind < u.Kind
	}
	if t.Name != u.Name {
		return t.Name < u.Name
	}
	return false

}
