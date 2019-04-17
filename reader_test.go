package radio_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spy16/radio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader_Read(suite *testing.T) {
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
	par := radio.NewReader(strings.NewReader(s), false)
	val, err := par.Read()
	fx(val, err)
}

func TestReader_Read_ServerMode(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		input string
		mb    radio.MultiBulk
		err   bool
	}{
		{
			input: "",
			err:   true,
		},
		{
			input: "hello\r\n",
			err:   false,
			mb: radio.MultiBulk{
				Items: []radio.BulkStr{
					{
						Value: []byte("hello"),
					},
				},
			},
		},
		{
			input: "hello",
			err:   true,
		},
		{
			input: "*-1\r\n",
			err:   false,
			mb:    radio.MultiBulk{},
		},
		{
			input: "*abc\r\n",
			err:   true,
		},
		{
			input: "*0\r\n",
			err:   false,
			mb: radio.MultiBulk{
				Items: []radio.BulkStr{},
			},
		},
		{
			input: "*1\r\n",
			err:   true,
		},
		{
			input: "*1\r\n+hello\r\n",
			err:   true,
		},
		{
			input: "*1\r\n$-1\r\n",
			err:   false,
			mb: radio.MultiBulk{
				Items: []radio.BulkStr{
					{},
				},
			},
		},
		{
			input: "*1\r\n$5\r\n",
			err:   true,
		},
		{
			input: "*1\r\n$5\r\nhello\r\n",
			err:   false,
			mb: radio.MultiBulk{
				Items: []radio.BulkStr{
					{
						Value: []byte("hello"),
					},
				},
			},
		},
		{
			input: "*2\r\n$5\r\nhello\r\n$4\r\ncool\r\n",
			err:   false,
			mb: radio.MultiBulk{
				Items: []radio.BulkStr{
					{
						Value: []byte("hello"),
					},
					{
						Value: []byte("cool"),
					},
				},
			},
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("Case#%d", id), func(t *testing.T) {
			rd := radio.NewReader(bytes.NewBuffer([]byte(cs.input)), true)

			val, err := rd.Read()
			if cs.err {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, val)

				mb, ok := val.(*radio.MultiBulk)
				require.True(t, ok)
				assert.Equal(t, cs.mb, *mb)
			}
		})
	}
}
