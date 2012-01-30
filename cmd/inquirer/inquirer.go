package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"errors"
	"time"
	"runtime"
	"sync"
)

const NCPU = 4

var (
	raterList *RaterList
	inChannels [NCPU]chan string
	outChannels [NCPU]chan string
	multiplexerIndex int 
	mu sync.Mutex
	sem = make(chan int, NCPU)
)



/*
Handler for the statistics web client
*/
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><ol>")
	for _, addr := range raterList.clientAddresses {
		fmt.Fprint(w, fmt.Sprintf("<li>Client: %v</li>", addr))
	}
	fmt.Fprint(w, fmt.Sprintf("<li>Gorutines: %v</li>", runtime.Goroutines()))
	fmt.Fprint(w, "</ol></body></html>")
}

/*
Creates a gorutine for every cpu core and the multiplexses the calls to each of them.
*/
func initThreadedCallRater(){
	multiplexerIndex = 0
	runtime.GOMAXPROCS(NCPU)
	fmt.Println(runtime.GOMAXPROCS(NCPU))
	for i:= 0; i< NCPU; i++ {
		inChannels[i] = make(chan string)
		outChannels[i] = make(chan string)
		go func(in, out chan string){
			for {
				key := <- in
				out <- CallRater(key)
			}
		}(inChannels[i], outChannels[i])
	}	
}

/*
*/
func ThreadedCallRater(key string) (replay string) {
	mu.Lock()
	defer mu.Unlock()
	if multiplexerIndex >= NCPU {
		multiplexerIndex = 0
	}
	inChannels[multiplexerIndex] <- key
	replay = <- outChannels[multiplexerIndex]
	multiplexerIndex++
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func CallRater(key string) (reply string) {
	err := errors.New("") //not nil value
	for err != nil {
		client:= raterList.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			err = client.Call("Storage.Get", key, &reply)
			if err != nil {
				log.Printf("Got en error from rater: %v", err)
			}
		}			
	}
	return 
}

func listenToTheWorld() {
	l, err := net.Listen("tcp", ":5090")
	defer l.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Print("listening:", l.Addr())

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

func main() {
	raterList = NewRaterList()
	raterServer := new(RaterServer)
	rpc.Register(raterServer)
	rpc.HandleHTTP()
	
	go StopSingnalHandler()	
	go listenToTheWorld()
	//initThreadedCallRater()
	http.HandleFunc("/", handler)	
	log.Print("The server is listening...")
	http.ListenAndServe(":2000", nil)
}
