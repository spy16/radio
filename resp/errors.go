package resp

import "errors"

var (
	// ErrProtocol is returned when the data stream is not according to
	// the RESP spec.
	ErrProtocol = errors.New("Protocol Error")

	// ErrNumberFormat is returned when parsing a number fails.
	ErrNumberFormat = errors.New("invalid integer format")
)
