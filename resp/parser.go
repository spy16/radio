package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

var (
	// ErrProtocol is returned when the data stream is not according to
	// the RESP spec.
	ErrProtocol = errors.New("failed to read")

	// ErrNumberFormat is returned when parsing a number fails.
	ErrNumberFormat = errors.New("invalid integer format")
)

// New initializes an instance of parser with given reader. Reader will be
// buffered using 'bufio.Scanner'.
func New(rdr io.Reader) *Parser {
	return &Parser{
		sc: bufio.NewScanner(rdr),
		consumers: map[byte]consumeFunc{
			'+': consumeSimpleStr,
			'-': consumeErrorStr,
			':': consumeInteger,
			'$': consumeBulkStr,
			'*': consumeArray,
		},
	}
}

// Parser parser a given reader and emits RESP values.
type Parser struct {
	inline    bool
	sc        *bufio.Scanner
	consumers map[byte]consumeFunc
}

// consumeFunc is responsible for parsing curLine while consuming more
// tokens from Parser if required and returning a Value.
type consumeFunc func(par *Parser, curLine string) (Value, error)

// Next reads next set of bytes from the stream and emits the Value.
func (par *Parser) Next() (Value, error) {
	if !par.sc.Scan() {
		err := par.sc.Err()
		if err == nil {
			return nil, io.EOF
		}
		return nil, err

	}

	line := par.sc.Text()
	prefix := line[0]

	consume, found := par.consumers[prefix]
	if !found {
		return InlineStr(line), nil
	}

	return consume(par, line)
}

func consumeSimpleStr(_ *Parser, line string) (Value, error) {
	return SimpleStr(line[1:]), nil
}

func consumeErrorStr(_ *Parser, line string) (Value, error) {
	return ErrorStr(line[1:]), nil
}

func consumeInteger(_ *Parser, line string) (Value, error) {
	val, err := getNum(line)
	if err != nil {
		return nil, err
	}

	return Integer(val), nil
}

func consumeBulkStr(par *Parser, line string) (Value, error) {
	size, err := getNum(line)
	if err != nil {
		return nil, err
	}

	if size == -1 {
		return &BulkStr{}, nil
	}

	read := ""
	for par.sc.Scan() {
		read += par.sc.Text()
		if len(read) >= size {
			break
		}
	}

	if len(read) > size {
		return nil, ErrProtocol
	}

	if len(read) < size {
		err := par.sc.Err()
		if err == nil {
			return nil, io.EOF
		}

		return nil, err
	}

	return &BulkStr{
		Value: []byte(read),
	}, nil
}

func consumeArray(par *Parser, line string) (Value, error) {
	size, err := getNum(line)
	if err != nil {
		return nil, err
	}

	if size == -1 {
		return &Array{
			Items: nil,
		}, nil
	}

	arr := &Array{
		Items: []Value{},
	}
	for i := 0; i < size; i++ {
		val, err := par.Next()
		if err != nil {
			return nil, err
		}

		arr.Items = append(arr.Items, val)
	}

	return arr, nil
}

func getNum(line string) (int, error) {
	val, err := strconv.Atoi(line[1:])
	if err != nil {
		return 0, ErrNumberFormat
	}
	return val, nil
}
