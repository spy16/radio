package resp_test

import (
	"io"
	"strings"
	"testing"

	"github.com/spy16/radio/resp"
	"github.com/stretchr/testify/assert"
)

func TestParser_Next(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		data string
		val  resp.Value
		err  error
	}{
		{
			data: "@hello\r\n",
			val:  resp.InlineStr("@hello"),
			err:  nil,
		},
		{
			data: "+hello\r\n",
			val:  resp.SimpleStr("hello"),
			err:  nil,
		},
		{
			data: "+hello\r\n+should-be-ignored\r\n",
			val:  resp.SimpleStr("hello"),
			err:  nil,
		},
		{
			data: "+\r\n",
			val:  resp.SimpleStr(""),
			err:  nil,
		},
		{
			data: "+",
			val:  resp.SimpleStr(""),
			err:  nil,
		},
		{
			data: "-ERR hello\r\n",
			val:  resp.ErrorStr("ERR hello"),
			err:  nil,
		},
		{
			data: "$2\r\n",
			val:  nil,
			err:  io.EOF,
		},
		{
			data: "$5\r\nhello\r\n",
			val: &resp.BulkStr{
				Value: []byte("hello"),
			},
			err: nil,
		},
		{
			data: "$-1\r\n",
			val:  &resp.BulkStr{},
			err:  nil,
		},
		{
			data: "$1.4\r\n",
			val:  nil,
			err:  resp.ErrNumberFormat,
		},
		{
			data: "$3\r\nhello\r\n",
			val:  nil,
			err:  resp.ErrProtocol,
		},
		{
			data: ":100\r\n",
			val:  resp.Integer(100),
			err:  nil,
		},
		{
			data: ":100.2\r\n",
			val:  nil,
			err:  resp.ErrNumberFormat,
		},
		{
			data: ":100",
			val:  resp.Integer(100),
			err:  nil,
		},
		{
			data: "*-1\r\n",
			val: &resp.Array{
				Items: nil,
			},
			err: nil,
		},
		{
			data: "*0\r\n",
			val: &resp.Array{
				Items: []resp.Value{},
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
			err:  resp.ErrNumberFormat,
		},
		{
			data: "*1\r\n+hello\r\n",
			val: &resp.Array{
				Items: []resp.Value{
					resp.SimpleStr("hello"),
				},
			},
			err: nil,
		},
	}

	for _, cs := range cases {
		suite.Run(cs.data, func(t *testing.T) {
			parse(cs.data, func(val resp.Value, err error) {
				assert.Equal(t, cs.val, val)
				assert.Equal(t, cs.err, err)
			})
		})
	}
}

func parse(s string, fx func(val resp.Value, err error)) {
	par := resp.New(strings.NewReader(s))
	val, err := par.Next()
	fx(val, err)
}
