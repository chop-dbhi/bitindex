package bitindex

import (
	"sort"
	"testing"
)

var (
	fruit = []uint32{
		1, // Apples
		2, // Cherries
		3, // Peaches
		4, // Grapes
	}

	people = []uint32{
		100, // Bob
		101, // Sue
		102, // Joe
	}

	pairs = [][2]uint32{
		{100, 1},
		{100, 3},
		{101, 4},
		{102, 4},
		{102, 2},
		{102, 3},
	}
)

func TestDomain(t *testing.T) {
	ms := []uint32{1, 3, 4, 10, 12, 49}

	d := NewDomain(ms)

	if d.Size() != len(ms) {
		t.Errorf("expected size %d, got %d", len(ms), d.Size())
	}

	for i, m := range ms {
		b := uint32(i)

		if d.Bit(m) != b {
			t.Errorf("member %d should be bit %d", m, b)
		}

		if d.Member(b) != m {
			t.Errorf("bit %d should be member %d", b, m)
		}
	}

	// Add a 6th bit.
	if d.Add(101) != 6 {
		t.Errorf("error adding 6th bit")
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
	ix := NewIndex(NewDomain(fruit))

	for _, p := range pairs {
		ix.Add(p[0], p[1])
	}

	if ix.Table.Size() != 3 {
		t.Errorf("expected size 3, got %d", ix.Table.Size())
	}

	// Domain fits in one byte with three keys.
	if ix.Table.Bytes() != 3 {
		t.Errorf("expected bytes 3, got %d", ix.Table.Bytes())
	}

	// All required bytes are also allocated.
	if ix.Sparsity() != 0 {
		t.Errorf("expected sparsity 0, got %f", ix.Sparsity())
	}

}

type uint32Array []uint32

func (a uint32Array) Len() int {
	return len(a)
}

func (a uint32Array) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a uint32Array) Less(i, j int) bool {
	return a[i] < a[j]
}

func TestIndexOperations(t *testing.T) {
	ix := NewIndex(NewDomain(fruit))

	for _, p := range pairs {
		ix.Add(p[0], p[1])
	}

	o, err := ix.Any(1, 2)

	if err != nil {
		t.Error(err)
	}

	a := uint32Array(o)
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

	a = uint32Array(o)
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
