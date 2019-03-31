package radio

import (
	"fmt"
)

// ProtocolError represents violations of RESP protocol.
type ProtocolError struct {
	Reason string
}

func (e ProtocolError) Error() string {
	return fmt.Sprintf("Protocol Error: %s", e.Reason)
}

// DispatcherError represents error during dispatcher operations.
type DispatcherError struct {
	Reason string
}

func (e DispatcherError) Error() string {
	return e.Reason
}
