package resp_test

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spy16/radio/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(suite *testing.T) {
	suite.Parallel()

	suite.Run("WithExtensions", func(t *testing.T) {
		par := resp.NewParser(strings.NewReader("?1.4\r\n"), true)

		val, err := par.Next()
		assert.NoError(t, err)
		require.NotNil(t, val)
		assert.True(t, resp.Float == val.Kind())
	})

	suite.Run("WithNoExtensions", func(t *testing.T) {
		par := resp.NewParser(strings.NewReader("?1.4\r\n"), false)

		val, err := par.Next()
		require.Error(t, err)
		assert.Nil(t, val)
	})
}

func TestParser_EmptyString(t *testing.T) {
	par := reader("")

	val, err := par.Next()
	require.Error(t, err)
	assert.Equal(t, io.EOF, err)
	assert.Nil(t, val)
}

func TestParser_InvalidProtocolPrefix(t *testing.T) {
	par := reader("@hello")

	val, err := par.Next()
	require.Error(t, err)
	assert.Contains(t, "invalid prefix", err.Error())
	assert.Nil(t, val)
}

func TestParser_SimpleStr(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		title     string
		src       string
		expectErr bool
		expectVal string
	}{
		{
			title:     "Valid",
			src:       "+hello\r\n",
			expectErr: false,
			expectVal: "hello",
		},
		{
			title:     "EmptyStr",
			src:       "+\r\n",
			expectErr: false,
			expectVal: "",
		},
		{
			title:     "MultipleCRLF",
			src:       "+hello\r\nworld\r\n",
			expectErr: false,
			expectVal: "hello",
		},
		{
			title:     "MissingTerminalCRLF",
			src:       "+hello",
			expectErr: true,
		},
		{
			title:     "MissingCR",
			src:       "+hello\n",
			expectErr: false,
			expectVal: "hello",
		},
	}

	for _, cs := range cases {
		suite.Run(cs.title, func(t *testing.T) {
			par := reader(cs.src)

			val, err := par.Next()
			if cs.expectErr {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, val)
				assert.Equal(t, resp.SimpleStr, val.Kind())
				assert.Equal(t, cs.expectVal, val.String())
			}
		})
	}
}

func TestParser_SimpleErr(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		title     string
		src       string
		expectErr bool
		expectVal string
	}{
		{
			title:     "Valid",
			src:       "-ERR hello\r\n",
			expectErr: false,
			expectVal: "ERR hello",
		},
		{
			title:     "EmptyErr",
			src:       "-\r\n",
			expectErr: false,
			expectVal: "",
		},
		{
			title:     "MultipleCRLF",
			src:       "-WRONG hello\r\nworld\r\n",
			expectErr: false,
			expectVal: "WRONG hello",
		},
	}

	for _, cs := range cases {
		suite.Run(cs.title, func(t *testing.T) {
			par := reader(cs.src)

			val, err := par.Next()
			if cs.expectErr {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, val)
				assert.Equal(t, resp.ErrStr, val.Kind())
				assert.Equal(t, cs.expectVal, val.String())
			}
		})
	}
}

func TestParser_Integer(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		title     string
		src       string
		expectErr bool
		expectVal string
	}{
		{
			title:     "Valid",
			src:       ":1000\r\n",
			expectErr: false,
			expectVal: "1000",
		},
		{
			title:     "EmptyErr",
			src:       ":18\r\n",
			expectErr: false,
			expectVal: "18",
		},
		{
			title:     "MultipleCRLF",
			src:       ":19\r\nworld\r\n",
			expectErr: false,
			expectVal: "19",
		},
	}

	for _, cs := range cases {
		suite.Run(cs.title, func(t *testing.T) {
			par := reader(cs.src)

			val, err := par.Next()
			if cs.expectErr {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, val)
				assert.True(t, resp.Integer == val.Kind(), fmt.Sprintf("not valid kind: %s", val.Kind()))
				assert.Equal(t, cs.expectVal, val.String())
			}
		})
	}
}

func TestParser_Float(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		title     string
		src       string
		expectErr bool
		expectVal string
	}{
		{
			title:     "Valid",
			src:       "?100.0\r\n",
			expectErr: false,
			expectVal: "100.0",
		},
		{
			title:     "EmptyErr",
			src:       "?1.8\r\n",
			expectErr: false,
			expectVal: "1.8",
		},
		{
			title:     "MultipleCRLF",
			src:       "?1.9\r\nworld\r\n",
			expectErr: false,
			expectVal: "1.9",
		},
	}

	for _, cs := range cases {
		suite.Run(cs.title, func(t *testing.T) {
			par := reader(cs.src)

			val, err := par.Next()
			if cs.expectErr {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, val)
				assert.True(t, resp.Float == val.Kind(), fmt.Sprintf("not valid kind: %s", val.Kind()))
				assert.Equal(t, cs.expectVal, val.String())
			}
		})
	}
}

func TestParser_BulkStr(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		title     string
		src       string
		expectErr bool
		expectVal string
		expectNil bool
	}{
		{
			title:     "Valid",
			src:       "$5\r\nhello\r\n",
			expectErr: false,
			expectVal: "hello",
		},
		{
			title:     "WithCRLFInString",
			src:       "$12\r\nhello\r\nworld\r\n",
			expectErr: false,
			expectVal: "hello\r\nworld",
		},
		{
			title:     "WithCRLFInString",
			src:       "$15\r\nhello\r\nworld\r\n",
			expectErr: true,
			expectVal: "hello\r\nworld",
		},
		{
			title:     "WithEmptyStr",
			src:       "$0\r\nhello\r\nworld\r\n",
			expectErr: false,
			expectVal: "",
		},
		{
			title:     "WithNilStr",
			src:       "$-1\r\nhello\r\nworld\r\n",
			expectErr: false,
			expectVal: "",
			expectNil: true,
		},
		{
			title:     "WithEOF",
			src:       "$10",
			expectErr: true,
			expectVal: "",
		},
		{
			title:     "InvalidSize",
			src:       "$1a\r\n",
			expectErr: true,
			expectVal: "",
		},
	}

	for _, cs := range cases {
		suite.Run(cs.title, func(t *testing.T) {
			par := reader(cs.src)

			val, err := par.Next()
			if cs.expectErr {
				assert.Error(t, err)
				assert.Nil(t, val)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, val)
				assert.True(t, resp.BulkStr == val.Kind(), fmt.Sprintf("not valid kind: %s", val.Kind()))
				assert.Equal(t, cs.expectVal, val.String())
				assert.Equal(t, cs.expectNil, val.IsNil())
			}
		})
	}
}

func TestParser_Array(suite *testing.T) {
	suite.Parallel()

	suite.Run("InvalidStr", func(t *testing.T) {
		par := reader("*1.9")

		val, err := par.Next()
		require.Error(t, err)
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, val)
	})

	suite.Run("InvalidSize", func(t *testing.T) {
		par := reader("*1.9\r\n+hello\r\n")

		val, err := par.Next()
		assert.Error(t, err)
		assert.Nil(t, val)
	})

	suite.Run("InvalidItem", func(t *testing.T) {
		par := reader("*1\r\n$1\r\n")

		val, err := par.Next()
		require.Error(t, err)
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, val)
	})

	suite.Run("ValidArr", func(t *testing.T) {
		par := reader("*1\r\n+hello\r\n")

		val, err := par.Next()
		assert.NoError(t, err)
		require.NotNil(t, val)
		assert.True(t, resp.Array == val.Kind())
		assert.False(t, val.IsNil())

		arr := val.Array()
		require.NotNil(t, arr)
		require.Equal(t, 1, len(arr))

		arr0 := arr[0]
		require.NotNil(t, arr0)
		assert.True(t, resp.SimpleStr == arr0.Kind())
	})

	suite.Run("NestedArray", func(t *testing.T) {
		par := reader("*2\r\n+hello\r\n*2\r\n+hello\r\n:100\r\n")

		val, err := par.Next()
		assert.NoError(t, err)
		require.NotNil(t, val)
		assert.True(t, resp.Array == val.Kind())
		assert.False(t, val.IsNil())

		arr := val.Array()
		require.NotNil(t, arr)
		require.Equal(t, 2, len(arr))

		arr0 := arr[0]
		require.NotNil(t, arr0)
		assert.True(t, resp.SimpleStr == arr0.Kind())

		arr1 := arr[1]
		require.NotNil(t, arr1)
		assert.True(t, resp.Array == arr1.Kind())
		assert.Equal(t, 2, len(arr1.Array()))
	})
}

func reader(s string) *resp.Parser {
	return resp.NewParser(strings.NewReader(s), true)
}
