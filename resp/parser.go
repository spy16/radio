package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

// NewParser initializes an instance of parser with given reader.
func NewParser(rdr io.Reader, extensions bool) *Parser {
	par := &Parser{}
	par.rdr = bufio.NewReader(rdr)
	par.consumers = map[byte]consumerFunc{
		'+': par.consumeSimpleStr,
		'-': par.consumeErrStr,
		':': par.consumeInteger,
		'$': par.consumeBulkStr,
		'*': par.consumeArr,
	}

	if extensions {
		par.consumers['?'] = par.consumeFloat
	}

	return par
}

// Parser parses RESP protocol stream from a Reader instance.
type Parser struct {
	rdr       *bufio.Reader
	consumers map[byte]consumerFunc
}

// Next consumes and parses next set of bytes.
func (par *Parser) Next() (*Value, error) {
	prefix, err := par.rdr.ReadByte()
	if err != nil {
		return nil, err
	}

	consume, found := par.consumers[prefix]
	if !found {
		return nil, errors.New("invalid prefix")
	}

	return consume()
}

func (par *Parser) consumeArr() (*Value, error) {
	s, err := readLine(par.rdr)
	if err != nil {
		return nil, err
	}

	sz, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}

	arr := []*Value{}
	for i := 0; i < sz; i++ {
		val, err := par.Next()
		if err != nil {
			return nil, err
		}

		arr = append(arr, val)
	}

	return &Value{
		kind: Array,
		arr:  arr,
	}, nil
}

func (par *Parser) consumeBulkStr() (*Value, error) {
	s, err := readLine(par.rdr)
	if err != nil {
		return nil, err
	}

	sz, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}

	if sz == 0 {
		readLine(par.rdr) // skip \r\n
		return &Value{
			kind: BulkStr,
			val:  "",
		}, nil
	} else if sz == -1 {
		readLine(par.rdr)
		return &Value{
			kind:  BulkStr,
			val:   "",
			isNil: true,
		}, nil
	}

	dat := make([]byte, sz)
	n, err := par.rdr.Read(dat)
	if err != nil {
		return nil, err
	}

	if n < sz {
		return nil, errors.New("read finished prematurely")
	}
	readLine(par.rdr)

	return &Value{
		kind: BulkStr,
		val:  string(dat),
	}, nil
}

func (par *Parser) consumeSimpleStr() (*Value, error) {
	return readOneLiner(par.rdr, SimpleStr)
}

func (par *Parser) consumeErrStr() (*Value, error) {
	return readOneLiner(par.rdr, ErrStr)
}

func (par *Parser) consumeInteger() (*Value, error) {
	return readOneLiner(par.rdr, Integer)
}

func (par *Parser) consumeFloat() (*Value, error) {
	return readOneLiner(par.rdr, Float)
}

func readOneLiner(rdr *bufio.Reader, kind Kind) (*Value, error) {
	s, err := readLine(rdr)
	if err != nil {
		return nil, err
	}

	return &Value{
		kind: kind,
		val:  s,
	}, nil
}

func readLine(rdr *bufio.Reader) (string, error) {
	dat, err := rdr.ReadBytes('\n')
	if err != nil {
		return "", err
	}

	dat = dropCR(dat[:len(dat)-1])
	return string(dat), nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

type consumerFunc func() (*Value, error)
