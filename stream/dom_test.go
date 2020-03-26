package stream

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestDOM(t *testing.T) {
	var d DOM
	err := d.Unmarshal(strings.NewReader(`
	
	root {
		child {
			name: "Test"
		}
		child {
			name: "Test2"
		}
	}
	
	`))

	if err != nil {
		t.Fatal(err)
	}
	if d.Children[0].Children[0].Name != "child" {
		t.Error(d)
	}
	if d.Children[0].Children[1].Fields["name"] != "Test2" {
		t.Error(d)
	}

	r, err := d.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != `root{child{name:"Test"}child{name:"Test2"}}` {
		t.Error(buf)
	}
}
