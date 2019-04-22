package radio_test

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/spy16/radio"
)

func TestNewReader(suite *testing.T) {
	suite.Parallel()

	suite.Run("WithServerMode", func(t *testing.T) {
		rd := radio.NewReader(bytes.NewBufferString("hello"), true)
		if rd == nil {
			t.Errorf("return value must not be nil")
		}
		if !rd.IsServer {
			t.Errorf("reader expected to be in server mode, but in client mode")
		}
	})

	suite.Run("WithClientMode", func(t *testing.T) {
		rd := radio.NewReader(bytes.NewBufferString("hello"), false)
		if rd == nil {
			t.Errorf("return value must not be nil")
		}
		if rd.IsServer {
			t.Errorf("reader expected to be in client mode, but in server mode")
		}
	})

	suite.Run("WithSize", func(t *testing.T) {
		rd := radio.NewReaderSize(bytes.NewBufferString("+hello\r\n"), false, 2)
		if rd == nil {
			t.Errorf("return value must not be nil")
		}
		if rd.IsServer {
			t.Errorf("reader expected to be in client mode, but in server mode")
		}

		v, err := rd.Read()
		if err != nil {
			t.Errorf("not expecting error, got '%v'", err)
		}

		if v == nil {
			t.Errorf("expecting non-nil value from reader, got nil")
		}

		minSz, curSz := rd.Size()
		if minSz != 2 {
			t.Errorf("expected minimum buffer size to be 2, got %d", minSz)
		}

		if curSz != 8 {
			t.Errorf("expected current buffer size to be 8, got %d", curSz)
		}
	})

	suite.Run("WithFixedSize", func(t *testing.T) {
		rd := radio.NewReaderSize(bytes.NewBufferString("*hello"), false, 2)
		if rd == nil {
			t.Errorf("return value must not be nil")
		}
		rd.FixedBuffer = true

		v, err := rd.Read()
		if err != radio.ErrBufferFull {
			t.Errorf("expecting error '%v', got '%v'", radio.ErrBufferFull, err)
		}

		if v != nil {
			t.Errorf("expecting nil value from reader, got '%v'", v)
		}
	})
}

func TestReader_Read_ClientMode(suite *testing.T) {
	suite.Parallel()

	cases := []readTestCase{
		{
			title: "NoInput",
			input: "",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "BadPrefix",
			input: "?helo",
			val:   nil,
			err:   errors.New("bad prefix '?'"),
		},
		{
			title: "SimpleStr",
			input: "+hello\r\n",
			val:   radio.SimpleStr("hello"),
			err:   nil,
		},
		{
			title: "SimpleStr-Empty",
			input: "+\r\n",
			val:   radio.SimpleStr(""),
			err:   nil,
		},
		{
			title: "SimpleStr-NoCRLF",
			input: "+hello",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "SimpleStr-MultiValue",
			input: "+hello\r\n+world\r\n",
			val:   radio.SimpleStr("hello"),
			err:   nil,
		},
		{
			title: "ErrorStr",
			input: "-ERR failed\r\n",
			val:   radio.ErrorStr("ERR failed"),
			err:   nil,
		},
		{
			title: "ErrorStr-NoValue",
			input: "-\r\n",
			val:   radio.ErrorStr(""),
			err:   nil,
		},
		{
			title: "ErrorStr-NoCRLF",
			input: "-ERR failed",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "Integer",
			input: ":100\r\n",
			val:   radio.Integer(100),
			err:   nil,
		},
		{
			title: "Integer-NoValue",
			input: ":\r\n",
			val:   nil,
			err:   errors.New("no number"),
		},
		{
			title: "Integer-NoCRLF",
			input: ":100",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "Integer-BadFormat",
			input: ":10.5\r\n",
			val:   nil,
			err:   errors.New("invalid number format"),
		},
		{
			title: "Integer-EOF",
			input: ":",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "BulkStr",
			input: "$5\r\nhello\r\n",
			val: &radio.BulkStr{
				Value: []byte("hello"),
			},
			err: nil,
		},
		{
			title: "BulkStr-NoSize",
			input: "$\r\n",
			val:   nil,
			err:   errors.New("no number"),
		},
		{
			title: "BulkStr-NegativeSize",
			input: "$-1\r\n",
			val:   &radio.BulkStr{},
			err:   nil,
		},
		{
			title: "BulkStr-NoData",
			input: "$10\r\nhel\r\n",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "Array",
			input: "*1\r\n+hello\r\n",
			val: &radio.Array{
				Items: []radio.Value{
					radio.SimpleStr("hello"),
				},
			},
			err: nil,
		},
		{
			title: "Array-NoSize",
			input: "*\r\n",
			val:   nil,
			err:   errors.New("no number"),
		},
		{
			title: "Array-NegativeSize",
			input: "*-1\r\n",
			val:   &radio.Array{},
			err:   nil,
		},
		{
			title: "Array-InsufficientData",
			input: "*2\r\n+hello\r\n",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "Array-InvalidSize",
			input: "*2.5\r\n",
			val:   nil,
			err:   errors.New("invalid number format"),
		},
	}

	runAllCases(suite, cases, false)
}

func TestReader_Read_ServerMode(suite *testing.T) {
	suite.Parallel()

	cases := []readTestCase{
		{
			title: "InlineStr",
			input: "hello\r\n",
			val: &radio.Array{
				Items: []radio.Value{
					&radio.BulkStr{
						Value: []byte("hello"),
					},
				},
			},
			err: nil,
		},
		{
			title: "InlineStr-EOF",
			input: "hello",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "MultiBulk-NullValue",
			input: "*-1\r\n",
			val:   &radio.Array{},
			err:   nil,
		},
		{
			title: "MultiBulk-EOF",
			input: "*1\r\n",
			val:   nil,
			err:   io.EOF,
		},
		{
			title: "SimpleStrInArray",
			input: "*1\r\n+hello\r\n",
			val:   nil,
			err:   errors.New("Protocol error: expected '$', got '+'"),
		},
		{
			title: "MultiBulk-InvalidSize",
			input: "*1.4\r\n",
			val:   nil,
			err:   errors.New("Protocol error: invalid multibulk length"),
		},
		{
			title: "MultiBulk-InvalidBulkSize",
			input: "*1\r\n$1.5\r\n",
			val:   nil,
			err:   errors.New("Protocol error: invalid bulk length"),
		},
		{
			title: "MultiBulk-NegativeBulkSize",
			input: "*1\r\n$-1\r\n",
			val:   nil,
			err:   errors.New("Protocol error: invalid bulk length"),
		},
	}

	runAllCases(suite, cases, true)
}

func runAllCases(suite *testing.T, cases []readTestCase, serverMode bool) {
	for _, cs := range cases {
		if cs.title == "" {
			cs.title = cs.input
		}

		suite.Run(cs.title, func(t *testing.T) {
			par := radio.NewReader(strings.NewReader(cs.input), serverMode)
			val, err := par.Read()

			if !reflect.DeepEqual(cs.err, err) {
				t.Errorf("expecting error '%v', got '%v'", cs.err, err)
			}

			if !reflect.DeepEqual(cs.val, val) {
				t.Errorf("expecting RESP value '%s{%v}', got '%s{%v}'",
					reflect.TypeOf(cs.val), cs.val, reflect.TypeOf(val), val)
			}
		})
	}

}

type readTestCase struct {
	title string
	input string
	val   radio.Value
	err   error
}
