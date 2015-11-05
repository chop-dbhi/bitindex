package bitindex

import (
	"encoding/csv"
	"io"
)

// CSVIndexer is an indexer for CSV structured data.
type CSVIndexer struct {
	*csv.Reader

	// If true, the first line will be skipped.
	Header bool

	// A function that takes a CSV row and returns the key and member
	// to be index.
	Parse func([]string) (uint32, uint32, error)
}

// NewCSVIndexer initializes a new CSV parser for building an index.
func NewCSVIndexer(r io.Reader) *CSVIndexer {
	return &CSVIndexer{
		Reader: csv.NewReader(r),
	}
}

// Build implements the Indexer interface and builds an index from a CSV file.
func (p *CSVIndexer) Index() (*Index, error) {
	ix := NewIndex(nil)

	// Skip the header.
	if p.Header {
		_, err := p.Read()

		// Includes EOF.
		if err != nil {
			return nil, err
		}
	}

	var k, m uint32

	for {
		row, err := p.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if k, m, err = p.Parse(row); err != nil {
			return nil, err
		}

		ix.Add(k, m)
	}

	return ix, nil
}
