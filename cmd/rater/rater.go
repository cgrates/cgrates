package main

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"os"
)

var (
	server = flag.String("server", "127.0.0.1:2000", "target host:port")
	listen = flag.String("listen", "127.0.0.1:1234", "target host:port")
	storage Storage
)

type Storage struct {
	sg StorageGetter
}

func NewStorage(nsg StorageGetter) *Storage{
	return &Storage{sg: nsg}
}

func (s *Storage) Get(args string, reply *string) (err error) {
	*reply, err = s.sg.Get(args)
	return err
}

func (s *Storage) Shutdown(args string, reply *string) (err error) {
	s.sg.Close()
	defer os.Exit(0)
	*reply = "Done!"
	return nil
}

func main() {	
	flag.Parse()
	getter, err := NewKyotoStorage("storage.kch")
	//getter, err := NewRedisStorage("tcp:127.0.0.1:6379")
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
