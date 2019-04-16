package radio_test

import (
	"bytes"
	"testing"

	"github.com/spy16/radio"
	"github.com/stretchr/testify/assert"
)

func TestWriter_Write(t *testing.T) {
	b := &bytes.Buffer{}
	wr := radio.NewWriter(b)
	wr.Write(radio.SimpleStr("hello"))

	assert.Equal(t, "+hello\r\n", b.String())
}
