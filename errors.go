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
