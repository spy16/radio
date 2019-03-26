package resp

import (
	"bufio"
	"io"
	"strconv"

	"github.com/pkg/errors"
)

// New initializes an instance of parser with given reader. Reader will be
// buffered using 'bufio.Scanner'.
func New(rdr io.Reader, inline bool) *Parser {
	return &Parser{
		inline: inline,
		sc:     bufio.NewScanner(rdr),
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

	inArray bool
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

	if par.inline && prefix != '*' && !par.inArray {
		return InlineStr(line), nil
	}

	consume, found := par.consumers[prefix]
	if !found {
		return InlineStr(line), nil
	}

	val, err := consume(par, line)
	if err != nil {
		return nil, err
	}

	if par.inline && par.inArray {
		if _, ok := val.(*BulkStr); !ok {
			return nil, errors.Wrapf(ErrProtocol, "expected '$', got '%c'", val.Serialize()[0])
		}
	}

	return val, nil
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
	par.inArray = true
	defer func() {
		par.inArray = false
	}()

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

func ensureBulkStrArray(arr *Array) error {
	for _, itm := range arr.Items {
		_, isBulkStr := itm.(*BulkStr)

		if !isBulkStr {
		}
	}

	return nil
}
