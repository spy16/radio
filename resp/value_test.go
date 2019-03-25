package resp_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spy16/radio/resp"
	"github.com/stretchr/testify/assert"
)

func TestSerialize(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		val  resp.Value
		resp string
		str  string
	}{
		{
			val:  resp.SimpleStr("hello"),
			resp: "+hello\r\n",
			str:  "hello",
		},
		{
			val:  resp.InlineStr("hello"),
			resp: "+hello\r\n",
			str:  "hello",
		},
		{
			val:  resp.Integer(10),
			resp: ":10\r\n",
			str:  "10",
		},
		{
			val:  resp.ErrorStr("failed"),
			resp: "-failed\r\n",
			str:  "failed",
		},
		{
			val: &resp.BulkStr{
				Value: nil,
			},
			resp: "$-1\r\n",
			str:  "",
		},
		{
			val: &resp.BulkStr{
				Value: []byte(""),
			},
			resp: "$0\r\n\r\n",
			str:  "",
		},
		{
			val: &resp.BulkStr{
				Value: []byte("hello"),
			},
			resp: "$5\r\nhello\r\n",
			str:  "hello",
		},
		{
			val: &resp.Array{
				Items: nil,
			},
			resp: "*-1\r\n",
			str:  "",
		},
		{
			val: &resp.Array{
				Items: []resp.Value{
					resp.SimpleStr("hello"),
					resp.Integer(10),
				},
			},
			resp: "*2\r\n+hello\r\n:10\r\n",
			str:  "hello\n10",
		},
		{
			val: &resp.Array{
				Items: []resp.Value{},
			},
			resp: "*0\r\n",
			str:  "",
		},
	}

	for _, cs := range cases {
		suite.Run(reflect.TypeOf(cs.val).String(), func(t *testing.T) {
			assert.Equal(t, cs.resp, cs.val.Serialize())

			var str string
			if stringer, ok := cs.val.(fmt.Stringer); ok {
				str = stringer.String()
			} else {
				str = fmt.Sprintf("%s", cs.val)
			}
			assert.Equal(t, cs.str, str)
		})
	}
}

func TestBulkStr_IsNil(suite *testing.T) {
	suite.Parallel()

	suite.Run("WhenEmpty", func(t *testing.T) {
		bs := &resp.BulkStr{
			Value: []byte(""),
		}

		assert.False(t, bs.IsNil())
	})

	suite.Run("WithValue", func(t *testing.T) {
		bs := &resp.BulkStr{
			Value: []byte("helllo"),
		}

		assert.False(t, bs.IsNil())
	})

	suite.Run("WhenNil", func(t *testing.T) {
		bs := &resp.BulkStr{
			Value: nil,
		}

		assert.True(t, bs.IsNil())
	})
}

func TestArray_IsNil(suite *testing.T) {
	suite.Parallel()

	suite.Run("WhenEmpty", func(t *testing.T) {
		arr := &resp.Array{
			Items: []resp.Value{},
		}

		assert.False(t, arr.IsNil())
	})

	suite.Run("WithValue", func(t *testing.T) {
		arr := &resp.Array{
			Items: []resp.Value{
				resp.SimpleStr("hello"),
			},
		}

		assert.False(t, arr.IsNil())
	})

	suite.Run("WhenNil", func(t *testing.T) {
		arr := &resp.Array{
			Items: nil,
		}

		assert.True(t, arr.IsNil())
	})
}
