package radio_test

import (
	"bytes"
	"testing"

	"github.com/spy16/radio"
)

func TestWriter_Write(t *testing.T) {
	b := &bytes.Buffer{}
	wr := radio.NewWriter(b)
	wr.Write(radio.SimpleStr("hello"))

	expected := "+hello\r\n"

	if actual := b.String(); actual != expected {
		t.Errorf("expecting '%s', got '%s'", expected, actual)
	}
}
