package radio

// Handler represents a RESP command handler.
type Handler interface {
	ServeRESP(wr ResponseWriter, req *Request)
}

// ResponseWriter represents a RESP writer object.
type ResponseWriter interface {
	Write(v Value) (int, error)
}

// Request represents a RESP request.
type Request struct {
	Command string
	Args    []string
}

// HandlerFunc implements Handler interface using a function type.
type HandlerFunc func(wr ResponseWriter, req *Request)

// ServeRESP dispatches the request to the wrapped function.
func (handlerFunc HandlerFunc) ServeRESP(wr ResponseWriter, req *Request) {
	handlerFunc(wr, req)
}
