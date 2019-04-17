package radio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
)

const defaultBufSize = 4096

// NewReader initializes the reader with default buffer size.
func NewReader(reader io.Reader, serverMode bool) *Reader {
	return &Reader{
		ir:         reader,
		buf:        make([]byte, defaultBufSize),
		sz:         defaultBufSize,
		serverMode: serverMode,
	}
}

// Reader implements both Server side and Client side RESP protocol parsing.
type Reader struct {
	ir    io.Reader
	start int
	end   int
	buf   []byte
	sz    int

	serverMode bool
}

func (rd *Reader) Read() (Value, error) {
	if err := rd.buffer(false); err != nil {
		return nil, err
	}

	prefix := rd.buf[rd.start]
	if rd.serverMode {
		var mb MultiBulk
		var err error
		if prefix == '*' {
			err = rd.readMultiBulk(&mb)
		} else {
			err = rd.readInline(&mb)
		}

		if err != nil {
			return nil, err
		}

		return &mb, nil
	}

	switch rd.buf[rd.start] {
	case '+':
		return rd.readSimpleStr()

	case '-':
		return rd.readErrorStr()

	case ':':
		return rd.readInteger()

	case '$':
		bs, err := rd.readBulkStr()
		if err != nil {
			return nil, err
		}
		return bs, nil

	case '*':
		arr, err := rd.readArray()
		if err != nil {
			return nil, err
		}
		return arr, nil

	}

	return nil, fmt.Errorf("unexpected byte '%c'", rd.buf[rd.start])
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

func (rd *Reader) readNumber() (int, error) {
	data, err := rd.readTillCRLF()
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, errors.New("no number")
	}

	return toInt(data)
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

func (rd *Reader) readInteger() (Integer, error) {
	rd.start++ // skip over ':'

	n, err := rd.readNumber()
	return Integer(n), err
}

func (rd *Reader) readArray() (*Array, error) {
	rd.start++ // skip over '+'

	size, err := rd.readNumber()
	if err != nil {
		return nil, err
	}

	if size < 0 {
		return &Array{}, nil
	}

	arr := &Array{}
	arr.Items = []Value{}

	for i := 0; i < size; i++ {
		item, err := rd.Read()
		if err != nil {
			return nil, err
		}

		arr.Items = append(arr.Items, item)
	}

	return arr, nil
}

func (rd *Reader) readSimpleStr() (SimpleStr, error) {
	rd.start++ // skip over '+'

	data, err := rd.readTillCRLF()
	return SimpleStr(data), err
}

func (rd *Reader) readErrorStr() (ErrorStr, error) {
	rd.start++ // skip over '-'

	data, err := rd.readTillCRLF()
	return ErrorStr(data), err
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
			if err == io.EOF {
				break
			}

			return nil, err
		}
	}

	if crlf == 0 {
		return []byte(""), nil
	} else if crlf < 0 {
		data := rd.buf[rd.start:rd.end]
		rd.start = rd.end
		return data, io.EOF
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
	rd.end += n

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
