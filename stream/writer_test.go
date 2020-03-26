package stream

import (
	"errors"
	"strings"
	"testing"
)

func TestWriteField(t *testing.T) {
	var buf strings.Builder
	w := NewWriter(&buf)
	w.StringField("string", "hello")
	w.IntField("int", 1)
	w.BoolField("bool", true)
	w.err = errors.New("simulated failure")
	w.StringField("string", "hello again")
	expected := `string:"hello"int:"1"bool:"true"`
	if buf.String() != expected {
		t.Errorf("got %q, expecting %q", buf.String(), expected)
	}
}
