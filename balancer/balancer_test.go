/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package balancer

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
	if balancer.clientConnections[0] != c1 ||
		balancer.clientConnections[1] != c3 ||
		len(balancer.clientConnections) != 2 {
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

func Test100Balancer(t *testing.T) {
	balancer := NewBalancer()
	var clients []*rpc.Client
	for i := 0; i < 100; i++ {
		c := new(rpc.Client)
		clients = append(clients, c)
		balancer.AddClient("client 1", c)
	}
	for i := 0; i < 100; i++ {
		c := balancer.Balance()
		if c != clients[i] {
			t.Error("Balance did not iterate all the available clients")
		}
	}
	c := balancer.Balance()
	if c != clients[0] {
		t.Error("Balance did not lopped from the begining")
	}

}
