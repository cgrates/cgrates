package main

import (
	"github.com/rif/cgrates/timespans"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Responder byte

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *Responder) GetGost(arg timespans.CallDescriptor, replay *timespans.CallCost) (err error) {
	*replay = *GetCost(&arg)
	return
}

func (r *Responder) DebitBalance(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = Debit(&arg, "Storage.DebitCents")
	return
}

func (r *Responder) DebitSMS(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = Debit(&arg, "Storage.DebitSMS")
	return
}

func (r *Responder) DebitSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = Debit(&arg, "Storage.DebitSeconds")
	return
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
