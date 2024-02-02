package block

import "testing"

func TestByteConversion(t *testing.T) {
	bl := Block{2: 257}
	bs := bl.Bytes()
	for i, x := range bs {
		switch i {
		case 16:
			if x != 1 {
				t.Fail()
			}
		case 17:
			if x != 1 {
				t.Fail()
			}
		default:
			if x != 0 {
				t.Fail()
			}
		}
	}
	t.Log(bs)
}
