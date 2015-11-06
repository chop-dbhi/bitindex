package bitindex

import (
	"sort"
	"testing"
)

var (
	fruit = []uint32{
		1, // Apple
		2, // Cherry
		3, // Peach
		4, // Grape
		5, // Pear
		6, // Pineapple
		7, // Kiwi
		8, // Watermelon
		9, // Orange
	}

	people = []uint32{
		100, // Bob
		101, // Sue
		102, // Joe
	}

	pairs = map[uint32][]uint32{
		100: {1, 3},
		101: {4, 9},
		102: {4, 2, 3},
	}
)

func TestDomain(t *testing.T) {
	d := NewDomain(fruit)

	if d.Size() != len(fruit) {
		t.Errorf("expected size %d, got %d", len(fruit), d.Size())
	}

	for i, m := range fruit {
		b := uint32(i)

		if d.Bit(m) != b {
			t.Errorf("member %d should be bit %d", m, b)
		}

		if d.Member(b) != m {
			t.Errorf("bit %d should be member %d", b, m)
		}
	}

	// Add a 6th bit.
	if d.Add(10) != 9 {
		t.Errorf("error adding 9th bit")
	}
}

func TestArray(t *testing.T) {
	bits := []uint32{2, 17, 30, 54, 63}

	a := NewArray()

	for _, b := range bits {
		a.Set(b)

		if !a.Has(b) {
			t.Errorf("bit %d should be set", b)
		}

		a.Clear(b)

		if a.Has(b) {
			t.Errorf("bit %d should be cleared", b)
		}

		a.Flip(b)

		if !a.Has(b) {
			t.Errorf("bit %d should be set", b)
		}
	}

	// Each bit is on a separate byte.
	if a.Bytes() != len(bits) {
		t.Errorf("expected size %d, got %d", len(bits), a.Bytes())
	}
}

func TestIndex(t *testing.T) {
	ix := NewIndex(fruit)

	for k, s := range pairs {
		for _, b := range s {
			ix.Add(k, b)
		}
	}

	// The number of keys in the table.
	if ix.Table.Size() != 3 {
		t.Errorf("expected size 3, got %d", ix.Table.Size())
	}

	// The bytes required to encode the table.
	if ix.Table.Bytes() != 4 {
		t.Errorf("expected bytes 4, got %d", ix.Table.Bytes())
	}

	// Have of the bits are used.
	if ix.Sparsity() == 1 {
		t.Errorf("expected sparsity < 1, got %f", ix.Sparsity())
	}
}

func TestIndexOperations(t *testing.T) {
	ix := NewIndex(fruit)

	for k, s := range pairs {
		for _, b := range s {
			ix.Add(k, b)
		}
	}

	o, err := ix.Any(1, 2)

	if err != nil {
		t.Error(err)
	}

	a := Uint32Array(o)
	sort.Sort(a)

	if len(a) != 2 || a[0] != 100 || a[1] != 102 {
		t.Errorf("expected [100, 102], got %v", o)
	}

	o, err = ix.All(1, 3)

	if err != nil {
		t.Error(err)
	}

	if len(o) != 1 || o[0] != 100 {
		t.Errorf("expected [100], got %v", o)
	}

	o, err = ix.NotAny(3, 1)

	if err != nil {
		t.Error(err)
	}

	if len(o) != 1 || o[0] != 101 {
		t.Errorf("expected [101], got %v", o)
	}

	o, err = ix.NotAll(4, 2)

	if err != nil {
		t.Error(err)
	}

	a = Uint32Array(o)
	sort.Sort(a)

	if len(a) != 2 || a[0] != 100 || a[1] != 101 {
		t.Errorf("expected [100, 101], got %v", o)
	}
}

func BenchmarkDomainAdd(b *testing.B) {
	d := NewDomain(nil)

	for i := 0; i < b.N; i++ {
		d.Add(uint32(i))
	}
}

/*
var seed []uint32

func init() {
	seed = newSlice(1000000)
}

func newSlice(n uint32) []uint32 {
	s := make([]uint32, n)

	var i uint32

	for i = 0; i < n; i++ {
		s[i] = i
	}

	return s
}

func BenchmarkTestIndex(b *testing.B) {
	ix := NewIndex()

	for i := 0; i < b.N; i++ {

	}
}
*/
