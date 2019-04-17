package radio

import (
	"context"
	"io"
	"net"
	"time"
)

// ListenAndServe starts a RESP server on the given listener. Parsed requests will
// be passed to the given handler.
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
	rdr := NewReader(rwc, true)
	rw := NewWriter(rwc)
	defer rwc.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		mb, err := rdr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			rw.Write(ErrorStr("ERR " + err.Error()))
			continue
		}

		req := newRequest(mb)
		if req != nil {
			handler.ServeRESP(rw, req)
		}
	}
}

func newRequest(v Value) *Request {
	mb, ok := v.(*MultiBulk)
	if !ok {
		return nil
	}

	if mb.IsNil() || len(mb.Items) == 0 {
		return nil
	}

	return &Request{
		Command: mb.Items[0].String(),
		Args:    mb.Items[1:],
		Value:   mb,
	}
}
