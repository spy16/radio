package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/spy16/radio"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":9736", "TCP address to listen for connections")
	flag.Parse()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err.Error())
	}

	respHandler := radio.HandlerFunc(serveRESP)

	log.Printf("listening for clients on '%s'...", addr)
	log.Fatalf("server exited: %v", radio.ListenAndServe(context.Background(), l, respHandler))
}

func serveRESP(wr radio.ResponseWriter, req *radio.Request) {
	switch req.Command {
	case "ping":
		wr.Write(radio.SimpleStr("PONG"))

	case "COMMAND":
		wr.Write(&radio.Array{})

	default:
		wr.Write(radio.ErrorStr(fmt.Sprintf("unknown command '%s'", req.Command)))
	}
}
