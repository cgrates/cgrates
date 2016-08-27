package balancer2go

import (
	"net/rpc"
	"testing"
)

func BenchmarkBalance(b *testing.B) {
	balancer := NewBalancer()
	balancer.AddClient("client 1", new(rpc.Client))
	balancer.AddClient("client 2", new(rpc.Client))
	balancer.AddClient("client 3", new(rpc.Client))
	for i := 0; i < b.N; i++ {
		balancer.Balance()
	}
}

func TestRemoving(t *testing.T) {
	balancer := NewBalancer()
	c1 := new(rpc.Client)
	c2 := new(rpc.Client)
	c3 := new(rpc.Client)
	balancer.AddClient("client 1", c1)
	balancer.AddClient("client 2", c2)
	balancer.AddClient("client 3", c3)
	balancer.RemoveClient("client 2")
	if balancer.clients["client 1"] != c1 ||
		balancer.clients["client 3"] != c3 ||
		len(balancer.clients) != 2 {
		t.Error("Failed removing rater")
	}
}

func TestGet(t *testing.T) {
	balancer := NewBalancer()
	c1 := new(rpc.Client)
	balancer.AddClient("client 1", c1)
	result, ok := balancer.GetClient("client 1")
	if !ok || c1 != result {
		t.Error("Get failed")
	}
}

func TestOneBalancer(t *testing.T) {
	balancer := NewBalancer()
	balancer.AddClient("client 1", new(rpc.Client))
	c1 := balancer.Balance()
	c2 := balancer.Balance()
	if c1 != c2 {
		t.Error("With only one rater these shoud be equal")
	}
}
