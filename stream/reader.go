package stream

import (
	"errors"
	"io"
	"strconv"
	"text/scanner"
)

// Reader provides services to objects that parse streams.
type Reader struct {
	src   scanner.Scanner
	depth int
	tok   rune

	name string
	kind Kind
	err  error
}

// Kind represents the kind of the section.
type Kind byte

const (
	// Field kind
	Field Kind = iota + 1

	// Record kind
	Record
)

// NewReader creates a new reader from an io.Reader.
func NewReader(src io.Reader) *Reader {
	return &Reader{src: *new(scanner.Scanner).Init(src)}
}

// Next moves to the next named section (field or record) of the input. It returns whether progress
// can be made by a consumer.
//
// If Next is called twice without a call to either a Field() method or the Record() method in
// between then an error is signalled and the second Next() returns false.
func (p *Reader) Next() bool {
	if p == nil {
		return false
	}
	if p.kind != 0 {
		p.setErr(errors.New("unexpected identifier " + p.name))
		return false
	}
	if !p.nextTok() {
		return false
	}
	if p.depth > 0 && p.tok == '}' {
		p.depth--
		return false
	}
	if p.tok != scanner.Ident {
		p.setErr(errors.New("expecting ident"))
		return false
	}
	p.name = p.src.TokenText()
	if !p.nextTok() {
		return false
	}
	switch p.tok {
	case ':':
		p.kind = Field
	case '{':
		p.kind = Record
	default:
		p.setErr(errors.New("expecting field or record"))
		return false
	}
	return true
}

// Name gives the name of the current section under consideration. Note that if Kind() does not
// return Field or Value then this value is undefined.
func (p *Reader) Name() string {
	return p.name
}

// Kind gives the kind of the section under consideration, either Field or Record. If it returns
// anything other than one of these values, the reader is in a bad state and it would be unwise to
// continue parsing.
func (p *Reader) Kind() Kind {
	return p.kind
}

// Err gives the error that has been set, if any.
func (p *Reader) Err() error {
	return p.err
}

// Record uses the provided parser to parse a record. An error will be signalled if the current
// section is not a record.
func (p *Reader) Record() *Reader {
	if p.kind != Record {
		p.setErr(errors.New("expecting record"))
		return nil
	}
	p.kind = 0
	p.depth++
	return p
}

// StringField reads a field and interprets it as a string. If the current
// section is not a field then an error is signalled.
func (p *Reader) StringField() string {
	attr := p.parseAttr()
	if attr == "" {
		return ""
	}
	res, _ := strconv.Unquote(attr)
	return res
}

// IntField reads a field and interprets it as an integer. If the current
// section is not a field containing an integer then an error is signalled.
func (p *Reader) IntField() int {
	attr := p.parseAttr()
	if attr == "" {
		return 0
	}
	res, err := strconv.Atoi(attr[1 : len(attr)-1])
	if err != nil {
		p.setErr(err)
	}
	return res
}

// BoolField reads a field and interprets it as a boolean. If the current
// section is not a field containing a boolean then an error is signalled.
func (p *Reader) BoolField() bool {
	attr := p.parseAttr()
	if attr == "" {
		return false
	}
	res, err := strconv.ParseBool(attr[1 : len(attr)-1])
	if err != nil {
		p.setErr(err)
	}
	return res
}

// ExpectEOF asserts that the entire file was parsed.
func (p *Reader) ExpectEOF() {
	if p.tok != scanner.EOF {
		p.setErr(errors.New("expecting EOF"))
	}
}

func (p *Reader) nextTok() bool {
	if p.err != nil || p.tok == scanner.EOF {
		return false
	}
	p.tok = p.src.Scan()
	return p.tok != scanner.EOF
}

func (p *Reader) parseAttr() string {
	if p.kind != Field {
		p.setErr(errors.New("expecting field"))
	}
	p.kind = 0
	if !p.nextTok() {
		return ""
	}
	if p.tok != scanner.String {
		p.setErr(errors.New("expecting field value"))
		return ""
	}
	return p.src.TokenText()
}

func (p *Reader) setErr(err error) {
	if p.err != nil {
		return
	}
	p.err = err
}
