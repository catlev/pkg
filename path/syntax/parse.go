package syntax

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"text/scanner"
	"unsafe"
)

func Must(t Tree, err error) Tree {
	if err != nil {
		panic(err)
	}
	return t
}

func ParseString(s string) (Tree, error) {
	return Parse(strings.NewReader(s))
}

func Parse(r io.Reader) (Tree, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return Tree{}, err
	}
	s := new(scanner.Scanner).Init(bytes.NewReader(buf))
	p := parser{buf: buf, s: *s}
	p.s.Error = func(s *scanner.Scanner, err string) {
		p.err = fmt.Errorf("%w: %s", ErrFailedToMatch, err)
	}
	res := p.parse(0)
	return res, p.err
}

// Information required to parse a path expression, including scanning the
// source and producing the results.
type parser struct {
	buf    []byte
	s      scanner.Scanner
	tok    rune
	cached bool
	data   Tree
	err    error
}

// Set the parser's error if not already set.
func (p *parser) setErr(err error) {
	if p.err == nil {
		p.err = err
	}
}

// Parse an expression at a particular precedence level.
func (p *parser) parse(prec int) Tree {
	left := p.prefix()
	for {
		if p.err != nil || p.tok == scanner.EOF {
			break
		}
		right := p.infix(prec, left)
		if right.Kind == Invalid {
			break
		}
		left = right
	}
	return left
}

// Parse an expression in prefix position.
func (p *parser) prefix() Tree {
	switch p.next() {
	case scanner.String:
		start, end := p.pos()
		return p.leaf(Value, start, end)
	case scanner.Ident, '*':
		start, end := p.pos()
		return p.leaf(Term, start, end)
	case '~':
		return p.branch(Inverse, p.parse(100))
	case '(':
		path := p.parse(0)
		if p.next() != ')' {
			p.setErr(fmt.Errorf("%w: expecting ')'", ErrFailedToMatch))
		}
		return path
	}
	p.setErr(ErrFailedToMatch)
	return Tree{}
}

// Infix parsing productions.
var infixes = [256]struct {
	kind         Kind
	outer, inner int
}{
	'/': {Join, 80, 80},
	'&': {Intersection, 70, 70},
	'|': {Union, 70, 70},
}

// Parse an expression in infix position.
func (p *parser) infix(prec int, left Tree) Tree {
	var kind Kind
	var outer, inner int
	c := p.peek()
	if c >= 0 && c < 256 {
		inf := infixes[c]
		kind = inf.kind
		outer = inf.outer
		inner = inf.inner
	}
	if p.err != nil || kind == Value || prec >= outer {
		return Tree{}
	}
	p.next()
	return p.branch(kind, left, p.parse(inner))
}

// Look at the next token without advancing position.
func (p *parser) peek() rune {
	p.tok = p.next()
	p.cached = true
	return p.tok
}

// Look at the next token.
func (p *parser) next() rune {
	if p.err != nil {
		return 0
	}
	if p.cached {
		p.cached = false
		return p.tok
	}
	p.tok = p.s.Scan()
	return p.tok
}

// Get the (start,end) of the current token.
func (p *parser) pos() (int, int) {
	off := p.s.Pos().Offset
	return off - len(p.s.TokenText()), off
}

func (p *parser) leaf(kind Kind, start, end int) Tree {
	return Tree{Kind: kind, Value: p.value(start, end)}
}

func (p *parser) branch(kind Kind, children ...Tree) Tree {
	return Tree{Kind: kind, Children: children}
}

func (p *parser) value(start, end int) string {
	return (*((*string)(unsafe.Pointer(&p.buf))))[start:end]
}
