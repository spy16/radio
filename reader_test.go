package radio_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spy16/radio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader_Read(suite *testing.T) {
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
			rd := radio.NewReader(bytes.NewBuffer([]byte(cs.input)))

			mb, err := rd.Read()
			if cs.err {
				assert.Error(t, err)
				assert.Nil(t, mb)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, mb)
				assert.Equal(t, cs.mb, *mb)
			}
		})
	}
}
