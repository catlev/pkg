package syntax

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func leaf(kind Kind, value string) Tree {
	return Tree{Kind: kind, Value: value}
}

func branch(kind Kind, children ...Tree) Tree {
	return Tree{Kind: kind, Children: children}
}

func TestExpr(t *testing.T) {
	e := `a/(~name|boom)&"hello"/123`
	n, err := Parse(strings.NewReader(e))
	if err != nil {
		t.Fatalf("%s", err)
	}
	expected := branch(Intersection,
		branch(Join,
			leaf(Term, "a"),
			branch(Union,
				branch(Inverse, leaf(Term, "name")),
				leaf(Term, "boom"),
			),
		),
		branch(Join,
			leaf(String, `"hello"`),
			leaf(Integer, "123"),
		),
	)
	if !reflect.DeepEqual(n, expected) {
		t.Errorf("%v != %v", n, expected)
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
