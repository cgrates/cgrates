package main

import (
	"flag"
	"github.com/rif/cgrates/timespans"
	"log"
	"net"
	"net/rpc"
	"os"
)

var (
	server  = flag.String("server", "127.0.0.1:2000", "target host:port")
	listen  = flag.String("listen", "127.0.0.1:1234", "target host:port")
	storage Storage
)

type Storage struct {
	sg timespans.StorageGetter
}

func NewStorage(nsg timespans.StorageGetter) *Storage {
	return &Storage{sg: nsg}
}

/*
RPC method providing the rating information from the storage.
*/
func (s *Storage) GetCost(cd timespans.CallDescriptor, reply *timespans.CallCost) (err error) {
	descriptor := &cd
	descriptor.StorageGetter = s.sg
	r, e := descriptor.GetCost()
	*reply, err = *r, e
	return nil
}

/*
RPC method that trigers rater shutdown in case of server exit.
*/
func (s *Storage) Shutdown(args string, reply *string) (err error) {
	s.sg.Close()
	defer os.Exit(0)
	*reply = "Done!"
	return nil
}

func main() {
	flag.Parse()
	getter, err := timespans.NewKyotoStorage("storage.kch")
	//getter, err := NewRedisStorage("tcp:127.0.0.1:6379", 10)
	//defer getter.Close()
	if err != nil {
		log.Printf("Cannot open storage file: %v", err)
		os.Exit(1)
	}
	storage := NewStorage(getter)
	rpc.Register(storage)
	rpc.HandleHTTP()
	go RegisterToServer(server, listen)
	go StopSingnalHandler(server, listen, getter)
	addr, err1 := net.ResolveTCPAddr("tcp", *listen)
	l, err2 := net.ListenTCP("tcp", addr)
	if err1 != nil || err2 != nil {
		log.Print("cannot create listener for specified address ", *listen)
		os.Exit(1)
	}
	rpc.Accept(l)
}
