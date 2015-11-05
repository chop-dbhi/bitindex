package bitindex

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
)

func writeInt(w io.Writer, b []byte, i int) error {
	binary.PutUvarint(b, uint64(i))

	if n, err := w.Write(b[:4]); err != nil {
		return err
	} else if n != 4 {
		panic("expected to write 4 bytes")
	}

	return nil
}

func writeUint32(w io.Writer, b []byte, i uint32) error {
	binary.PutUvarint(b, uint64(i))

	if n, err := w.Write(b[:4]); err != nil {
		return err
	} else if n != 4 {
		panic("expected to write 4 bytes")
	}

	return nil
}

func dump(w io.Writer, idx *Index) error {
	var (
		err error
		n   int
		un  uint32

		// All parts fit into 4 bytes.
		byts = make([]byte, 4, 4)

		buf = bytes.NewBuffer(byts)
	)

	// Length of the domain. 4 bytes.
	n = idx.Domain.Size()

	if err := writeInt(w, byts, n); err != nil {
		return fmt.Errorf("Error writing domain length: %s", err)
	}

	// Domain members. 4 bytes.
	for _, un = range idx.Domain.r {
		buf.Reset()

		if err = writeUint32(w, byts, un); err != nil {
			return fmt.Errorf("Error writing domain member: %s", err)
		}
	}

	buf.Reset()

	// Length of the table. 4 bytes.
	n = idx.Table.Size()

	if err := writeInt(w, byts, n); err != nil {
		return fmt.Errorf("Error writing table length: %s", err)
	}

	var (
		a Array
		p uint32
		b byte
	)

	// Encode table entries.
	for un, a = range idx.Table {
		buf.Reset()

		// Entry key. 4 bytes.
		if err = writeUint32(w, byts, un); err != nil {
			return fmt.Errorf("Error writing array key: %s", err)
		}

		buf.Reset()

		// Array length. 4 bytes.
		if err = writeInt(w, byts, len(a)); err != nil {
			return fmt.Errorf("Error writing array length: %s", err)
		}

		// Encode array items.
		for p, b = range a {
			buf.Reset()

			// Byte position.
			if err = writeUint32(w, byts, p); err != nil {
				return fmt.Errorf("Error writing byte position: %s", err)
			}

			byts[0] = b

			if _, err = w.Write(byts[:1]); err != nil {
				return fmt.Errorf("Error writing byte: %s", err)
			}
		}
	}

	return nil
}

// Dump writes an Index to it binary representation.
func Dump(w io.Writer, idx *Index) error {
	gw := gzip.NewWriter(w)

	if err := dump(gw, idx); err != nil {
		return err
	}

	gw.Flush()
	gw.Close()

	return nil
}

func readUint32(r io.Reader, b []byte) (uint32, error) {
	if n, err := r.Read(b[:4]); err != nil {
		return 0, err
	} else if n != 4 {
		panic("expected 4 bytes to be read")
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
		panic("expected 4 bytes to be read")
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
		panic("expected 1 byte to be read")
	}

	return b[0], nil
}

func load(r io.Reader, idx *Index) error {
	var (
		n   int
		un  uint32
		err error

		// All parts fit into 4 bytes.
		byts = make([]byte, 4, 4)
	)

	// N members in the domain.
	if n, err = readInt(r, byts); err != nil {
		return fmt.Errorf("Error decoding domain length: %s", err)
	}

	// Initialize array for members.
	members := make([]uint32, n)

	// Domain members are encoded as an array of 4 bytes
	for i := 0; i < n; i++ {
		if un, err = readUint32(r, byts); err != nil {
			return fmt.Errorf("Error decoding domain member: %s", err)
		}

		members[i] = un
	}

	// Initialize and set the domain.
	idx.Domain = NewDomain(members)

	// N entries in the table.
	if n, err = readInt(r, byts); err != nil {
		return fmt.Errorf("Error decoding table length: %s", err)
	}

	var (
		key, pos uint32
		eb       byte
		a        Array
		m        int
	)

	// Decode all table entries.
	for i := 0; i < n; i++ {
		// Key for the entry.
		if key, err = readUint32(r, byts); err != nil {
			return fmt.Errorf("Error decoding array key: %s", err)
		}

		// Decode array length.
		if m, err = readInt(r, byts); err != nil {
			return fmt.Errorf("Error decoding array length: %s", err)
		}

		// Initialize the array.
		a = make(Array, m)

		// Array items.
		for i := 0; i < m; i++ {
			if pos, err = readUint32(r, byts); err != nil {
				return fmt.Errorf("Error reading byte position: %s", err)
			}

			// Byte entry.
			if eb, err = readByte(r, byts); err != nil {
				return fmt.Errorf("Error reading byte: %s", err)
			}

			a[pos] = eb
		}

		// Add entry to table.
		idx.Table[key] = a
	}

	return nil
}

func Load(r io.Reader, idx *Index) error {
	gr, err := gzip.NewReader(r)

	if err != nil {
		return err
	}

	if err = load(gr, idx); err != nil {
		return err
	}

	gr.Close()

	return nil
}
