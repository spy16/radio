package radio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
)

const bufSz = 4096

// NewReader initializes the reader with default buffer size.
func NewReader(reader io.Reader) *Reader {
	return &Reader{
		bufferedReader: &bufferedReader{
			ir:  reader,
			buf: make([]byte, bufSz),
			sz:  bufSz,
		},
	}
}

// NewReaderSize initializes the reader with given buffer size.
func NewReaderSize(reader io.Reader, size int) *Reader {
	return &Reader{
		bufferedReader: &bufferedReader{
			ir:  reader,
			buf: make([]byte, size),
			sz:  size,
		},
	}
}

// Reader implements server-side RESP protocol parsing.
type Reader struct {
	*bufferedReader

	input bool
}

// Read reads the next command available from the stream.
func (rd *Reader) Read() (*MultiBulk, error) {
	if err := rd.buffer(false); err != nil {
		return nil, err
	}

	var mb MultiBulk
	if rd.buf[rd.start] == '*' {
		if err := rd.readMultiBulk(&mb); err != nil {
			return nil, err
		}
	} else {
		if err := rd.readInline(&mb); err != nil {
			return nil, err
		}
	}

	return &mb, nil
}

func (rd *Reader) readMultiBulk(mb *MultiBulk) error {
	rd.start++ // skip the '*' character

	size, err := rd.readNumber()
	if err != nil {
		return err
	}

	if size >= 0 {
		mb.Items = []BulkStr{}
	} else {
		// negative size -> nil multi-bulk
		return nil
	}

	for i := 0; i < int(size); i++ {
		itm, err := rd.readBulkStr()
		if err != nil {
			return err
		}

		mb.Items = append(mb.Items, *itm)
	}

	return nil
}

func toInt(data []byte) (int, error) {
	var d, sign int
	L := len(data)
	for i, b := range data {
		if i == 0 {
			if b == '-' {
				sign = -1
				continue
			}

			sign = 1
		}

		if b < '0' || b > '9' {
			return 0, errors.New("invalid number format")
		}

		if b == '0' {
			continue
		}

		pos := int(math.Pow(10, float64(L-i-1)))

		d += int(b-'0') * pos
	}

	return sign * d, nil
}

type bufferedReader struct {
	ir    io.Reader
	start int
	end   int
	buf   []byte
	sz    int
}

func (rd *bufferedReader) readNumber() (int, error) {
	data, err := rd.readTillCRLF()
	if err != nil {
		return 0, err
	}

	return toInt(data)
}

func (rd *bufferedReader) readExactly(n int) ([]byte, error) {
	for rd.end-rd.start < n {
		if err := rd.buffer(true); err != nil {
			return nil, err
		}
	}

	data := rd.buf[rd.start : rd.start+n]
	rd.start += n
	return data, nil
}

func (rd *bufferedReader) readTillCRLF() ([]byte, error) {
	var crlf int
	for crlf = bytes.Index(rd.buf[rd.start:rd.end], []byte("\r\n")); crlf < 0; {
		if err := rd.buffer(true); err != nil {
			return nil, err
		}
	}

	data := make([]byte, crlf)
	copy(data, rd.buf[rd.start:rd.end])
	rd.start += crlf + 2
	return data, nil
}

func (rd *bufferedReader) buffer(force bool) error {
	if !force && rd.end > rd.start {
		return nil // buffer already has some data.
	}

	if rd.end > 0 && rd.start >= rd.end {
		rd.start = 0
		rd.end = 0
	} else if rd.end == len(rd.buf) {
		rd.buf = append(rd.buf, make([]byte, rd.sz)...)
	}

	n, err := rd.ir.Read(rd.buf[rd.end:])
	if err != nil {
		return err
	}
	rd.end += n

	return nil
}

func (rd *bufferedReader) readInline(mb *MultiBulk) error {
	var crlf int
	for crlf = bytes.Index(rd.buf[rd.start:rd.end], []byte("\r\n")); crlf < 0; {
		if err := rd.buffer(true); err != nil {
			return err
		}
	}

	// TODO: split the string using space as delimiter while
	// taking escape sequences into consideration.

	mb.Items = append(mb.Items, BulkStr{
		Value: rd.buf[rd.start : rd.start+crlf],
	})
	rd.start += crlf + 2

	return nil
}

func (rd *bufferedReader) readBulkStr() (*BulkStr, error) {
	if err := rd.buffer(false); err != nil {
		return nil, err
	}

	if rd.buf[rd.start] != '$' {
		return nil, fmt.Errorf("Protocol error: expecting '$', got '%c'", rd.buf[rd.start])
	}
	rd.start++

	size, err := rd.readNumber()
	if err != nil {
		return nil, err
	}

	if size < 0 {
		return &BulkStr{}, nil
	}

	data, err := rd.readExactly(size)
	if err != nil {
		return nil, err
	}
	rd.start += 2 // skip over CRLF

	return &BulkStr{
		Value: data,
	}, nil
}
