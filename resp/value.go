package resp

// Different kinds of values
const (
	SimpleStr = Kind("simpleStr")
	BulkStr   = Kind("bulkStr")
	ErrStr    = Kind("errStr")
	Integer   = Kind("integer")
	Array     = Kind("array")

	// extension kinds (not part of RESP)
	Float = Kind("float")
)

// Value represents any
type Value struct {
	kind  Kind
	val   string
	arr   []*Value
	isNil bool
}

// Array returns the value of the underlying array.
func (val *Value) Array() []*Value {
	return val.arr
}

// IsNil returns if the value is nil as per RESP spec.
func (val *Value) IsNil() bool {
	return val.isNil
}

// Kind returns the type of the underlying value.
func (val *Value) Kind() Kind {
	return val.kind
}

func (val *Value) String() string {
	return val.val
}

// Kind represents the type of the value.
type Kind string
