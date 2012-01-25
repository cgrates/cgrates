package registration

import (
	"log"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

func StopSingnalHandler(server, listen *string) {
	log.Print("Handling stop signals...")
	sig := <-signal.Incoming
	if usig, ok := sig.(os.UnixSignal); ok {
		switch usig {
		case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
			log.Printf("Caught signal %v, unregistering from server\n", usig)
			unregisterFromServer(server, listen)
			os.Exit(1)
		}
	}
}

func unregisterFromServer(server, listen *string) {
	client, err := rpc.DialHTTP("tcp", *server)
	if err != nil {
		log.Panic("Cannot register to server!")
	}
	var reply byte
	log.Print("Unregistering from server ", *server)
	client.Call("RaterList.UnRegisterRater", *listen, &reply)
	if err := client.Close(); err != nil {
		log.Panic("Could not close server unregistration!")
	}
}

func RegisterToServer(server, listen *string) {
	client, err := rpc.DialHTTP("tcp", *server)
	if err != nil {
		log.Panic("Cannot register to server!")
	}
	var reply byte
	log.Print("Registering to server ", *server)
	client.Call("RaterList.RegisterRater", *listen, &reply)
	if err := client.Close(); err != nil {
		log.Panic("Could not close server registration!")
	}
}