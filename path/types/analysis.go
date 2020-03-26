package types

import (
	"errors"
	"github.com/catlev/pkg/path/syntax"
	"sort"
)

var ErrUnknown = errors.New("unknown path expression")

func Analyze(m Model, expr syntax.Tree) (Path, error) {
	switch expr.Kind {
	case syntax.Value:
		return analyzeValue(m, expr)
	case syntax.Term:
		return analyzeTerm(m, expr)
	case syntax.Inverse:
		return analyzeInverse(m, expr)
	case syntax.Join:
		return analyzeJoin(m, expr)
	case syntax.Intersection:
		return analyzeIntersection(m, expr)
	case syntax.Union:
		return analyzeUnion(m, expr)
	}
	panic("unreachable")
}

func (p *Path) normalize() {
	sort.Slice(p.Alternatives, func(i, j int) bool {
		x, y := p.Alternatives[i], p.Alternatives[j]
		return x.Source.lte(y.Source) && x.Target.lte(y.Target)
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

func (t Type) lte(u Type) bool {
	if t.Kind != u.Kind {
		return t.Kind < u.Kind
	}
	if t.Name != u.Name {
		return t.Name < u.Name
	}
	return true
}

func analyzeValue(m Model, expr syntax.Tree) (Path, error) {
	return Path{expr, []Alternative{{absoluteType, attributeType}}}, nil
}

func analyzeTerm(m Model, expr syntax.Tree) (Path, error) {
	return m.Lookup(expr.Value)
}

func analyzeInverse(m Model, expr syntax.Tree) (Path, error) {
	path, err := Analyze(m, expr.Children[0])
	if err != nil {
		return Path{}, err
	}
	alts := make([]Alternative, len(path.Alternatives))
	for i, a := range path.Alternatives {
		alts[i] = Alternative{
			Source: a.Target,
			Target: a.Source,
		}
	}
	path.Alternatives = alts
	path.normalize()
	return path, nil
}

func analyzeJoin(m Model, expr syntax.Tree) (Path, error) {
	return analyzeComposition(m, expr, func(left, right []Alternative) []Alternative {
		var alts []Alternative
		for _, a := range left {
			idx := sort.Search(len(right), func(idx int) bool {
				return a.Target.lte(right[idx].Source)
			})
			for ; idx < len(right) && right[idx].Source == a.Target; idx++ {
				alts = append(alts, Alternative{
					Source: a.Source,
					Target: right[idx].Target,
				})
			}
		}
		return alts
	})
}

func analyzeIntersection(m Model, expr syntax.Tree) (Path, error) {
	return analyzeComposition(m, expr, func(left, right []Alternative) []Alternative {
		var alts []Alternative
		for _, a := range left {
			idx := sort.Search(len(right), func(idx int) bool {
				b := right[idx]
				return a.Source.lte(b.Source)
			})
			for ; idx < len(right) && right[idx].Source == a.Source; idx++ {
				if right[idx] == a {
					alts = append(alts, a)
					break
				}
			}
		}
		return alts
	})
}

func analyzeUnion(m Model, expr syntax.Tree) (Path, error) {
	return analyzeComposition(m, expr, func(left, right []Alternative) []Alternative {
		var alts []Alternative
		alts = append(alts, left...)
		alts = append(alts, right...)
		return alts
	})
}

func analyzeComposition(m Model, expr syntax.Tree, f func(left, right []Alternative) []Alternative) (Path, error) {
	left, err := Analyze(m, expr.Children[0])
	if err != nil {
		return Path{}, err
	}
	right, err := Analyze(m, expr.Children[1])
	if err != nil {
		return Path{}, err
	}
	alts := f(left.Alternatives, right.Alternatives)
	p := Path{expr, alts}
	p.normalize()
	return p, nil
}
