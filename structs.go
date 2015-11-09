package bitindex

type Uint32Array []uint32

func (a Uint32Array) Len() int {
	return len(a)
}

func (a Uint32Array) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Uint32Array) Less(i, j int) bool {
	return a[i] < a[j]
}

type Uint32Set map[uint32]struct{}

func (s Uint32Set) Items() []uint32 {
	a := make([]uint32, len(s))
	i := 0

	for k, _ := range s {
		a[i] = k
		i++
	}

	return a
}

func (s Uint32Set) Len() int {
	return len(s)
}

func (s Uint32Set) Contains(i uint32) bool {
	_, ok := s[i]
	return ok
}

func (s Uint32Set) Add(is ...uint32) {
	for _, i := range is {
		s[i] = struct{}{}
	}
}

func (s Uint32Set) Remove(is ...uint32) {
	for _, i := range is {
		delete(s, i)
	}
}

func (s Uint32Set) Clear() {
	for k, _ := range s {
		delete(s, k)
	}
}

func (s Uint32Set) Intersect(b Uint32Set) Uint32Set {
	o := make(Uint32Set)

	// Pick the smallest set.
	x := s

	if len(b) < len(s) {
		x = b
	}

	for k, _ := range x {
		if _, ok := b[k]; ok {
			o[k] = struct{}{}
		}
	}

	return o
}
