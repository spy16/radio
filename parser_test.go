package radio_test

import (
	"io"
	"strings"
	"testing"

	"github.com/spy16/radio"
	"github.com/stretchr/testify/assert"
)

func TestParser_Next(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		data string
		val  radio.Value
		err  error
	}{
		{
			data: "@hello\r\n",
			val:  radio.InlineStr("@hello"),
			err:  nil,
		},
		{
			data: "+hello\r\n",
			val:  radio.SimpleStr("hello"),
			err:  nil,
		},
		{
			data: "+hello\r\n+should-be-ignored\r\n",
			val:  radio.SimpleStr("hello"),
			err:  nil,
		},
		{
			data: "+\r\n",
			val:  radio.SimpleStr(""),
			err:  nil,
		},
		{
			data: "+",
			val:  radio.SimpleStr(""),
			err:  nil,
		},
		{
			data: "-ERR hello\r\n",
			val:  radio.ErrorStr("ERR hello"),
			err:  nil,
		},
		{
			data: "$2\r\n",
			val:  nil,
			err:  io.EOF,
		},
		{
			data: "$5\r\nhello\r\n",
			val: &radio.BulkStr{
				Value: []byte("hello"),
			},
			err: nil,
		},
		{
			data: "$-1\r\n",
			val:  &radio.BulkStr{},
			err:  nil,
		},
		{
			data: "$1.4\r\n",
			val:  nil,
			err: &radio.ProtocolError{
				Reason: "invalid bulk length",
			},
		},
		{
			data: "$3\r\nhello\r\n",
			val:  nil,
			err: &radio.ProtocolError{
				Reason: "required 5 bytes for bulk-string, got 3",
			},
		},
		{
			data: ":100\r\n",
			val:  radio.Integer(100),
			err:  nil,
		},
		{
			data: ":100.2\r\n",
			val:  nil,
			err: &radio.ProtocolError{
				Reason: "invalid integer format",
			},
		},
		{
			data: ":100",
			val:  radio.Integer(100),
			err:  nil,
		},
		{
			data: "*-1\r\n",
			val: &radio.Array{
				Items: nil,
			},
			err: nil,
		},
		{
			data: "*0\r\n",
			val: &radio.Array{
				Items: []radio.Value{},
			},
			err: nil,
		},
		{
			data: "*1\r\n",
			val:  nil,
			err:  io.EOF,
		},
		{
			data: "*1.4\r\n",
			val:  nil,
			err: &radio.ProtocolError{
				Reason: "invalid array length",
			},
		},
		{
			data: "*1\r\n+hello\r\n",
			val: &radio.Array{
				Items: []radio.Value{
					radio.SimpleStr("hello"),
				},
			},
			err: nil,
		},
	}

	for _, cs := range cases {
		suite.Run(cs.data, func(t *testing.T) {
			parse(cs.data, func(val radio.Value, err error) {
				assert.Equal(t, cs.val, val)
				assert.Equal(t, cs.err, err)
			})
		})
	}
}

func parse(s string, fx func(val radio.Value, err error)) {
	par := radio.NewParser(strings.NewReader(s), false)
	val, err := par.Next()
	fx(val, err)
}
