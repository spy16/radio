package radio

import (
	"bufio"
	"context"
	"log"
	"net"
)

// Server represents a RESP compatible server.
type Server struct {
}

// Serve starts the server loop for accepting client connections.
func (srv *Server) Serve(ctx context.Context, l net.Listener) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			con, err := l.Accept()
			if err != nil {
				return err
			}
			go srv.handle(ctx, con)
		}
	}
}

func (srv *Server) handle(ctx context.Context, con net.Conn) {
	rdr := bufio.NewReader(con)

	b, err := rdr.ReadByte()
	if err != nil {
		log.Printf("error: %v", err)
	}

	switch b {
	case '*':

	}
}
