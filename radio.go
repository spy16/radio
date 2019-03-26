package radio

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spy16/radio/resp"
)

// New initializes the server
func New() *Server {
	return &Server{}
}

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

			if tc, ok := con.(*net.TCPConn); ok {
				tc.SetKeepAlive(true)
				tc.SetKeepAlivePeriod(10 * time.Minute)
			}

			go srv.clientLoop(ctx, con)
		}
	}
}

func (srv *Server) clientLoop(ctx context.Context, con net.Conn) {
	parser := resp.New(con, true)
	defer con.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		val, err := parser.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			con.Write([]byte(resp.ErrorStr("ERR " + err.Error()).Serialize()))
			continue
		}

		spew.Dump(val)
		con.Write([]byte(resp.SimpleStr("PONG").Serialize()))
	}
}
