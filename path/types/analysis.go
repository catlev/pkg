package types

import (
	"errors"

	"github.com/catlev/pkg/path"
)

var ErrUnknown = errors.New("unknown path expression")

func Analyze(m Model, expr path.Expr) (Path, error) {
	switch expr.Kind {
	case path.Integer:
		return analyzeValue(m, expr)
	case path.Rel:
		return analyzeTerm(m, expr)
	case path.Op:
		return analyzeOp(m, expr)
	}
	panic("unreachable")
}

func analyzeValue(m Model, expr path.Expr) (Path, error) {
	return Path{expr, []Alternative{{absoluteType, attributeType}}}, nil
}

func analyzeTerm(m Model, expr path.Expr) (Path, error) {
	return m.Lookup(expr.Value)
}

func analyzeOp(m Model, expr path.Expr) (Path, error) {
	switch expr.Value {
	case "inverse":
		return analyzeInverse(m, expr)
	case "union":
		return analyzeUnion(m, expr)
	case "intersection":
		return analyzeIntersection(m, expr)
	case "join":
		return analyzeJoin(m, expr)
	}
	return Path{}, ErrUnknown
}

func analyzeInverse(m Model, expr path.Expr) (Path, error) {
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

func analyzeJoin(m Model, expr path.Expr) (Path, error) {
	return analyzeComposition(m, expr, func(left, right Path) []Alternative {
		var alts []Alternative
		for _, a := range left.Alternatives {
			for _, b := range right.filterAlternatives(a.Target) {
				alts = append(alts, Alternative{
					Source: a.Source,
					Target: b.Target,
				})
			}
		}
		return alts
	})
}

func analyzeIntersection(m Model, expr path.Expr) (Path, error) {
	return analyzeComposition(m, expr, func(left, right Path) []Alternative {
		var alts []Alternative
		for _, a := range left.Alternatives {
			for _, b := range right.filterAlternatives(a.Source) {
				if a.Target != b.Target {
					continue
				}
				alts = append(alts, a)
			}
		}
		return alts
	})
}

func analyzeUnion(m Model, expr path.Expr) (Path, error) {
	return analyzeComposition(m, expr, func(left, right Path) []Alternative {
		var alts []Alternative
		alts = append(alts, left.Alternatives...)
		alts = append(alts, right.Alternatives...)
		return alts
	})
}

func analyzeComposition(m Model, expr path.Expr, f func(left, right Path) []Alternative) (Path, error) {
	left, err := Analyze(m, expr.Children[0])
	if err != nil {
		return Path{}, err
	}
	right, err := Analyze(m, expr.Children[1])
	if err != nil {
		return Path{}, err
	}
	alts := f(left, right)
	p := Path{expr, alts}
	p.normalize()
	return p, nil
}
