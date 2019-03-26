package radio

import (
	"errors"

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
func (d *Dispatcher) Dispatch(arr []resp.BulkStr) (resp.Value, error) {
	return nil, nil
}

// Handler represents a command handler.
type Handler interface {
	Handle()
}
