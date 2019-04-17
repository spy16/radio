package radio

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"
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

		val, err := rdr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			rw.Write(ErrorStr("ERR " + err.Error()))
			return
		}

		req, err := newRequest(val)
		if err != nil {
			rw.Write(ErrorStr("ERR " + err.Error()))
			return
		}

		if req == nil {
			continue
		}

		handler.ServeRESP(rw, req)
	}
}

func newRequest(val Value) (*Request, error) {
	arr, ok := val.(*Array)
	if !ok {
		return nil, nil
	}

	if arr.IsNil() || len(arr.Items) == 0 {
		return nil, nil
	}

	req := &Request{}
	for i, itm := range arr.Items {
		v, ok := itm.(*BulkStr)
		if !ok {
			return nil, fmt.Errorf("unexpected type '%s'", reflect.TypeOf(itm))
		}

		if i == 0 {
			req.Command = v.String()
		} else {
			req.Args = append(req.Args, v.String())
		}
	}

	return req, nil
}
