package radio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
)

const defaultBufSize = 4096

// ErrBufferFull is returned when there is no space left on the buffer to read
// more data.
var ErrBufferFull = errors.New("buffer is full")

// NewReader initializes the RESP reader with given reader. In server mode,
// input data will be read line-by-line except in case of array of bulkstrings.
//
// Read https://redis.io/topics/protocol#sending-commands-to-a-redis-server for
// more information on how clients interact with server.
func NewReader(r io.Reader, isServer bool) *Reader {
	return NewReaderSize(r, isServer, defaultBufSize)
}

// NewReaderSize initializes the RESP reader with given buffer size.
// See NewReader for more information.
func NewReaderSize(r io.Reader, isServer bool, size int) *Reader {
	return &Reader{
		ir:       r,
		IsServer: isServer,
		buf:      make([]byte, size),
		sz:       size,
	}
}

// Reader implements server and client RESP protocol parser. IsServer flag
// controls the RESP parsing mode. When IsServer set to true, only Multi Bulk
// (Array of Bulk strings) and inline commands are supported. When IsServer set
// to false, all RESP values are enabled. FixedBuffer fields allows controlling
// the growth of buffer.
// Read https://redis.io/topics/protocol for RESP protocol specification.
type Reader struct {
	// IsServer controls the RESP parsing mode. If set, only inline string
	// and multi-bulk (array of bulk strings) will be enabled.
	IsServer bool

	// FixedBuffer if set does not allow the buffer to grow in case of
	// large incoming data and instead returns ErrBufferFull. If this is
	// false, buffer grows by doubling the buffer size as needed.
	FixedBuffer bool

	ir      io.Reader
	start   int
	end     int
	buf     []byte
	sz      int
	inArray bool

	vals []Value
}

func (rd *Reader) Read() (Value, error) {
	var err error
	for len(rd.vals) == 0 {
		err = rd.readAll()
		if err != nil && err != errIncompleteData {
			break
		}
	}

	if err == errIncompleteData {
		err = nil
	}

	if len(rd.vals) == 0 {
		return nil, err
	}

	v := rd.vals[0]
	rd.vals = rd.vals[1:]
	return v, err
}

// Size returns the current buffer size and the minimum buffer size
// reader is configured with.
func (rd *Reader) Size() (minSize int, currentSize int) {
	return rd.sz, len(rd.buf)
}

func (rd *Reader) readAll() error {
	if _, err := rd.buffer(); err != nil {
		return err
	}

	for rd.start < rd.end {
		val, err := rd.readOne()
		if err != nil {
			return err
		}

		rd.vals = append(rd.vals, val)
	}

	return nil
}

func (rd *Reader) readOne() (Value, error) {
	if rd.start >= rd.end {
		return nil, errIncompleteData
	}

	prefix := rd.buf[rd.start]

	if rd.IsServer {
		if rd.inArray && prefix != '$' {
			return nil, fmt.Errorf("Protocol error: expected '$', got '%c'", prefix)
		}

		if prefix != '*' && prefix != '$' {
			offset, v, err := readInline(rd.buf[rd.start:rd.end])
			if err != nil {
				return nil, err
			}
			rd.start += offset

			return v, nil
		}
	}

	switch prefix {
	case '+':
		data, err := tillCRLF(rd.buf[rd.start:rd.end])
		if err != nil {
			return nil, err
		}
		rd.start += len(data)

		value := SimpleStr(bytes.TrimRight(data[1:], "\r\n"))
		return value, nil

	case '-':
		data, err := tillCRLF(rd.buf[rd.start:rd.end])
		if err != nil {
			return nil, err
		}
		rd.start += len(data)

		value := ErrorStr(bytes.TrimRight(data[1:], "\r\n"))
		return value, nil

	case ':':
		offset, num, err := readNumber(rd.buf[rd.start:rd.end])
		if err != nil {
			return nil, err
		}
		rd.start += offset

		value := Integer(num)
		return value, nil

	case '$':
		consumed, data, err := readBulkStr(rd.buf[rd.start:rd.end], rd.IsServer)
		if err != nil {
			return nil, err
		}
		rd.start += consumed

		return &BulkStr{
			Value: data,
		}, nil

	case '*':
		rd.inArray = true
		defer func() {
			rd.inArray = false
		}()

		backupStart, backupEnd := rd.start, rd.end

		offset, size, err := readNumber(rd.buf[rd.start:rd.end])
		if err != nil {
			if rd.IsServer && (err == errInvalidNumber || err == errNoNumber) {
				return nil, errors.New("Protocol error: invalid multibulk length")
			}

			return nil, err
		}

		rd.start += offset

		var arr Array
		if size < 0 {
			// -1 (negative size) means a null array
			return &arr, nil
		} else if size == 0 {
			arr.Items = []Value{}
			return &arr, nil
		}

		for len(arr.Items) < size {
			val, err := rd.readOne()
			if err != nil {
				rd.start = backupStart
				rd.end = backupEnd
				return nil, err
			}

			if val != nil {
				arr.Items = append(arr.Items, val)
			}
		}

		return &arr, nil
	}

	return nil, fmt.Errorf("bad prefix '%c'", prefix)
}

func (rd *Reader) buffer() (int, error) {
	if rd.end > 0 && rd.start >= rd.end {
		rd.start = 0
		rd.end = 0
	} else if rd.end == len(rd.buf) {
		if rd.FixedBuffer {
			return 0, ErrBufferFull
		}

		rd.buf = append(rd.buf, make([]byte, len(rd.buf))...)
	}

	n, err := rd.ir.Read(rd.buf[rd.end:])
	if err != nil {
		return 0, err
	}
	rd.end += n

	return n, nil
}

func readNumber(buf []byte) (consumed int, size int, err error) {
	sizeData, err := tillCRLF(buf)
	if err != nil {
		return 0, 0, err
	}

	size, err = toInt(bytes.TrimRight(sizeData[1:], "\r\n"))
	if err != nil {
		return 0, 0, err
	}

	return len(sizeData), size, nil
}

func readInline(buf []byte) (int, *Array, error) {
	data, err := tillCRLF(buf)
	if err != nil {
		return 0, nil, err
	}

	return len(data), &Array{
		Items: []Value{
			&BulkStr{
				Value: bytes.TrimRight(data, "\r\n"),
			},
		},
	}, nil
}

func readBulkStr(buf []byte, isServer bool) (int, []byte, error) {
	offset, size, err := readNumber(buf)
	if err != nil {
		if isServer && (err == errInvalidNumber || err == errNoNumber) {
			return 0, nil, errors.New("Protocol error: invalid bulk length")
		}

		return 0, nil, err
	}

	if size < 0 {
		if isServer {
			return 0, nil, errors.New("Protocol error: invalid bulk length")
		}

		// -1 (negative size) means a null bulk string
		// Refer https://redis.io/topics/protocol#resp-bulk-strings
		return offset, nil, nil
	}

	if offset >= len(buf) || offset+size >= len(buf) {
		return 0, nil, errIncompleteData
	}

	data := buf[offset : offset+size]
	return offset + len(data) + 2, data, nil
}

func tillCRLF(buf []byte) ([]byte, error) {
	i := 0
	L := len(buf)
	found := false
	for i < len(buf) {
		if buf[i] == '\r' {
			if i+1 < L && buf[i+1] != 0 {
				i++
			}
			found = true
			break
		}

		i++
	}

	if !found {
		return nil, errIncompleteData
	}

	return buf[0 : i+1], nil
}

func toInt(data []byte) (int, error) {
	var d, sign int
	L := len(data)

	if L == 0 {
		return 0, errNoNumber
	}

	for i, b := range data {
		if i == 0 {
			if b == '-' {
				sign = -1
				continue
			}

			sign = 1
		}

		if b < '0' || b > '9' {
			return 0, errInvalidNumber
		}

		if b == '0' {
			continue
		}

		pos := int(math.Pow(10, float64(L-i-1)))
		d += int(b-'0') * pos
	}

	return sign * d, nil
}

var (
	errInvalidNumber  = errors.New("invalid number format")
	errNoNumber       = errors.New("no number")
	errIncompleteData = errors.New("error incomplete data")
)
