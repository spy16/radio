package radio_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spy16/radio"
	"github.com/stretchr/testify/assert"
)

func TestOutputReader_Read(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		data string
		val  radio.Value
		err  error
	}{
		{
			data: "",
			val:  nil,
			err:  io.EOF,
		},
		{
			data: "+hello\r\n",
			val:  radio.SimpleStr("hello"),
			err:  nil,
		},
		{
			data: "+",
			val:  radio.SimpleStr(""),
			err:  io.EOF,
		},
		{
			data: "+\r\n",
			val:  radio.SimpleStr(""),
			err:  nil,
		},
		{
			data: "+hello",
			val:  radio.SimpleStr("hello"),
			err:  io.EOF,
		},
		{
			data: "-",
			val:  radio.ErrorStr(""),
			err:  io.EOF,
		},
		{
			data: "-\r\n",
			val:  radio.ErrorStr(""),
			err:  nil,
		},
		{
			data: "-ERR failed",
			val:  radio.ErrorStr("ERR failed"),
			err:  io.EOF,
		},
		{
			data: "-ERR failed\r\n",
			val:  radio.ErrorStr("ERR failed"),
			err:  nil,
		},
		{
			data: ":",
			val:  radio.Integer(0),
			err:  io.EOF,
		},
		{
			data: ":10",
			val:  radio.Integer(0),
			err:  io.EOF,
		},
		{
			data: ":10\r\n",
			val:  radio.Integer(10),
			err:  nil,
		},
		{
			data: ":\r\n",
			val:  radio.Integer(0),
			err:  errors.New("no number"),
		},
		{
			data: "$5\r\n",
			val:  nil,
			err:  io.EOF,
		},
		{
			data: "$5\r\nhello",
			val: &radio.BulkStr{
				Value: []byte("hello"),
			},
			err: nil,
		},
		{
			data: "$5\r\nhel\r\n",
			val: &radio.BulkStr{
				Value: []byte("hel\r\n"),
			},
			err: nil,
		},
		{
			data: "$1.5\r\nssd\r\n",
			val:  nil,
			err:  errors.New("invalid number format"),
		},
		{
			data: "$-1\r\n",
			val:  &radio.BulkStr{},
			err:  nil,
		},
		{
			data: "$0\r\n",
			val: &radio.BulkStr{
				Value: []byte{},
			},
			err: nil,
		},
		{
			data: "*",
			val:  nil,
			err:  io.EOF,
		},
		{
			data: "*-1\r\n",
			val:  &radio.Array{},
			err:  nil,
		},

		{
			data: "*0\r\n",
			val: &radio.Array{
				Items: []radio.Value{},
			},
			err: nil,
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
		{
			data: "*2\r\n+hello\r\n-ERR failed\r\n",
			val: &radio.Array{
				Items: []radio.Value{
					radio.SimpleStr("hello"),
					radio.ErrorStr("ERR failed"),
				},
			},
		},
		{
			data: "*2\r\n*1\r\n+hello\r\n-ERR failed\r\n",
			val: &radio.Array{
				Items: []radio.Value{
					&radio.Array{
						Items: []radio.Value{
							radio.SimpleStr("hello"),
						},
					},
					radio.ErrorStr("ERR failed"),
				},
			},
		},
		{
			data: "*2\r\n*1\r\n+hello\r\n-ERR failed",
			val:  nil,
			err:  io.EOF,
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
	par := radio.NewOutputReader(strings.NewReader(s))
	val, err := par.Read()
	fx(val, err)
}
