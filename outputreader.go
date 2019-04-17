package radio

import (
	"fmt"
	"io"
)

// NewOutputReader initializes a buffered RESP reader for server output.
func NewOutputReader(rd io.Reader) *OutputReader {
	return &OutputReader{
		bufferedReader: &bufferedReader{
			ir: rd,
			sz: bufSz,
		},
	}
}

// OutputReader implements RESP protocol parser for server output.
type OutputReader struct {
	*bufferedReader

	inputParser bool
}

// Read consumes data from the stream and returns the next RESP value.
func (rd *OutputReader) Read() (Value, error) {
	if err := rd.buffer(false); err != nil {
		return nil, err
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

func (rd *OutputReader) readInteger() (Integer, error) {
	rd.start++ // skip over ':'

	n, err := rd.readNumber()
	return Integer(n), err
}

func (rd *OutputReader) readArray() (*Array, error) {
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

func (rd *OutputReader) readSimpleStr() (SimpleStr, error) {
	rd.start++ // skip over '+'

	data, err := rd.readTillCRLF()
	return SimpleStr(data), err
}

func (rd *OutputReader) readErrorStr() (ErrorStr, error) {
	rd.start++ // skip over '-'

	data, err := rd.readTillCRLF()
	return ErrorStr(data), err
}
