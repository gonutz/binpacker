package binpacker

import "testing"

func TestEnlarge(t *testing.T) {
	p := New(5, 5)
	p.Enlarge(20, 20)
	_, err := p.Insert(15, 15)
	if err != nil {
		t.Fatal(err)
	}
}
