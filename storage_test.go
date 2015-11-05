package bitindex

import (
	"bytes"
	"testing"
)

func TestDumpLoadIndex(t *testing.T) {
	ix := NewIndex(NewDomain(fruit))

	for _, p := range pairs {
		ix.Add(p[0], p[1])
	}

	buf := bytes.NewBuffer(nil)

	if err := Dump(buf, ix); err != nil {
		t.Error(err)
	}

	ix2 := NewIndex(nil)

	if err := Load(buf, ix2); err != nil {
		t.Error(err)
	}

	//ix2.Domain
}

func BenchmarkDumpIndex(b *testing.B) {
	ix := NewIndex(NewDomain(fruit))

	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(nil)
		Dump(buf, ix)
	}
}

func BenchmarkLoadIndex(b *testing.B) {
	ix := NewIndex(NewDomain(fruit))
	buf := bytes.NewBuffer(nil)
	Dump(buf, ix)

	for i := 0; i < b.N; i++ {
		ix := NewIndex(nil)
		Load(buf, ix)
	}
}
