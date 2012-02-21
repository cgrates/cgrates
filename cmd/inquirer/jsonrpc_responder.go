package main

import (
	"github.com/rif/cgrates/timespans"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"log"
)

type Responder byte

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *Responder) Get(arg timespans.CallDescriptor, replay *timespans.CallCost) error {
	*replay = *CallRater(&arg)
	return nil
}

/*
Creates the json rpc server.
*/
func listenToJsonRPCRequests() {
	l, err := net.Listen("tcp", *jsonRpcAddress)
	defer l.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening for incomming json RPC requests on ", l.Addr())

	responder := new(Responder)
	rpc.Register(responder)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %s", c)
			continue
		}

		log.Printf("connection started: %v", c.RemoteAddr())
		go jsonrpc.ServeConn(c)
	}
}
