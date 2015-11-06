package bitindex

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Resets the buffer
func clearBuffer(s []byte) {
	for i, _ := range s {
		s[i] = 0x0
	}
}

func writeInt(w io.Writer, b []byte, i int) error {
	binary.PutUvarint(b, uint64(i))

	if n, err := w.Write(b[:4]); err != nil {
		return err
	} else if n != 4 {
		return fmt.Errorf("Expected to write 4 bytes, wrote %d", n)
	}

	return nil
}

func writeUint32(w io.Writer, b []byte, i uint32) error {
	binary.PutUvarint(b, uint64(i))

	if n, err := w.Write(b[:4]); err != nil {
		return err
	} else if n != 4 {
		return fmt.Errorf("Expected to write 4 bytes, wrote %d", n)
	}

	return nil
}

func dumpDomain(w io.Writer, d *Domain, b []byte) error {
	// Length of the domain. 4 bytes.
	if err := writeInt(w, b, d.Size()); err != nil {
		return fmt.Errorf("Error writing domain length: %s", err)
	}

	// Domain members. 4 bytes each.
	for _, n := range d.Members() {
		clearBuffer(b)

		if err := writeUint32(w, b, n); err != nil {
			return fmt.Errorf("Error writing domain member: %s", err)
		}
	}

	return nil
}

func dumpArray(w io.Writer, a Array, b []byte) error {
	clearBuffer(b)

	// Array length. 4 bytes.
	if err := writeInt(w, b, len(a)); err != nil {
		return fmt.Errorf("Error writing array length: %s", err)
	}

	// Encode array items (which is a map).
	for p, y := range a {
		clearBuffer(b)

		// Byte position.
		if err := writeUint32(w, b, p); err != nil {
			return fmt.Errorf("Error writing byte position: %s", err)
		}

		b[0] = y

		if n, err := w.Write(b[:1]); err != nil {
			return fmt.Errorf("Error writing byte: %s", err)
		} else if n != 1 {
			return fmt.Errorf("Expected to write 1 byte, wrote %d", n)
		}
	}

	return nil
}

func dumpTable(w io.Writer, t Table, b []byte) error {
	clearBuffer(b)

	// Length of the table. 4 bytes.
	if err := writeInt(w, b, t.Size()); err != nil {
		return fmt.Errorf("Error writing table length: %s", err)
	}

	// Encode table entries.
	for k, a := range t {
		clearBuffer(b)

		// Entry key. 4 bytes.
		if err := writeUint32(w, b, k); err != nil {
			return fmt.Errorf("Error writing array key: %s", err)
		}

		if err := dumpArray(w, a, b); err != nil {
			return err
		}
	}

	return nil
}

func dumpIndex(w io.Writer, idx *Index) error {
	// Shared buffer. Nothing exceeds 4 bytes.
	b := make([]byte, 4, 4)

	if err := dumpDomain(w, idx.Domain, b); err != nil {
		return err
	}

	if err := dumpTable(w, idx.Table, b); err != nil {
		return err
	}

	return nil
}

// DumpIndex writes an Index to it binary representation.
func DumpIndex(w io.Writer, idx *Index) error {
	if err := dumpIndex(w, idx); err != nil {
		return err
	}

	return nil
}

func readUint32(r io.Reader, b []byte) (uint32, error) {
	if len(b) != 4 {
		panic("need 4 bytes")
	}
	if n, err := r.Read(b[:4]); err != nil {
		return 0, err
	} else if n != 4 {
		return 0, fmt.Errorf("Expected to read 4 bytes; read %d", n)
	}

	val, ierr := binary.Uvarint(b)

	if ierr <= 0 {
		panic("error decoding int")
	}

	return uint32(val), nil
}

func readInt(r io.Reader, b []byte) (int, error) {
	if n, err := r.Read(b[:4]); err != nil {
		return 0, err
	} else if n != 4 {
		return 0, fmt.Errorf("Expected to read 4 bytes; read %d", n)
	}

	val, ierr := binary.Uvarint(b)

	if ierr <= 0 {
		panic("error decoding int")
	}

	return int(val), nil
}

func readByte(r io.Reader, b []byte) (byte, error) {
	if n, err := r.Read(b[:1]); err != nil {
		return 0, err
	} else if n != 1 {
		return 0, fmt.Errorf("Expected to read 1 byte; read %d", n)
	}

	return b[0], nil
}

func readArray(r io.Reader, n int, b []byte) (Array, error) {
	a := make(Array, n)

	var (
		pos uint32
		bb  byte
		err error
	)

	for i := 0; i < n; i++ {
		// Read byte position.
		if pos, err = readUint32(r, b); err != nil {
			return nil, fmt.Errorf("Error reading byte position at %d: %s", i, err)
		}

		// Byte entry.
		if bb, err = readByte(r, b); err != nil {
			return nil, fmt.Errorf("Error reading byte at %d: %s", i, err)
		}

		a[pos] = bb
	}

	return a, nil
}

func readDomain(r io.Reader, b []byte) (*Domain, error) {
	var (
		n   int
		m   uint32
		err error
	)

	// N members in the domain.
	if n, err = readInt(r, b); err != nil {
		return nil, fmt.Errorf("Error decoding domain length: %s", err)
	}

	// Initialize array for members.
	ms := make([]uint32, n)

	// Domain members are encoded as an array of 4 bytes
	for i := 0; i < n; i++ {
		if m, err = readUint32(r, b); err != nil {
			return nil, fmt.Errorf("Error decoding domain member at %d: %s", i, err)
		}

		ms[i] = m
	}

	// Initialize and set the domain.
	return NewDomain(ms), nil
}

func readTable(r io.Reader, b []byte) (Table, error) {
	var (
		n   int
		err error
	)

	// N entries in the table.
	if n, err = readInt(r, b); err != nil {
		return nil, fmt.Errorf("Error decoding table length: %s", err)
	}

	t := make(Table, n)

	var (
		l int
		k uint32
		a Array
	)

	// Decode all table entries.
	for i := 0; i < n; i++ {
		// Key for the entry.
		if k, err = readUint32(r, b); err != nil {
			return nil, fmt.Errorf("Error decoding array key: %s", err)
		}

		// Decode array length.
		if l, err = readInt(r, b); err != nil {
			return nil, fmt.Errorf("Error decoding array length: %s", err)
		}

		if a, err = readArray(r, l, b); err != nil {
			return nil, err
		}

		// Add entry to table.
		t[k] = a
	}

	return t, nil
}

// LoadDomain loads only the domain from an io.Reader.
func LoadDomain(r io.Reader) (*Domain, error) {
	var (
		d   *Domain
		err error
	)

	b := make([]byte, 4, 4)

	if d, err = readDomain(r, b); err != nil {
		return nil, err
	}

	return d, nil
}

// LoadIndex loads the index from an io.Reader.
func LoadIndex(r io.Reader) (*Index, error) {
	var (
		d   *Domain
		t   Table
		err error
	)

	b := make([]byte, 4, 4)

	if d, err = readDomain(r, b); err != nil {
		return nil, err
	}

	if t, err = readTable(r, b); err != nil {
		return nil, err
	}

	idx := &Index{
		Domain: d,
		Table:  t,
	}

	return idx, nil
}
