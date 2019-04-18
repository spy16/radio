package radio_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/spy16/radio"
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
			actualResp := cs.val.Serialize()
			if cs.resp != actualResp {
				t.Errorf("expecting serialization '%s', got '%s'", cs.resp, actualResp)
			}

			var actualStr string
			if stringer, ok := cs.val.(fmt.Stringer); ok {
				actualStr = stringer.String()
			} else {
				actualStr = fmt.Sprintf("%s", cs.val)
			}

			if cs.str != actualStr {
				t.Errorf("expecting string '%s', got '%s'", cs.str, actualStr)
			}
		})
	}
}

func TestValue_IsNil(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		title       string
		val         radio.Value
		shouldBeNil bool
	}{
		{
			title: "BulkStr-WhenEmpty",
			val: &radio.BulkStr{
				Value: []byte(""),
			},
			shouldBeNil: false,
		},
		{
			title: "BulkStr-WithValue",
			val: &radio.BulkStr{
				Value: []byte("hello"),
			},
			shouldBeNil: false,
		},
		{
			title: "BulkStr-WhenNil",
			val: &radio.BulkStr{
				Value: nil,
			},
			shouldBeNil: true,
		},
		{
			title: "Array-WhenEmpty",
			val: &radio.Array{
				Items: []radio.Value{},
			},
			shouldBeNil: false,
		},
		{
			title: "Array-WithValue",
			val: &radio.Array{
				Items: []radio.Value{
					radio.SimpleStr("hello"),
				},
			},
			shouldBeNil: false,
		},
		{
			title: "Array-WhenNil",
			val: &radio.Array{
				Items: nil,
			},
			shouldBeNil: true,
		},
	}

	for _, cs := range cases {
		suite.Run(cs.title, func(t *testing.T) {
			nillable, ok := cs.val.(interface{ IsNil() bool })
			if ok {
				if nillable.IsNil() != cs.shouldBeNil {
					t.Errorf("expecting '%t', got '%t'", cs.shouldBeNil, !cs.shouldBeNil)
				}
			} else {
				t.Logf("type %s is not nillable", reflect.TypeOf(cs.val))
			}
		})
	}
}
