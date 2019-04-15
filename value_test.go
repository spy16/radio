package radio_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spy16/radio"
	"github.com/stretchr/testify/assert"
)

func TestSerialize(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		val  radio.Value
		resp string
		str  string
	}{
		{
			val:  radio.SimpleStr("hello"),
			resp: "+hello\r\n",
			str:  "hello",
		},
		{
			val:  radio.Integer(10),
			resp: ":10\r\n",
			str:  "10",
		},
		{
			val:  radio.ErrorStr("failed"),
			resp: "-failed\r\n",
			str:  "failed",
		},
		{
			val: &radio.BulkStr{
				Value: nil,
			},
			resp: "$-1\r\n",
			str:  "",
		},
		{
			val: &radio.BulkStr{
				Value: []byte(""),
			},
			resp: "$0\r\n\r\n",
			str:  "",
		},
		{
			val: &radio.BulkStr{
				Value: []byte("hello"),
			},
			resp: "$5\r\nhello\r\n",
			str:  "hello",
		},
		{
			val: &radio.Array{
				Items: nil,
			},
			resp: "*-1\r\n",
			str:  "",
		},
		{
			val: &radio.Array{
				Items: []radio.Value{
					radio.SimpleStr("hello"),
					radio.Integer(10),
				},
			},
			resp: "*2\r\n+hello\r\n:10\r\n",
			str:  "hello\n10",
		},
		{
			val: &radio.Array{
				Items: []radio.Value{},
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
		bs := &radio.BulkStr{
			Value: []byte(""),
		}

		assert.False(t, bs.IsNil())
	})

	suite.Run("WithValue", func(t *testing.T) {
		bs := &radio.BulkStr{
			Value: []byte("helllo"),
		}

		assert.False(t, bs.IsNil())
	})

	suite.Run("WhenNil", func(t *testing.T) {
		bs := &radio.BulkStr{
			Value: nil,
		}

		assert.True(t, bs.IsNil())
	})
}

func TestArray_IsNil(suite *testing.T) {
	suite.Parallel()

	suite.Run("WhenEmpty", func(t *testing.T) {
		arr := &radio.Array{
			Items: []radio.Value{},
		}

		assert.False(t, arr.IsNil())
	})

	suite.Run("WithValue", func(t *testing.T) {
		arr := &radio.Array{
			Items: []radio.Value{
				radio.SimpleStr("hello"),
			},
		}

		assert.False(t, arr.IsNil())
	})

	suite.Run("WhenNil", func(t *testing.T) {
		arr := &radio.Array{
			Items: nil,
		}

		assert.True(t, arr.IsNil())
	})
}
