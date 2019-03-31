package radio

import (
	"context"
	"io"
	"net"
	"time"
)

// ListenAndServe starts a RESP server on the given listener. Commands will
// be dispatched to appropriate handlers using the dispatcher.
func ListenAndServe(ctx context.Context, l net.Listener, handler Handler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
		}

		con, err := l.Accept()
		if err != nil {
			return err
		}

		if tc, ok := con.(*net.TCPConn); ok {
			tc.SetKeepAlive(true)
			tc.SetKeepAlivePeriod(10 * time.Minute)
		}
		go clientLoop(ctx, con, handler)
	}
}

func clientLoop(ctx context.Context, rwc io.ReadWriteCloser, handler Handler) {
	parser := NewParser(rwc, true)
	defer rwc.Close()

	rw := &respWriter{
		w: rwc,
	}

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

			rwc.Write([]byte(ErrorStr("ERR " + err.Error()).Serialize()))
			continue
		}

		handler.ServeRESP(rw, &Request{
			Command: "PING",
			Value:   val,
		})
	}
}

type respWriter struct {
	w io.Writer
}

func (rw *respWriter) Write(v Value) (int, error) {
	return rw.w.Write([]byte(v.Serialize()))
}
