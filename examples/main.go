package main

import (
	"context"
	"net"

	"github.com/spy16/radio"
)

func main() {
	srv := radio.New()

	l, err := net.Listen("tcp", ":9736")
	if err != nil {
		panic(err)
	}

	srv.Serve(context.Background(), l)
}
