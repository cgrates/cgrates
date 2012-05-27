/*
Rating system designed to be used in VoIP Carriers Wobld
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
	"sync"
	"log"
)

type Balancer struct {
	clientAddresses   []string // we need to hold these two slices because maps fo not keep order
	clientConnections []*rpc.Client
	balancerIndex     int
	mu                sync.RWMutex
}

/*
Constructor for RateList holding one slice for addreses and one slice for connections.
*/
func NewBalancer() *Balancer {
	r := &Balancer{balancerIndex: 0} // leaving both slices to nil
	return r
}

/*
Adds a client to the two  internal slices.
*/
func (bl *Balancer) AddClient(address string, client *rpc.Client) {
	bl.clientAddresses = append(bl.clientAddresses, address)
	bl.clientConnections = append(bl.clientConnections, client)
	return
}

/*
Removes a client from the slices locking the readers and reseting the balancer index.
*/
func (bl *Balancer) RemoveClient(address string) {
	index := -1
	for i, v := range bl.clientAddresses {
		if v == address {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	bl.clientAddresses = append(bl.clientAddresses[:index], bl.clientAddresses[index+1:]...)
	bl.clientConnections = append(bl.clientConnections[:index], bl.clientConnections[index+1:]...)
	bl.balancerIndex = 0
}

/*
Returns a client for the specifed address.
*/
func (bl *Balancer) GetClient(address string) (*rpc.Client, bool) {
	for i, v := range bl.clientAddresses {
		if v == address {
			return bl.clientConnections[i], true
		}
	}
	return nil, false
}

/*
Returns the next available connection at each call looping at the end of connections.
*/
func (bl *Balancer) Balance() (result *rpc.Client) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	if bl.balancerIndex >= len(bl.clientAddresses) {
		bl.balancerIndex = 0
	}
	if len(bl.clientAddresses) > 0 {
		result = bl.clientConnections[bl.balancerIndex]
		bl.balancerIndex++
	}

	return
}

func (bl *Balancer) Shutdown() {
	var reply string
	for i, client := range bl.clientConnections {
		client.Call("Responder.Shutdown", "", &reply)
		log.Printf("Shutdown rater %v: %v ", bl.clientAddresses[i], reply)
	}
}

func (bl *Balancer) GetClientAddresses() []string {
	return bl.clientAddresses
}
