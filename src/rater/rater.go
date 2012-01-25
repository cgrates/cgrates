package main

import (
	"flag"
	"log"
	"math"
	"net"
	"net/rpc"
)

var (
	server = flag.String("server", "127.0.0.1:2000", "target host:port")
	listen = flag.String("listen", "127.0.0.1:1234", "target host:port")
)

type Sumer int

func (t *Sumer) Square(n float64, reply *float64) error {
	*reply = math.Sqrt(n)
	return nil
}

func main() {
	flag.Parse()
	arith := new(Sumer)
	rpc.Register(arith)
	rpc.HandleHTTP()
	go RegisterToServer(server, listen)
	go registration.StopSingnalHandler(server, listen)
	addr, err1 := net.ResolveTCPAddr("tcp", *listen)
	l, err2 := net.ListenTCP("tcp", addr)
	if err1 != nil || err2 != nil {
		log.Panic("cannot create listener for specified address ", *listen)
	}
	rpc.Accept(l)
}
