package radio

import (
	"fmt"
	"strconv"
	"strings"
)

// Value represents the RESP protocol values.
type Value interface {
	// Serialize should return the RESP serialized representation of
	// the value.
	Serialize() string
}

// Nullable represents a null support value in RESP.
type Nullable interface {
	IsNil() bool
}

// SimpleStr represents a simple string in RESP.
// Refer https://redis.io/topics/protocol#resp-simple-strings
type SimpleStr string

// Serialize returns RESP representation of simple string.
func (ss SimpleStr) Serialize() string {
	return fmt.Sprintf("+%s\r\n", string(ss))
}

// ErrorStr represents a error string in RESP.
// Refer https://redis.io/topics/protocol#resp-errors
type ErrorStr string

// Serialize returns RESP representation of ErrorStr.
func (es ErrorStr) Serialize() string {
	return fmt.Sprintf("-%s\r\n", string(es))
}

// Integer represents RESP integer value.
// Refer https://redis.io/topics/protocol#resp-integers
type Integer int

// Serialize returns RESP representation of Integer.
func (in Integer) Serialize() string {
	return fmt.Sprintf(":%d\r\n", in)
}

func (in Integer) String() string {
	return strconv.Itoa(int(in))
}

// BulkStr represents a binary safe string in RESP.
// Refer https://redis.io/topics/protocol#resp-bulk-strings
type BulkStr struct {
	Value []byte
}

// Serialize returns RESP representation of the Bulk String.
func (bs *BulkStr) Serialize() string {
	if bs.Value == nil {
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(bs.Value), bs.Value)
}

// IsNil returns true if the value is a Null bulk string as per
// RESP protocol specification.
func (bs *BulkStr) IsNil() bool {
	return bs.Value == nil
}

func (bs *BulkStr) String() string {
	return string(bs.Value)
}

// Array represents Array RESP type.
// Refer https://redis.io/topics/protocol#resp-arrays
type Array struct {
	Items []Value
}

// Serialize returns RESP representation of the Array.
func (arr *Array) Serialize() string {
	if arr.Items == nil {
		return fmt.Sprintf("*-1\r\n")
	}

	s := fmt.Sprintf("*%d\r\n", len(arr.Items))
	for _, val := range arr.Items {
		s += val.Serialize()
	}

	return s
}

// IsNil returns true if the underlying items slice is nil.
func (arr *Array) IsNil() bool {
	return arr.Items == nil
}

func (arr *Array) String() string {
	strs := []string{}
	for _, itm := range arr.Items {
		strs = append(strs, fmt.Sprintf("%s", itm))
	}

	return strings.Join(strs, "\n")
}
