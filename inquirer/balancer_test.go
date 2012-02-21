package main

import (
	"net/rpc"
	"testing"
)

func BenchmarkBalance(b *testing.B) {
	raterlist := NewRaterList()
	raterlist.AddClient("client 1", new(rpc.Client))
	raterlist.AddClient("client 2", new(rpc.Client))
	raterlist.AddClient("client 3", new(rpc.Client))
	for i := 0; i < b.N; i++ {
		raterlist.Balance()
	}
}

func TestRemoving(t *testing.T) {
	raterlist := NewRaterList()
	c1 := new(rpc.Client)
	c2 := new(rpc.Client)
	c3 := new(rpc.Client)
	raterlist.AddClient("client 1", c1)
	raterlist.AddClient("client 2", c2)
	raterlist.AddClient("client 3", c3)
	raterlist.RemoveClient("client 2")
	if raterlist.clientConnections[0] != c1 ||
		raterlist.clientConnections[1] != c3 ||
		len(raterlist.clientConnections) != 2 {
		t.Error("Failed removing rater")
	}
}

func TestGet(t *testing.T) {
	raterlist := NewRaterList()
	c1 := new(rpc.Client)
	raterlist.AddClient("client 1", c1)
	result, ok := raterlist.GetClient("client 1")
	if !ok || c1 != result {
		t.Error("Get failed")
	}
}

func TestOneBalancer(t *testing.T) {
	raterlist := NewRaterList()
	raterlist.AddClient("client 1", new(rpc.Client))
	c1 := raterlist.Balance()
	c2 := raterlist.Balance()
	if c1 != c2 {
		t.Error("With only one rater these shoud be equal")
	}
}

func Test100Balancer(t *testing.T) {
	raterlist := NewRaterList()
	var clients []*rpc.Client
	for i := 0; i < 100; i++ {
		c := new(rpc.Client)
		clients = append(clients, c)
		raterlist.AddClient("client 1", c)
	}
	for i := 0; i < 100; i++ {
		c := raterlist.Balance()
		if c != clients[i] {
			t.Error("Balance did not iterate all the available clients")
		}
	}
	c := raterlist.Balance()
	if c != clients[0] {
		t.Error("Balance did not lopped from the begining")
	}

}
