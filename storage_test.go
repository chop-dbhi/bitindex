package bitindex

import (
	"bytes"
	"testing"
)

func TestDumpLoadIndex(t *testing.T) {
	ix1 := NewIndex(fruit)

	for k, s := range pairs {
		for _, b := range s {
			ix1.Add(k, b)
		}
	}

	var err error

	buf := bytes.NewBuffer(nil)

	if err = DumpIndex(buf, ix1); err != nil {
		t.Error(err)
	}

	var ix2 *Index

	if ix2, err = LoadIndex(buf); err != nil {
		t.Error(err)
	}

	mems := ix2.Domain.Members()

	if len(mems) != len(fruit) {
		t.Fatalf("expected %d members, got %d", len(fruit), len(mems))
	}

	for i, x := range mems {
		if x != fruit[i] {
			t.Errorf("expected member %d, got %d", fruit[i], x)
		}
	}

	for _, k := range ix2.Table.Keys() {
		// Check all bits are properly set.
		for _, m := range pairs[k] {
			if !ix2.Has(k, m) {
				t.Errorf("key %d should have member %d", k, m)
			}
		}
	}
}

func BenchmarkDumpIndex(b *testing.B) {
	ix := NewIndex(fruit)

	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(nil)
		DumpIndex(buf, ix)
	}
}

func BenchmarkLoadIndex(b *testing.B) {
	ix := NewIndex(fruit)
	buf := bytes.NewBuffer(nil)
	DumpIndex(buf, ix)

	for i := 0; i < b.N; i++ {
		LoadIndex(buf)
	}
}
