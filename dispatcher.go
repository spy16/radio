package radio

import (
	"errors"
	"fmt"

	"github.com/spy16/radio/resp"
)

// ErrUnknownCommand is returned when a handler for a command is not
// found.
var ErrUnknownCommand = errors.New("unknown command")

// Dispatcher routes BulkStr commands to appropriate handlers.
// Refer https://redis.io/topics/protocol#inline-commands for understanding
// enableInline flag.
type Dispatcher struct {
	enableInline bool
	handlers     map[string]Handler
}

// Dispatch matches the command name in the arr and calls the appropriate handler
// registered for the command.
func (d *Dispatcher) Dispatch(val resp.Value) (resp.Value, error) {
	return nil, nil
}

// Handler represents a command handler.
type Handler interface {
	Handle()
}

func ensureBulkStrArray(arr resp.Array) error {
	for _, itm := range arr.Items {
		_, isBulkStr := itm.(*resp.BulkStr)

		if !isBulkStr {
			return fmt.Errorf("Protocol error: expected '$', got '%c'", itm.Serialize()[0])
		}
	}

	return nil
}
