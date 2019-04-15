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
		ir:  reader,
		buf: make([]byte, bufSz),
	}
}

// NewWriter initializes a RESP writer to write to given io.Writer.
func NewWriter(wr io.Writer) *Writer {
	return &Writer{
		w: wr,
	}
}

// Writer provides functions for writing RESP protocol values.
type Writer struct {
	w io.Writer
}

func (rw *Writer) Write(v Value) (int, error) {
	return rw.w.Write([]byte(v.Serialize()))
}

// Reader represents RESP protocol reader.
type Reader struct {
	ir    io.Reader
	start int
	end   int
	buf   []byte
	sz    int
}

// Read reads the next command available from the stream.
func (rd *Reader) Read() (*MultiBulk, error) {
	var mb MultiBulk

	if err := rd.buffer(false); err != nil {
		return nil, err
	}

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

func (rd *Reader) readInline(mb *MultiBulk) error {
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

func (rd *Reader) readMultiBulk(mb *MultiBulk) error {
	rd.start++ // skip the '*' character

	size, err := rd.readNumber()
	if err != nil {
		return err
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

func (rd *Reader) readBulkStr() (*BulkStr, error) {
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

	data, err := rd.readExactly(size)
	if err != nil {
		return nil, err
	}
	rd.start += 2 // skip over CRLF

	return &BulkStr{
		Value: data,
	}, nil
}

func (rd *Reader) readNumber() (int, error) {
	data, err := rd.readTillCRLF()
	if err != nil {
		return 0, err
	}

	return toInt(data)
}

func (rd *Reader) readExactly(n int) ([]byte, error) {
	for rd.end-rd.start < n {
		if err := rd.buffer(true); err != nil {
			return nil, err
		}
	}

	data := rd.buf[rd.start : rd.start+n]
	rd.start += n
	return data, nil
}

func (rd *Reader) readTillCRLF() ([]byte, error) {
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

func (rd *Reader) buffer(force bool) error {
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
	rd.end = n

	return nil
}

func toInt(data []byte) (int, error) {
	var d int
	L := len(data)
	for i, b := range data {
		if b < '0' || b > '9' {
			return 0, errors.New("invalid number format")
		}

		if b == '0' {
			continue
		}

		pos := int(math.Pow(10, float64(L-i-1)))

		d += int(b-'0') * pos
	}

	return d, nil
}
