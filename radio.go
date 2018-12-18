package radio

import (
	"context"
	"net"
	"time"

	"github.com/k0kubun/pp"
	"github.com/spy16/radio/resp"
)

// New initializes the server
func New(lg Logger) *Server {
	return &Server{
		lg: lg,
	}
}

// Server represents a RESP compatible server.
type Server struct {
	lg Logger
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
	parser := resp.NewParser(con, true)
	defer con.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		val, err := parser.Next()
		if err != nil {
			srv.lg.Errorf("failed to parse: %v", err)
			return
		}

		pp.Println(val)
		con.Write([]byte(""))
	}
}

// Logger is responsible for levelled logging.
type Logger interface {
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}
