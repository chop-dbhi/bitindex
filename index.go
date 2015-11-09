package bitindex

import (
	"fmt"
	"math"
)

// Domain maps a member to a position in the bit array.
type Domain struct {
	// Bit array index.
	b uint32

	// Member -> Bit. This is necessary to construct a bit mask
	// for a set of members that are being tested.
	f map[uint32]uint32

	// Bit -> Member
	r []uint32
}

// Members returns the members in the domain.
func (d *Domain) Members() []uint32 {
	return d.r
}

// Add adds a member to the domain.
func (d *Domain) Add(m uint32) uint32 {
	// Do not add duplicates.
	if b, ok := d.f[m]; ok {
		return b
	}

	d.f[m] = d.b

	if uint32(len(d.r)) > d.b {
		d.r[d.b] = m
	} else {
		d.r = append(d.r, m)
	}

	b := d.b
	d.b++

	return b
}

// Bit returns the bit of the member.
func (d *Domain) Bit(m uint32) uint32 {
	if b, ok := d.f[m]; ok {
		return b
	}

	panic("member not in domain")
}

// Member returns the member for the bit.
func (d *Domain) Member(b uint32) uint32 {
	return d.r[b]
}

// Size returns the size of the domain, which is also the number of bits.
func (d *Domain) Size() int {
	return int(d.b)
}

// Bytes returns the number of bytes required for the domain.
func (d *Domain) Bytes() int {
	return int(math.Ceil(float64(d.b) / 8.0))
}

// Mask builds a bitmask for a subset of members in the domain.
func (d *Domain) Mask(ms ...uint32) ([]uint32, error) {
	var (
		ok bool
		b  uint32
	)

	a := make([]uint32, len(ms))

	for i, m := range ms {
		if b, ok = d.f[m]; !ok {
			return nil, fmt.Errorf("%d is not a member", m)
		}

		a[i] = b
	}

	return a, nil
}

func NewDomain(ms []uint32) *Domain {
	l := len(ms)

	f := make(map[uint32]uint32, l)

	for i, m := range ms {
		f[m] = uint32(i)
	}

	return &Domain{
		b: uint32(l),
		f: f,
		r: ms,
	}
}

// Loc is the location of the bit in array consisting of the byte
// position and the relative bit.
type Loc [2]uint32

// Array represents an array of bytes.
type Array map[uint32]byte

// Size returns the number of bytes used by the array.
func (a Array) Bytes() int {
	return len(a)
}

// Set sets the bit to 1.
func (a Array) Set(bit uint32) {
	off := bit / 8
	bit = bit % 8

	if _, ok := a[off]; !ok {
		a[off] = 1 << bit
	} else {
		a[off] |= (1 << bit)
	}
}

// Clear sets the bit to 0.
func (a Array) Clear(bit uint32) {
	off := bit / 8
	bit = bit % 8

	if _, ok := a[off]; ok {
		a[off] &= ^(1 << bit)
	}
}

// Flip flips the bit.
func (a Array) Flip(bit uint32) bool {
	t := a.Has(bit)

	if t {
		a.Clear(bit)
	} else {
		a.Set(bit)
	}

	return !t
}

// Has returns true if the bit is set.
func (a Array) Has(bit uint32) bool {
	off := bit / 8
	bit = bit % 8

	if bt, ok := a[off]; ok {
		v := bt & (1 << bit)
		return v > 0
	}

	return false
}

// Loc returns the location of the bit in this array.
func (a Array) Loc(bit uint32) Loc {
	return Loc{bit / 8, bit % 8}
}

// Any returns true any of the bits set.
func (a Array) Any(bits ...uint32) bool {
	for _, bit := range bits {
		if a.Has(bit) {
			return true
		}
	}

	return false
}

// All returns true if all of the bits are set.
func (a Array) All(bits ...uint32) bool {
	for _, bit := range bits {
		if !a.Has(bit) {
			return false
		}
	}

	return true
}

// NotAny returns true if any of the bits are not set.
func (a Array) NotAny(bits ...uint32) bool {
	for _, bit := range bits {
		if a.Has(bit) {
			return false
		}
	}

	return true
}

// NotAll returns true if all of the bits are not set.
func (a Array) NotAll(bits ...uint32) bool {
	for _, bit := range bits {
		if !a.Has(bit) {
			return true
		}
	}

	return false
}

// NewArray initializes a new array.
func NewArray() Array {
	return make(Array)
}

// Table is an map of keys to bit arrays.
type Table map[uint32]Array

// Keys returns all keys in the table.
func (t Table) Keys() []uint32 {
	a := make([]uint32, len(t))
	i := 0

	for k, _ := range t {
		a[i] = k
		i++
	}

	return a
}

// Get gets the bit array for the key.
func (t Table) Get(k uint32) Array {
	return t[k]
}

// Size returns the number of items in the table.
func (t Table) Size() int {
	return len(t)
}

// Bytes returns the number of bytes allocated in the table.
func (t Table) Bytes() int {
	var c int

	for _, a := range t {
		c += a.Bytes()
	}

	return c
}

// Set sets the bit for key `k`. The array for `k` is created if it does not exist.
func (t Table) Set(k uint32, b uint32) {
	a, ok := t[k]

	if !ok {
		a = NewArray()
		t[k] = a
	}

	a.Set(b)
}

// Index combines and domain
type Index struct {
	Domain *Domain
	Table  Table
}

// Add adds sets the bit for key `k` for member `m` in the domain.
func (ix *Index) Add(k uint32, m uint32) {
	b := ix.Domain.Add(m)

	ix.Table.Set(k, b)
}

// Has returns true if the key has the member.
func (ix *Index) Has(k uint32, m uint32) bool {
	b := ix.Domain.Bit(m)
	return ix.Table.Get(k).Has(b)
}

// Any returns all keys that match any of the passed members.
func (ix *Index) Any(ms ...uint32) ([]uint32, error) {
	// Get the mask.
	bs, err := ix.Domain.Mask(ms...)

	if err != nil {
		return nil, err
	}

	var keys []uint32

	for k, a := range ix.Table {
		if a.Any(bs...) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// All returns all keys that match all of the passed members.
func (ix *Index) All(ms ...uint32) ([]uint32, error) {
	// Get the mask.
	bs, err := ix.Domain.Mask(ms...)

	if err != nil {
		return nil, err
	}

	var keys []uint32

	for k, a := range ix.Table {
		if a.All(bs...) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// NotAny returns all keys that do not match any of the passed members.
func (ix *Index) NotAny(ms ...uint32) ([]uint32, error) {
	// Get the mask.
	bs, err := ix.Domain.Mask(ms...)

	if err != nil {
		return nil, err
	}

	var keys []uint32

	for k, a := range ix.Table {
		if a.NotAny(bs...) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// NotAll returns all keys that do not match all of the passed members.
func (ix *Index) NotAll(ms ...uint32) ([]uint32, error) {
	// Get the mask.
	bs, err := ix.Domain.Mask(ms...)

	if err != nil {
		return nil, err
	}

	var keys []uint32

	for k, a := range ix.Table {
		if a.NotAll(bs...) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

func (ix *Index) Query(any, all, nany, nall []uint32) (*Result, error) {
	var (
		err  error
		set  Uint32Set
		tmp  = make(Uint32Set)
		keys []uint32
	)

	if any != nil {
		if keys, err = ix.Any(any...); err != nil {
			return nil, fmt.Errorf("Operation failed (any): %s\n", err)
		}

		if set == nil {
			set = make(Uint32Set, len(keys))
			set.Add(keys...)
		} else {
			tmp.Add(keys...)
			set = set.Intersect(tmp)
			tmp.Clear()
		}
	}

	if all != nil {
		if keys, err = ix.All(all...); err != nil {
			return nil, fmt.Errorf("Operation failed (all): %s\n", err)
		}

		if set == nil {
			set = make(Uint32Set, len(keys))
			set.Add(keys...)
		} else {
			tmp.Add(keys...)
			set = set.Intersect(tmp)
			tmp.Clear()
		}
	}

	if nany != nil {
		if keys, err = ix.NotAny(nany...); err != nil {
			return nil, fmt.Errorf("Operation failed (nany): %s\n", err)
		}

		if set == nil {
			set = make(Uint32Set, len(keys))
			set.Add(keys...)
		} else {
			tmp.Add(keys...)
			set = set.Intersect(tmp)
			tmp.Clear()
		}
	}

	if nall != nil {
		if keys, err = ix.NotAll(nall...); err != nil {
			return nil, fmt.Errorf("Operation failed (nall): %s\n", err)
		}

		if set == nil {
			set = make(Uint32Set, len(keys))
			set.Add(keys...)
		} else {
			tmp.Add(keys...)
			set = set.Intersect(tmp)
			tmp.Clear()
		}
	}

	return &Result{
		set: set,
		idx: ix,
	}, nil
}

// Sparsity returns the proportion of bits being represented in the domain
// to the bytes being allocated in the index.
func (ix *Index) Sparsity() float32 {
	alloc := float32(ix.Table.Bytes())
	avg := alloc / float32(ix.Table.Size())
	return 1 - avg/float32(ix.Domain.Bytes())
}

// NewIndex initializes a new index.
func NewIndex(d []uint32) *Index {
	return &Index{
		Domain: NewDomain(d),
		Table:  make(Table),
	}
}

// Indexer is an interface for defines a method for building an index
// from a source.
type Indexer interface {
	Index() (*Index, error)
}

type Result struct {
	set Uint32Set
	idx *Index
}

func (r *Result) Items() []uint32 {
	return r.set.Items()
}

func (r *Result) Len() int {
	return r.set.Len()
}

func (r *Result) Smallest(thres float32) bool {
	return float32(r.Len())/float32(len(r.idx.Table)) < thres
}

func (r *Result) Complement() []uint32 {
	items := make([]uint32, len(r.idx.Table)-r.set.Len())
	i := 0

	for k, _ := range r.idx.Table {
		if !r.set.Contains(k) {
			items[i] = k
			i++
		}
	}

	return items
}
